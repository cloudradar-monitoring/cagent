// +build windows

package docker

func ListContainers() (map[string]interface{}, error) {
	return nil, ErrorNotImplementedForOS
}

func ContainerNameByID(_ string) (string, error) {
	return "", ErrorNotImplementedForOS
}
