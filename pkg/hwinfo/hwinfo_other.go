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
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/jaypipes/ghw"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vcraescu/go-xrandr"
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
		pciInfo, ghwErr := ghw.PCI()
		if ghwErr == nil {
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
		log.Info("[HWINFO] lsusb command is not available. Skipping USB listing...")
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
		log.WithError(err).Info("[HWINFO] xrandr not installed or returned not expected result. Skipping display listing...")
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

// this part is extended version of github.com/shirou/gopsutil/cpu/cpu_linux.go
// it need own implementation as gopsutil does return amount of threads not CPUs
// and ignores siblings field which is indicating amount of threads
// /proc/cpuinfo is parsed in following manner
//   - physical id: actual CPU installed in the socket
//   - cores:       physical cores per CPU in the socket
//   - siblings:    amount of threads per CPU in the socket
// e.g. on HT CPU with 2 cores amount of siblings will be 4
func listCPUs() ([]cpuInfo, error) {
	lines, _ := readLines(getEnv("HOST_PROC", "/proc", "cpuinfo"))

	var cpus []cpuStat
	var processorName string

	c := cpuStat{CPU: -1, Cores: 1}
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
			if c.CPU >= 0 {
				err := finishCPUInfo(&c)
				if err != nil {
					return nil, err
				}
				cpus = append(cpus, c)
			}
			c = cpuStat{Cores: 1, ModelName: processorName}
			t, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return nil, err
			}
			c.CPU = int32(t)
		case "vendorId", "vendor_id":
			c.VendorID = value
		case "cpu family":
			c.Family = value
		case "model":
			c.Model = value
		case "model name", "cpu":
			c.ModelName = value
			if strings.Contains(value, "POWER8") ||
				strings.Contains(value, "POWER7") {
				c.Model = strings.Split(value, " ")[0]
				c.Family = "POWER"
				c.VendorID = "IBM"
			}
		case "stepping", "revision":
			val := value

			if key == "revision" {
				val = strings.Split(value, ".")[0]
			}

			t, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, err
			}
			c.Stepping = int32(t)
		case "cpu MHz", "clock":
			// treat this as the fallback value, thus we ignore error
			if t, err := strconv.ParseFloat(strings.Replace(value, "MHz", "", 1), 64); err == nil {
				c.Mhz = t
			}
		case "cache size":
			t, err := strconv.ParseInt(strings.Replace(value, " KB", "", 1), 10, 64)
			if err != nil {
				return nil, err
			}
			c.CacheSize = int32(t)
		case "physical id":
			c.PhysicalID = value
		case "core id":
			c.CoreID = value
		case "flags", "Features":
			c.Flags = strings.FieldsFunc(value, func(r rune) bool {
				return r == ',' || r == ' '
			})
		case "cpu cores":
			val := value

			t, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, err
			}

			c.Cores = int32(t)
		case "microcode":
			c.Microcode = value
		case "siblings":
			val := value

			t, err := strconv.ParseInt(val, 10, 64)
			if err != nil {
				return nil, err
			}

			c.Siblings = int32(t)
		}
	}

	if c.CPU >= 0 {
		err := finishCPUInfo(&c)
		if err != nil {
			return nil, err
		}
		cpus = append(cpus, c)
	}

	sorted := make(map[string]*cpuInfo)

	for _, cpu := range cpus {
		if _, found := sorted[cpu.PhysicalID]; !found {
			info := &cpuInfo{
				manufacturer:      cpu.VendorID,
				manufacturingInfo: fmt.Sprintf("Model %s Family %s Stepping %d", cpu.Model, cpu.Family, cpu.Stepping),
				description:       cpu.ModelName,
				coreCount:         fmt.Sprintf("%d", cpu.Cores),
				coreEnabled:       fmt.Sprintf("%d", cpu.Cores),
				threadCount:       fmt.Sprintf("%d", cpu.Siblings),
			}

			sorted[cpu.PhysicalID] = info
		}
	}

	var cInfo []cpuInfo
	for _, cpu := range sorted {
		cInfo = append(cInfo, *cpu)
	}

	return cInfo, nil
}

func finishCPUInfo(c *cpuStat) error {
	var lines []string
	var err error
	var value float64

	if len(c.CoreID) == 0 {
		lines, err = readLines(sysCPUPath(c.CPU, "topology/core_id"))
		if err == nil {
			c.CoreID = lines[0]
		}
	}

	// override the value of c.Mhz with cpufreq/cpuinfo_max_freq regardless
	// of the value from /proc/cpuinfo because we want to report the maximum
	// clock-speed of the CPU for c.Mhz, matching the behaviour of Windows
	lines, err = readLines(sysCPUPath(c.CPU, "cpufreq/cpuinfo_max_freq"))
	// if we encounter errors below such as there are no cpuinfo_max_freq file,
	// we just ignore. so let Mhz is 0.
	if err != nil {
		return nil
	}
	value, err = strconv.ParseFloat(lines[0], 64)
	if err != nil {
		return nil
	}
	c.Mhz = value / 1000.0 // value is in kHz
	if c.Mhz > 9999 {
		c.Mhz = c.Mhz / 1000.0 // value in Hz
	}
	return nil
}

// ReadLines reads contents from a file and splits them by new lines.
// A convenience wrapper to ReadLinesOffsetN(filename, 0, -1).
func readLines(filename string) ([]string, error) {
	return readLinesOffsetN(filename, 0, -1)
}

// ReadLines reads contents from file and splits them by new line.
// The offset tells at which line number to start.
// The count determines the number of lines to read (starting from offset):
//   n >= 0: at most n lines
//   n < 0: whole file
func readLinesOffsetN(filename string, offset uint, n int) ([]string, error) {
	f, err := os.Open(filename)
	if err != nil {
		return []string{""}, err
	}
	defer f.Close()

	var ret []string

	r := bufio.NewReader(f)
	for i := 0; i < n+int(offset) || n < 0; i++ {
		line, err := r.ReadString('\n')
		if err != nil {
			break
		}
		if i < int(offset) {
			continue
		}
		ret = append(ret, strings.Trim(line, "\n"))
	}

	return ret, nil
}

func sysCPUPath(cpu int32, relPath string) string {
	return getEnv("HOST_SYS", "/sys", fmt.Sprintf("devices/system/cpu/cpu%d", cpu), relPath)
}

func getEnv(key string, dfault string, combineWith ...string) string {
	value := os.Getenv(key)
	if value == "" {
		value = dfault
	}

	switch len(combineWith) {
	case 0:
		return value
	case 1:
		return filepath.Join(value, combineWith[0])
	default:
		all := make([]string, len(combineWith)+1)
		all[0] = value
		copy(all[1:], combineWith)
		return filepath.Join(all...)
	}
}
