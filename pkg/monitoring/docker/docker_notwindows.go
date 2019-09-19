// +build !windows

package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type dockerPsOutput struct {
	ID     string
	Image  string
	Status string
	Names  string
}

const dockerAvailabilityCheckCacheExpiration = 1 * time.Minute

var dockerIsAvailable bool
var dockerAvailabilityLastRequestedAt *time.Time

// isDockerAvailable maintains a simple cache to prevent executing shell commands too often
func isDockerAvailable() bool {
	now := time.Now()
	if dockerAvailabilityLastRequestedAt != nil &&
		now.Sub(*dockerAvailabilityLastRequestedAt) < dockerAvailabilityCheckCacheExpiration {
		return dockerIsAvailable
	}

	_, err := exec.LookPath("docker")
	dockerIsAvailable = err == nil

	if dockerIsAvailable {
		err = exec.Command("/bin/sh", "-c", "sudo docker info").Run()
		if err != nil {
			log.WithError(err).Debug("while executing 'docker info' to check if docker is available")
		}
		dockerIsAvailable = dockerIsAvailable && (err == nil)
	}

	dockerAvailabilityLastRequestedAt = &now
	return dockerIsAvailable
}

func containerStatusToState(status string) string {
	status = strings.ToLower(status)

	if strings.HasPrefix(status, "up") {
		if strings.HasSuffix(status, "(paused)") {
			return "paused"
		}

		return "running"
	}

	if strings.HasPrefix(status, "exited") {
		return "stopped"
	}

	// in other cases just use the first word
	// As of docker 18.09 it can be one of: created, restarting, dead, removing
	p := strings.Split(status, " ")
	if len(p) > 0 {
		return p[0]
	}

	// just in case we got empty status somehow
	return "unknown"
}

// ListContainers returns the parsed output of 'docker ps' command
func ListContainers() (map[string]interface{}, error) {
	if !isDockerAvailable() {
		return nil, ErrorDockerNotAvailable
	}

	out, err := exec.Command("/bin/sh", "-c", "sudo docker ps -a --format \"{{ json . }}\"").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			err = errors.New(ee.Error() + ": " + string(ee.Stderr))
		}
		log.WithError(err).Error("can't list containers")
		return nil, err
	}

	lines := strings.Split(string(out), "\n")
	var containersResults []map[string]interface{}

	for _, line := range lines {
		// skip empty lines
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		var container dockerPsOutput
		err := json.Unmarshal([]byte(line), &container)
		if err != nil {
			log.WithError(err).Error("container list error: error decoding json output")
			continue
		}

		containersResults = append(containersResults, map[string]interface{}{
			"id":     container.ID,
			"image":  container.Image,
			"name":   container.Names,
			"state":  containerStatusToState(container.Status),
			"status": container.Status,
		})
	}

	return map[string]interface{}{"containers": containersResults}, nil
}

// ContainerNameByID returns the name of a container identified by its id
func ContainerNameByID(id string) (string, error) {
	if !isDockerAvailable() {
		return "", ErrorDockerNotAvailable
	}

	out, err := exec.Command("/bin/sh", "-c", fmt.Sprintf("sudo docker inspect --format \"{{ .Name }}\" %s", id)).Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			err = errors.New(ee.Error() + ": " + string(ee.Stderr))
		}

		return "", err
	}

	// remove \n and possible spaces around
	name := strings.TrimSpace(string(out))

	// remove leading slash from the name
	name = strings.TrimPrefix(name, "/")

	return name, nil
}
