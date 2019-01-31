// +build windows

package docker

type DockerWatcher struct{}

func (_ *DockerWatcher) ListContainers() (map[string]interface{}, error) {
	return nil, ErrorNotImplementedForOS
}

func (_ *DockerWatcher) ContainerNameByID(_ string) (string, error) {
	return "", ErrorNotImplementedForOS
}
