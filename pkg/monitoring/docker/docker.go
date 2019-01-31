// +build !windows

package docker

import (
	"os/exec"
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
)

const dockerEndpointAddress = "unix:///var/run/docker.sock"

type DockerWatcher struct {
	client                *docker.Client
	connectionSucceedOnce bool
}

func (dw *DockerWatcher) initClient() error {
	if dw.client == nil {
		_, err := docker.NewClient(dockerEndpointAddress)
		if err != nil {
			return err
		}
	}
	return nil
}

func (dw *DockerWatcher) ListContainers() (map[string]interface{}, error) {
	err := dw.initClient()
	if err != nil {
		return nil, err
	}

	var dockerExecExists bool
	if _, err := exec.LookPath("docker"); err == nil {
		dockerExecExists = true
	}

	containers, err := dw.client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil &&
		!dockerExecExists && // docker executable not exists in the system
		!dw.connectionSucceedOnce && // connection with Docker via UNIX socket was never succeed in this session
		(strings.Contains(err.Error(), "no such file or directory") || err == docker.ErrConnectionRefused) {

		// do not produce error if Docker executable is missing and wasn't successfully connected in this session before
		return nil, nil
	} else if err != nil {
		log.Errorf("[Docker] Failed to list containers: %s", err.Error())
		return nil, err
	}

	// store ths to handle 'connection refused' as error afterwards
	dw.connectionSucceedOnce = true

	var containersResults []map[string]interface{}
	for _, container := range containers {
		containersResults = append(containersResults, map[string]interface{}{
			"id":     container.ID[0:12], // use short ID only as it enough to identify container
			"image":  container.Image,
			"name":   strings.Join(container.Names, ","),
			"state":  container.State,
			"status": container.Status,
		})
	}

	return map[string]interface{}{"containers": containersResults}, nil
}

func (dw *DockerWatcher) ContainerNameByID(id string) (string, error) {
	err := dw.initClient()
	if err != nil {
		return "", err
	}

	container, err := dw.client.InspectContainer(id)
	if err != nil {
		return "", err
	}

	return container.Name, nil
}
