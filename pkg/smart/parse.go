package smart

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"syscall"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

type smartctlStatus struct {
	Status   int `json:"exit_status"`
	Messages []struct {
		String string `json:"string"`
	}
}
type smartErrorResult struct {
	Smartctl smartctlStatus `json:"smartctl"`
}

type ataSMARTAttributes struct {
	Table []struct {
		ID         int    `json:"id"`
		Name       string `json:"name"`
		Value      int    `json:"value"`
		Worst      int    `json:"worst"`
		Thresh     int    `json:"thresh"`
		WhenFailed string `json:"when_failed"`
		Raw        struct {
			Value  int    `json:"value"`
			String string `json:"string"`
		} `json:"raw"`
	} `json:"table"`
}

type nvmeSmartHealthInformationLog struct {
	CriticalWarning         int   `json:"critical_warning"`
	Temperature             int   `json:"temperature"`
	AvailableSpare          int   `json:"available_spare"`
	AvailableSpareThreshold int   `json:"available_spare_threshold"`
	PercentageUsed          int   `json:"percentage_used"`
	DataUnitsRead           int64 `json:"data_units_read"`
	DataUnitsWritten        int64 `json:"data_units_written"`
	HostReads               int64 `json:"host_reads"`
	HostWrites              int64 `json:"host_writes"`
	ControllerBusyTime      int64 `json:"controller_busy_time"`
	PowerCycles             int   `json:"power_cycles"`
	PowerOnHours            int   `json:"power_on_hours"`
	UnsafeShutdowns         int   `json:"unsafe_shutdowns"`
	MediaErrors             int   `json:"media_errors"`
	NumErrLogEntries        int   `json:"num_err_log_entries"`
}

type parseResult struct {
	Smartctl smartctlStatus `json:"smartctl"`

	Device struct {
		Name     string `json:"name"`
		InfoName string `json:"info_name"`
		Type     string `json:"type"`
		Protocol string `json:"protocol"`
	} `json:"device"`

	ModelName          string  `json:"model_name"`
	SerialNumber       string  `json:"serial_number"`
	ModelFamily        *string `json:"model_family"`
	FirmwareVersion    string  `json:"firmware_version"`
	InSmartctlDatabase bool    `json:"in_smartctl_database"`

	SmartStatus *struct {
		Passed bool `json:"passed"`
	} `json:"smart_status"`

	Temperature *struct {
		Current float32 `json:"current"`
	} `json:"temperature"`

	PowerCycleCount *int `json:"power_cycle_count"`
	PowerOnTime     *struct {
		Hours int `json:"hours"`
	} `json:"power_on_time"`

	// RotationRate is of type interface as on Linux can return string
	// for SSD while on Windows it will be int with value 0
	RotationRate      interface{} `json:"rotation_rate,omitempty"`
	LogicalBlockSize  *int        `json:"logical_block_size,omitempty"`
	PhysicalBlockSize *int        `json:"physical_block_size,omitempty"`

	UserCapacity *struct {
		Blocks int   `json:"blocks"`
		Bytes  int64 `json:"bytes"`
	} `json:"user_capacity"`

	ATAVersion *struct {
		String     string `json:"string"`
		MajorValue int    `json:"major_value"`
		MinorValue int    `json:"minor_value"`
	} `json:"ata_version"`

	SATAVersion *struct {
		String string `json:"string"`
		Value  int    `json:"value"`
	} `json:"sata_version"`

	InterfaceSpeed *struct {
		Max struct {
			SataValue      int    `json:"sata_value"`
			String         string `json:"string"`
			UnitsPerSecond int64  `json:"units_per_second"`
			BitsPerUnit    int64  `json:"bits_per_unit"`
		} `json:"max"`
	} `json:"interface_speed"`

	NVMESmartHealthInformationLog *nvmeSmartHealthInformationLog `json:"nvme_smart_health_information_log"`

	ATASmartData *struct {
		SelfTest struct {
			String string `json:"string"`
			Passed bool   `json:"passed"`
		} `json:"self_test"`
		Capabilities struct {
			ExecOfflineImmediateSupport bool `json:"exec_offline_immediate_supported"`
			OfflineIsAbortedUponNewCmd  bool `json:"offline_is_aborted_upon_new_cmd"`
			OfflineSurfaceScanSupported bool `json:"offline_surface_scan_supported"`
			SelfTestsSupported          bool `json:"self_tests_supported"`
			ConveyanceSelfTestSupported bool `json:"conveyance_self_test_supported"`
			SelectiveSelfTestSupported  bool `json:"selective_self_test_supported"`
			AttributeAutoSaveEnabled    bool `json:"attribute_autosave_enabled"`
			ErrorLoggingSupported       bool `json:"error_logging_supported"`
			GpLoggingSupported          bool `json:"gp_logging_supported"`
		}
	} `json:"ata_smart_data"`

	ATASmartAttributes *ataSMARTAttributes `json:"ata_smart_attributes"`

	ATASmartSelectiveSelfTestLog *struct {
		Revision int `json:"revision"`

		Flags struct {
			ReminderScanEnabled bool `json:"reminder_scan_enabled"`
		} `json:"flags"`
		PowerUpScanResumeMinutes int `json:"power_up_scan_resume_minutes"`
	} `json:"ata_smart_selective_self_test_log"`
}

var smartctlVersionRegexp = regexp.MustCompile(`^smartctl\s(\d.\d)\s(\w|\W)+$`)

// Parse detect hardware disks and parse their S.M.A.R.T
func (sm *SMART) Parse() (common.MeasurementsMap, []error) {
	rawDisksOutput, err := sm.detectDisks()
	if err != nil {
		return nil, []error{err}
	}

	var disks []string
	if disks, err = parseDisks(rawDisksOutput); err != nil {
		return nil, []error{err}
	}

	var errs []error
	var jsonOutput []string
	if jsonOutput, err = sm.smartCtlRun(disks); err != nil {
		errs = append(errs, err)
	}

	result, parseErrors := smartCtlParse(jsonOutput)

	return result, append(errs, parseErrors...)
}

func (sm *SMART) smartCtlRun(disks []string) ([]string, error) {
	var result []string
	var errStr string

	for _, disk := range disks {
		cmd := sm.smartctlPrepare(disk)

		var err error
		var output []byte
		// The exit statuses of smartctl are defined by a bitmask.
		// If all is well with the disk, the exit status (return value) of smartctl is 0 (all bits turned off).
		// If a problem occurs, or an error, potential error, or fault is detected, then a non-zero status is returned
		// https://www.smartmontools.org/browser/trunk/smartmontools/smartctl.8.in
		if output, err = cmd.CombinedOutput(); err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				if status, ok := exitErr.Sys().(syscall.WaitStatus); ok && status.ExitStatus() == 1 {
					var errResult smartErrorResult
					if err = json.Unmarshal(output, &errResult); err == nil {
						var messages string
						for _, msg := range errResult.Smartctl.Messages {
							if messages != "" {
								messages += ";"
							}
							messages += msg.String
						}
						if errStr != "" {
							errStr += "\n"
						}
						errStr += messages
					}
					continue
				}
			}
		}

		result = append(result, string(output))
	}

	if errStr != "" {
		return result, errors.New("smart: " + errStr)
	}

	return result, nil
}

func smartCtlParse(raw []string) (common.MeasurementsMap, []error) {
	var parsedDisks []*parseResult

	var errs []error
	for _, r := range raw {
		res := &parseResult{}
		err := json.Unmarshal([]byte(r), res)
		if err != nil {
			errs = append(errs, errors.Wrap(err, "smart: unmarshal disk"))
		} else {
			parsedDisks = append(parsedDisks, res)
		}
	}

	marshaledDisks := make(common.MeasurementsMap)

	for _, disk := range parsedDisks {
		output := make(map[string]interface{})

		diskName := parseBase(output, disk)

		switch disk.Device.Protocol {
		case "NVMe":
		case "ATA":
			if disk.ATASmartAttributes != nil {
				parseATAAttributes(output, disk.ATASmartAttributes)
			}
		default:
		}

		marshaledDisks[diskName] = output
	}

	return marshaledDisks, errs
}

func smartctlParseVersion(buildStr string) (int, int, error) {
	if !smartctlVersionRegexp.Match([]byte(buildStr)) {
		return 0, 0, errors.New("smart: couldn't detect smartctl version")
	}

	ver := smartctlVersionRegexp.FindAllStringSubmatch(buildStr, -1)

	if len(ver) < 1 && len(ver[0]) < 2 {
		return 0, 0, ErrParseSmartctlVersion
	}

	tok := strings.Split(ver[0][1], ".")

	if len(tok) != 2 {
		return 0, 0, ErrParseSmartctlVersion
	}

	var major int
	var minor int
	var err error

	if major, err = strconv.Atoi(tok[0]); err != nil {
		return 0, 0, fmt.Errorf("smart: parse smartctl version: %s", err)
	}

	if minor, err = strconv.Atoi(tok[1]); err != nil {
		return 0, 0, fmt.Errorf("smart: parse smartctl version: %s", err)
	}

	return major, minor, nil
}

func parseBase(output map[string]interface{}, d *parseResult) string {
	output["device_type"] = d.Device.Type
	output["device_protocol"] = d.Device.Protocol
	output["model_name"] = d.ModelName
	output["model_family"] = ""
	if d.ModelFamily != nil {
		output["model_family"] = *d.ModelFamily
	}
	output["serial_number"] = d.SerialNumber
	output["firmware_version"] = d.FirmwareVersion
	output["power_cycle_count"] = 0
	if d.PowerCycleCount != nil {
		output["power_cycle_count"] = *d.PowerCycleCount
		output["power_cycle_count"] = *d.PowerCycleCount
	}

	if d.SmartStatus == nil {
		output["smart_status"] = "NOT AVAILABLE"
	} else if d.SmartStatus.Passed {
		output["smart_status"] = "PASSED"
	} else {
		output["smart_status"] = "FAILED"
	}

	if d.Temperature != nil {
		output["temperature_C"] = d.Temperature.Current
	}

	if d.PowerOnTime != nil {
		output["power_on_time_hours"] = d.PowerOnTime.Hours
	}

	switch rt := d.RotationRate.(type) {
	case nil:
		output["type_of"] = "SSD"
	case string:
		if rt == "Solid State Drive" || rt == "Solid State Device" {
			output["type_of"] = "SSD"
		} else {
			output["type_of"] = "HDD"
			output["rotation_rate"] = rt
		}
	case int:
		if rt == 0 {
			output["type_of"] = "SSD"
		} else {
			output["type_of"] = "HDD"
			output["rotation_rate"] = rt
		}
	default:
		log.Warnf("smart: parse type \"%s\" of rotation rate field of disk \"d.Device.Name\"", reflect.TypeOf(d.RotationRate).String())
		output["type_of"] = "SSD"
	}

	if d.InterfaceSpeed != nil {
		output["interface_speed_Bps"] = int64((d.InterfaceSpeed.Max.BitsPerUnit * d.InterfaceSpeed.Max.UnitsPerSecond) / 8)
	}

	return d.Device.Name
}

func parseATAAttributes(output map[string]interface{}, d *ataSMARTAttributes) {
	for _, at := range d.Table {
		if at.ID == 5 {
			output["reallocated_sector_count"] = at.Raw.Value
			break
		}
	}
}
