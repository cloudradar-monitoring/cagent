package docker

import (
	"strings"

	docker "github.com/fsouza/go-dockerclient"
	log "github.com/sirupsen/logrus"
)

const dockerEndpointAddress = "unix:///var/run/docker.sock"

var client *docker.Client
var connectionSucceedOnce bool

func ListContainers() (map[string]interface{}, error) {
	var err error
	if client == nil {
		client, err = docker.NewClient(dockerEndpointAddress)
		if err != nil {
			return nil, err
		}
	}

	containers, err := client.ListContainers(docker.ListContainersOptions{All: false})
	if err != nil && err == docker.ErrConnectionRefused && !connectionSucceedOnce {
		// do not produce error if Docker is missing and wasn't successfully connected in this session before
		return nil, nil
	} else if err != nil {
		log.Errorf("[Docker] Failed to list containers: %s", err.Error())
		return nil, err
	}

	// store ths to handle 'connection refused' as error afterwards
	connectionSucceedOnce = true

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

func ContainerNameByID(id string) (string, error) {
	var err error
	if client == nil {
		client, err = docker.NewClient(dockerEndpointAddress)
		if err != nil {
			return "", err
		}
	}

	container, err := client.InspectContainer(id)
	if err != nil {
		return "", err
	}

	return container.Name, nil
}
