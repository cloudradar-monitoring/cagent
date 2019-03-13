// +build windows

package cagent

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/registry"
)

func getProcessorModelName() (string, error) {
	k, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer func() {
		err = k.Close()
		if err != nil {
			log.Warnf("[SYSTEM] could not close registry key handler: %s", err.Error())
		}
	}()

	processorName, _, err := k.GetStringValue("ProcessorNameString")
	if err != nil {
		return "", err
	}

	return processorName, nil
}
