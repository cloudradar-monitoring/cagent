// +build !windows

package docker

import (
	"encoding/json"
	"errors"
	"os/exec"
	"strings"

	log "github.com/sirupsen/logrus"
)

// Don't use this struct directly.
// Use New() instead
type Watcher struct {
	connectionSucceedOnce bool
}

func isDockerAvailable() bool {
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}

	return true
}

type dockerPsOutput struct {
	ID     string
	Image  string
	Status string
	Names  string
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

func (dw *Watcher) ListContainers() (map[string]interface{}, error) {
	if !isDockerAvailable() {
		return nil, ErrorDockerNotFound
	}

	out, err := exec.Command("/bin/sh", "-c", "sudo docker ps -a --format \"{{ json . }}\"").Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			err = errors.New(ee.Error() + ": " + string(ee.Stderr))
		}

		log.Errorf("[docker] docker executable exists but failed to list containers: %s", err.Error())

		// looks like docker daemon is down
		// don not pass the error in case we never succeed with docker command within a session
		if !dw.connectionSucceedOnce {
			return nil, nil
		}

		return nil, err
	}

	dw.connectionSucceedOnce = true

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
			log.Errorf("[docker] container list error: error decoding json output: %s", err.Error())
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

func (dw *Watcher) ContainerNameByID(id string) (string, error) {
	if !isDockerAvailable() {
		return "", ErrorDockerNotFound
	}

	out, err := exec.Command("/bin/sh", "-c", "sudo docker inspect --format \"{{ .Name }}\"").Output()
	if err != nil {
		// looks like docker daemon is down
		// don not pass the error in case we never succeed with docker command within a session
		if !dw.connectionSucceedOnce {
			return "", nil
		}

		if ee, ok := err.(*exec.ExitError); ok {
			err = errors.New(ee.Error() + ": " + string(ee.Stderr))
		}

		return "", err
	}

	dw.connectionSucceedOnce = true

	// remove \n and possible spaces around
	name := strings.TrimSpace(string(out))

	// remove leading slash from the name
	name = strings.TrimPrefix(name, "/")

	return name, nil
}
