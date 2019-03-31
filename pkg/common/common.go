package common

import "math"

func MergeStringMaps(mapA, mapB map[string]interface{}) map[string]interface{} {
	for k, v := range mapB {
		mapA[k] = v
	}
	return mapA
}

func RoundToTwoDecimalPlaces(v float64) float64 {
	return math.Round(v*100) / 100
}
