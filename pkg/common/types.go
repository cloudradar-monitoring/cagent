package common

import (
	"fmt"
	"time"
)

type MeasurementsMap map[string]interface{}

func (mm MeasurementsMap) AddWithPrefix(prefix string, m MeasurementsMap) MeasurementsMap {
	for k, v := range m {
		mm[prefix+k] = v
	}
	return mm
}

func (mm MeasurementsMap) AddInnerWithPrefix(prefix string, m MeasurementsMap) MeasurementsMap {
	mm[prefix] = m

	return mm
}

// Timestamp type allows marshaling time.Time struct as Unix timestamp value
type Timestamp time.Time

func (t *Timestamp) MarshalJSON() ([]byte, error) {
	ts := time.Time(*t).Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}
