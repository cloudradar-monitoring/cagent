package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/pkg/errors"
)

type MeasurementsMap map[string]interface{}

func (mm MeasurementsMap) AddWithPrefix(prefix string, m MeasurementsMap) MeasurementsMap {
	if m == nil {
		return mm
	}

	for k, v := range m {
		mm[prefix+k] = v
	}
	return mm
}

func (mm MeasurementsMap) AddInnerWithPrefix(prefix string, m MeasurementsMap) MeasurementsMap {
	if m == nil {
		return mm
	}

	mm[prefix] = m

	return mm
}

// Timestamp type allows marshaling time.Time struct as Unix timestamp value
type Timestamp time.Time

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(*t).Unix()
	stamp := fmt.Sprint(ts)

	return json.Marshal(stamp)
}

func (t *Timestamp) UnmarshalJSON(raw []byte) error {
	var strTimestamp string
	if err := json.Unmarshal(raw, &strTimestamp); err != nil {
		return err
	}

	timestamp, err := strconv.ParseInt(strTimestamp, 10, 0)
	if err != nil {
		return errors.Wrap(err, "input is not Unix timestamp")
	}

	*t = Timestamp(time.Unix(timestamp, 0))
	return nil
}
