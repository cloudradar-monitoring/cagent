package storcli

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/pkg/errors"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
	"github.com/cloudradar-monitoring/cagent/pkg/monitoring"
)

type controllersResult struct {
	Controllers []struct {
		CommandStatus commandStatus          `json:"Command Status"`
		ResponseData  controllerResponseData `json:"Response Data"`
	} `json:"Controllers"`
}

type commandStatus struct {
	Status      string `json:"Status"`
	Description string `json:"Description"`
}

type controllerResponseData struct {
	// we can't parse some nested fields due to inconsistency (same field can have different types)
	// so we use json.RawMessage

	Basics              controllerBasicsData
	Status              map[string]*json.RawMessage   `json:"Status"`
	VirtualDrives       []map[string]*json.RawMessage `json:"VD LIST"`
	PhysicalDrivesCount int                           `json:"Physical Drives"`
	PhysicalDrives      []map[string]*json.RawMessage `json:"PD LIST"`
}

type controllerBasicsData struct {
	ControllerID int    `json:"Controller"`
	Model        string `json:"Model"`
	SerialNumber string `json:"Serial Number"`
}

func (c *controllerBasicsData) GetDisplayName() string {
	return fmt.Sprintf("%s %s", c.Model, c.SerialNumber)
}

const statusOptimal = "Optimal"
const stateOperational = "Optl"

var physDriveStatesMap = map[string]string{
	"UBad":   "Unconfigured Bad",
	"Offln":  "Offline",
	"UBUnsp": "UBad Unsupported",
	"UGUnsp": "Unsupported",

	"Onln":  "Online",
	"UGood": "Unconfigured Good",
	"GHS":   "Global Hot Spare",
	"DHS":   "Dedicated Hot Space",
	"JBOD":  "Just a Bunch of Disks",
}
var goodPhysicalDriveStates = []string{"Onln", "UGood", "GHS", "DHS", "JBOD"}

func getHumanReadablePhysDriveState(state string) string {
	res, exists := physDriveStatesMap[state]
	if !exists {
		return state
	}
	return res
}

func tryParseCmdOutput(outBytes *[]byte) (*controllersResult, error) {
	var output controllersResult
	err := json.Unmarshal(*outBytes, &output)
	if err != nil {
		err = errors.Wrap(err, "error while parsing storcli command output")
	}
	return &output, err
}

func getReportData(responseData *controllerResponseData) (
	measurements map[string]interface{},
	alerts []monitoring.Alert,
	warnings []monitoring.Warning,
	err error,
) {
	status := responseData.Status
	measurements = map[string]interface{}{}
	measurements["Status"] = status
	measurements["Number of attached physical drives"] = responseData.PhysicalDrivesCount

	// If the status is not Optimal an alert with the status is created.
	var controllerStatus string
	controllerStatus, err = extractFieldFromRawMap(&status, "Controller Status")
	if err != nil {
		return
	}
	if controllerStatus != statusOptimal {
		alerts = append(alerts, monitoring.Alert(fmt.Sprintf("Controller status not optimal (%s)", controllerStatus)))
	}

	var vdStates = make(map[string]string)

	// If one of the virtual disks is not in operational status, an alert with all details is created.
	for _, vd := range responseData.VirtualDrives {
		var vdState string
		vdState, err = extractFieldFromRawMap(&vd, "State")
		if err != nil {
			return
		}

		var dgVD string
		dgVD, err = extractFieldFromRawMap(&vd, "DG/VD")
		if err != nil {
			return
		}

		var vdType string
		vdType, err = extractFieldFromRawMap(&vd, "TYPE")
		if err != nil {
			return
		}

		var vdName string
		vdName, err = extractFieldFromRawMap(&vd, "Name")
		if err != nil {
			return
		}

		vdLabel := fmt.Sprintf("DG/VD %s %s %s", dgVD, vdType, vdName)
		vdStates[fmt.Sprintf("%s State", vdLabel)] = vdState

		if vdState != stateOperational {
			vdStatusAlertMsg := fmt.Sprintf("%s State not operational (%s)", vdLabel, vdState)
			alerts = append(alerts, monitoring.Alert(vdStatusAlertMsg))
		}
	}
	measurements["Virtual Drives"] = vdStates

	// If one of the physical disks is in bad state,
	// a warning with the details of the device is created.
	for _, pd := range responseData.PhysicalDrives {
		var pdState string
		pdState, err = extractFieldFromRawMap(&pd, "State")
		if err != nil {
			return
		}

		if !common.StrInSlice(pdState, goodPhysicalDriveStates) {
			humanReadableState := getHumanReadablePhysDriveState(pdState)
			var deviceID, eIDSlot, interfaceName, mediaType string

			deviceID, err = extractFieldFromRawMap(&pd, "DID")
			if err != nil {
				return
			}
			eIDSlot, err = extractFieldFromRawMap(&pd, "EID:Slt")
			if err != nil {
				return
			}
			interfaceName, err = extractFieldFromRawMap(&pd, "Intf")
			if err != nil {
				return
			}
			mediaType, err = extractFieldFromRawMap(&pd, "Med")
			if err != nil {
				return
			}

			warnMsg := fmt.Sprintf(
				"Physical device %s (%s) %s %s state is %s (%s)",
				deviceID, eIDSlot, interfaceName, mediaType, humanReadableState, pdState,
			)
			warnings = append(warnings, monitoring.Warning(warnMsg))
		}
	}

	return
}

// extractFieldFromRawMap tries to read map field as string value
// If it fails, tries to read as int value
// the return is always string
func extractFieldFromRawMap(raw *map[string]*json.RawMessage, key string) (string, error) {
	rawValue, exists := (*raw)[key]
	if !exists {
		return "", fmt.Errorf("unexpected json: %s field is not present", key)
	}
	var strValue string
	strUnmarshalErr := json.Unmarshal(*rawValue, &strValue)
	if strUnmarshalErr != nil {
		// try unmarshall int
		var intValue int
		err := json.Unmarshal(*rawValue, &intValue)
		if err != nil {
			return "", errors.Wrapf(strUnmarshalErr, "while retrieving %s from json", key)
		}
		strValue = strconv.Itoa(intValue)
	}

	return strValue, nil
}
