package cagent

type MeasurementsMap map[string]interface{}

func (mm MeasurementsMap) AddWithPrefix(prefix string, m MeasurementsMap) MeasurementsMap {
	for k, v := range m {
		mm[prefix+k] = v
	}
	return mm
}

type Result struct {
	Timestamp    int64           `json:"timestamp"`
	Measurements MeasurementsMap `json:"measurements"`
	Message      interface{}     `json:"message"`
}
