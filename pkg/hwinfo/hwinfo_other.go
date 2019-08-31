// +build !darwin,!windows

package hwinfo

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vcraescu/go-xrandr"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

var lsusbLineRegexp = regexp.MustCompile(`[0-9|a-z|A-Z|.|/|-|:|\[|\]|_|+| ]+`)

type cpuStat struct {
	CPU        int32    `json:"cpu"`
	VendorID   string   `json:"vendorId"`
	Family     string   `json:"family"`
	Model      string   `json:"model"`
	Stepping   int32    `json:"stepping"`
	PhysicalID string   `json:"physicalId"`
	CoreID     string   `json:"coreId"`
	Cores      int32    `json:"cores"`
	ModelName  string   `json:"modelName"`
	Mhz        float64  `json:"mhz"`
	CacheSize  int32    `json:"cacheSize"`
	Flags      []string `json:"flags"`
	Microcode  string   `json:"microcode"`
	Siblings   int32    `json:"siblings"`
}

func dmidecodeCommand() string {
	// expecting 'sudo' package is installed and /etc/sudoers.d/cagent-dmidecode is present
	return "sudo dmidecode"
}

func captureStderr(funcToExecute func()) (string, error) {
	r, w, err := os.Pipe()
	if err != nil {
		return "", err
	}

	// replace default stderr Writer with our own to capture output:
	defaultStderrWriter := os.Stderr
	os.Stderr = w

	// in any case, the default writer will be set again:
	defer func() {
		os.Stderr = defaultStderrWriter
	}()

	funcToExecute()
	err = w.Close()
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, r)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

func listPCIDevices() ([]*pciDeviceInfo, error) {
	var ghwErr error
	var devices []*ghw.PCIDevice

	// unfortunately, jaypipes/ghw library sometime writes error message directly to os.Stderr instead of returning from function call.
	// that's why we will try to capture stderr output and handle it:
	stderrOutput, err := captureStderr(func() {
		var pciInfo *ghw.PCIInfo
		if pciInfo, ghwErr = ghw.PCI(); ghwErr == nil {
			devices = pciInfo.ListDevices()
		}
	})
	if err != nil {
		return nil, errors.Wrap(err, "could not capture stderr when retrieving PCI information using ghw")
	}
	if ghwErr != nil {
		return nil, errors.Wrap(ghwErr, "there were error while retrieving PCI information using ghw")
	}
	if len(stderrOutput) > 0 {
		// ghw only reports to stderr in case if system files are missing or there was failure while reading it
		log.Warnf("[HWINFO] got error output while retrieving PCI information using ghw: %s\nProbably the system is not supported or system files unreadable.", stderrOutput)
	}

	result := make([]*pciDeviceInfo, 0, len(devices))
	for _, device := range devices {
		vendor := device.Vendor
		product := device.Product

		deviceType := device.Subclass.Name
		if deviceType == "unknown" {
			deviceType = ""
		}
		deviceClassName := device.Class.Name
		if deviceClassName == "unknown" {
			deviceClassName = ""
		}

		if deviceType == "" {
			deviceType = deviceClassName
		} else if deviceClassName != "" && deviceClassName != deviceType {
			deviceType = fmt.Sprintf("%s (%s)", deviceClassName, deviceType)
		}

		description := device.ProgrammingInterface.Name
		if description == "unknown" || description == "Normal decode" {
			description = ""
		}
		result = append(result, &pciDeviceInfo{
			DeviceType:  deviceType,
			Address:     device.Address,
			VendorName:  vendor.Name,
			ProductName: product.Name,
			Description: description,
		})
	}
	return result, nil
}

func listUSBDevices() ([]*usbDeviceInfo, error) {
	results := make([]*usbDeviceInfo, 0)
	reg := regexp.MustCompile(`[^:]+`)
	var lines []string

	cmd := exec.Command("lsusb")
	buf := bytes.Buffer{}
	cmd.Stdout = bufio.NewWriter(&buf)
	if err := cmd.Run(); err != nil {
		common.LogOncef(log.InfoLevel, "[HWINFO] lsusb command is not available: %s. Skipping USB listing...", err.Error())
		return nil, nil
	}

	outBytes, err := ioutil.ReadAll(bufio.NewReader(&buf))
	if err != nil {
		return nil, errors.Wrap(err, "could not read lsusb output")
	}

	lines = strings.Split(string(outBytes), "\n")

	// tokenize and parse command output line by line:
	const minExpectedTokensCount = 6
	for _, line := range lines {
		tokens := strings.Split(line, " ")
		sanitizedTokens := make([]string, 0)
		for _, t := range tokens {
			if t != "" && t != "\t" {
				sanitizedTokens = append(sanitizedTokens, t)
			}
		}
		sanitizedTokensCount := len(sanitizedTokens)
		if sanitizedTokensCount < minExpectedTokensCount {
			if sanitizedTokensCount > 0 {
				log.Warnf("[HWINFO] unexpected lsusb command output: got %d tokens in line: %s", sanitizedTokensCount, line)
			}
			continue
		}

		var description string
		if sanitizedTokensCount > minExpectedTokensCount {
			restTokens := sanitizedTokens[minExpectedTokensCount:]
			description = strings.Join(restTokens, " ")
		}

		address := fmt.Sprintf("bus %s device %s", sanitizedTokens[1], reg.FindString(sanitizedTokens[3]))
		devID := lsusbLineRegexp.FindString(sanitizedTokens[5])
		results = append(results, &usbDeviceInfo{
			Address:     address,
			VendorName:  "",
			DeviceID:    devID,
			Description: description,
		})
	}
	return results, nil
}

func listDisplays() ([]*monitorInfo, error) {
	results := make([]*monitorInfo, 0)
	screens, err := xrandr.GetScreens()
	if err != nil {
		common.LogOncef(log.InfoLevel, "[HWINFO] xrandr not installed or returned not expected result: %s. Skipping display listing...", err.Error())
		return nil, nil
	}
	for _, s := range screens {
		for _, m := range s.Monitors {
			physicalSizeStr := fmt.Sprintf("%dmm x %dmm", int(m.Size.Width), int(m.Size.Height))
			resolutionStr := fmt.Sprintf("%dx%d", int(m.Resolution.Width), int(m.Resolution.Height))
			results = append(results, &monitorInfo{
				ID:          m.ID,
				Size:        physicalSizeStr,
				Resolution:  resolutionStr,
				Description: "",
				VendorName:  "",
			})
		}
	}

	return results, nil
}

// this part is an extended version of github.com/shirou/gopsutil/cpu/cpu_linux.go
// own implementation is required as gopsutil does return amount of threads instead of CPUs
// and ignores siblings field which is indicating amount of threads
// /proc/cpuinfo is parsed in following manner
//   - physical id: actual CPU installed in the socket
//   - cores:       physical cores per CPU in the socket
//   - siblings:    amount of threads per CPU in the socket
// e.g. on HT CPU with 2 cores amount of siblings will be 4
func listCPUs() (map[string]interface{}, error) {
	lines, _ := common.ReadLines(common.GetEnv("HOST_PROC", "/proc", "cpuinfo"))

	var cpus []cpuStat
	var processorName string

	c := cpuStat{CPU: -1, Cores: 1}

	tryFinalizeCurrentCPU := func() {
		if c.CPU >= 0 {
			c.finalize()
			cpus = append(cpus, c)
		}
	}

	for _, line := range lines {
		fields := strings.Split(line, ":")
		if len(fields) < 2 {
			continue
		}
		key := strings.TrimSpace(fields[0])
		value := strings.TrimSpace(fields[1])

		switch key {
		case "Processor":
			processorName = value
		case "processor":
			tryFinalizeCurrentCPU()
			c = cpuStat{Cores: 1, ModelName: processorName}
			if err := c.parseFieldProcessor(value); err != nil {
				return nil, err
			}
		case "vendorId", "vendor_id":
			c.VendorID = value
		case "cpu family":
			c.Family = value
		case "model":
			c.Model = value
		case "model name", "cpu":
			c.parseFieldModelName(value)
		case "stepping", "revision":
			if err := c.parseFieldStepping(key, value); err != nil {
				return nil, err
			}
		case "cpu MHz", "clock":
			c.parseFieldClock(value)
		case "cache size":
			if err := c.parseFieldCacheSize(value); err != nil {
				return nil, err
			}
		case "physical id":
			c.PhysicalID = value
		case "core id":
			c.CoreID = value
		case "flags", "Features":
			c.Flags = strings.FieldsFunc(value, func(r rune) bool {
				return r == ',' || r == ' '
			})
		case "cpu cores":
			if err := c.parseFieldCPUCores(value); err != nil {
				return nil, err
			}
		case "microcode":
			c.Microcode = value
		case "siblings":
			if err := c.parseFieldSiblings(value); err != nil {
				return nil, err
			}
		}
	}

	tryFinalizeCurrentCPU()

	return encodeCPUs(cpus), nil
}

func (c *cpuStat) finalize() {
	var lines []string
	var err error
	var value float64

	if len(c.CoreID) == 0 {
		lines, err = common.ReadLines(sysCPUPath(c.CPU, "topology/core_id"))
		if err == nil {
			c.CoreID = lines[0]
		}
	}

	// override the value of c.Mhz with cpufreq/cpuinfo_max_freq regardless
	// of the value from /proc/cpuinfo because we want to report the maximum
	// clock-speed of the CPU for c.Mhz, matching the behavior of Windows
	lines, err = common.ReadLines(sysCPUPath(c.CPU, "cpufreq/cpuinfo_max_freq"))
	// if we encounter errors below such as there are no cpuinfo_max_freq file,
	// we just ignore. so let Mhz is 0.
	if err != nil {
		return
	}

	if value, err = strconv.ParseFloat(lines[0], 64); err != nil {
		return
	}

	c.Mhz = value / 1000.0 // value is in kHz
	if c.Mhz > 9999 {
		c.Mhz = c.Mhz / 1000.0 // value in Hz
	}
}

func (c *cpuStat) parseFieldProcessor(value string) error {
	t, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}
	c.CPU = int32(t)

	return nil
}

func (c *cpuStat) parseFieldModelName(value string) {
	c.ModelName = value
	if strings.Contains(value, "POWER8") ||
		strings.Contains(value, "POWER7") {
		c.Model = strings.Split(value, " ")[0]
		c.Family = "POWER"
		c.VendorID = "IBM"
	}
}

// treat this as the fallback value, thus we ignore error
func (c *cpuStat) parseFieldClock(value string) {
	if t, err := strconv.ParseFloat(strings.Replace(value, "MHz", "", 1), 64); err == nil {
		c.Mhz = t
	}
}

func (c *cpuStat) parseFieldStepping(key, value string) error {
	val := value

	if key == "revision" {
		val = strings.Split(value, ".")[0]
	}

	t, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return err
	}

	c.Stepping = int32(t)

	return nil
}

func (c *cpuStat) parseFieldSiblings(value string) error {
	t, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		return err
	}

	c.Siblings = int32(t)

	return nil
}

func (c *cpuStat) parseFieldCPUCores(value string) error {
	t, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return err
	}

	c.Cores = int32(t)

	return nil
}

func (c *cpuStat) parseFieldCacheSize(value string) error {
	t, err := strconv.ParseInt(strings.Replace(value, " KB", "", 1), 10, 64)
	if err != nil {
		return err
	}
	c.CacheSize = int32(t)

	return nil
}

func encodeCPUs(cpus []cpuStat) map[string]interface{} {
	sorted := make(map[string]bool)
	encodedCpus := make(map[string]interface{})

	for _, cpu := range cpus {
		if _, found := sorted[cpu.PhysicalID]; !found {
			sorted[cpu.PhysicalID] = true

			encodedCpus[fmt.Sprintf("cpu.%s.manufacturer", cpu.PhysicalID)] = cpu.VendorID
			encodedCpus[fmt.Sprintf("cpu.%s.manufacturing_info", cpu.PhysicalID)] = fmt.Sprintf("Model %s Family %s Stepping %d", cpu.Model, cpu.Family, cpu.Stepping)
			encodedCpus[fmt.Sprintf("cpu.%s.description", cpu.PhysicalID)] = cpu.ModelName
			encodedCpus[fmt.Sprintf("cpu.%s.core_count", cpu.PhysicalID)] = fmt.Sprintf("%d", cpu.Cores)
			encodedCpus[fmt.Sprintf("cpu.%s.thread_count", cpu.PhysicalID)] = fmt.Sprintf("%d", cpu.Siblings)
		}
	}

	return encodedCpus
}

func sysCPUPath(cpu int32, relPath string) string {
	return common.GetEnv("HOST_SYS", "/sys", fmt.Sprintf("devices/system/cpu/cpu%d", cpu), relPath)
}
