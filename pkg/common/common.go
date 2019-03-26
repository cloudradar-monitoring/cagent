package common

func MergeStringMaps(mapA, mapB map[string]interface{}) map[string]interface{} {
	for k, v := range mapB {
		mapA[k] = v
	}
	return mapA
}
