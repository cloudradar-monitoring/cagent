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
		ResponseData controllerResponseData `json:"Response Data"`
	} `json:"Controllers"`
}

type controllerResponseData struct {
	// we can't parse some nested fields due to inconsistency (same field can have different types)
	// so we use json.RawMessage

	Status         map[string]*json.RawMessage   `json:"Status"`
	VirtualDrives  []map[string]*json.RawMessage `json:"VD LIST"`
	PhysicalDrives []map[string]*json.RawMessage `json:"PD LIST"`
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

func tryParseCmdOutput(outBytes *[]byte) (
	measurements map[string]interface{},
	alerts []monitoring.Alert,
	warnings []monitoring.Warning,
	err error,
) {
	var output controllersResult
	err = json.Unmarshal(*outBytes, &output)
	if err != nil {
		err = errors.Wrap(err, "while Unmarshal storcli cmd output")
		return
	}

	if len(output.Controllers) < 1 {
		err = errors.New("unexpected json: no controllers listed")
		return
	}

	responseData := output.Controllers[0].ResponseData

	status := responseData.Status
	measurements = map[string]interface{}{}
	measurements["Status"] = status

	// If the status is not Optimal an alert with the status is created.
	var controllerStatus string
	controllerStatus, err = extractFieldFromRawMap(&status, "Controller Status")
	if err != nil {
		return
	}
	if controllerStatus != statusOptimal {
		alerts = append(alerts, monitoring.Alert(fmt.Sprintf("Controller status not optimal (%s)", controllerStatus)))
	}

	// If one of the virtual disks is not in operational status, an alert with all details is created.
	for _, vd := range responseData.VirtualDrives {
		var vdState string
		vdState, err = extractFieldFromRawMap(&vd, "State")
		if err != nil {
			return
		}
		if vdState != stateOperational {
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

			vdStatusAlertMsg := fmt.Sprintf(
				"DG/VD %s %s State not operational (%s)",
				dgVD,
				vdType,
				vdState,
			)
			alerts = append(alerts, monitoring.Alert(vdStatusAlertMsg))
		}

	}

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
