// +build windows

package cagent

import (
	log "github.com/sirupsen/logrus"
	"golang.org/x/sys/windows/registry"
)

func getProcessorModelName() (string, error) {
	registryKey, err := registry.OpenKey(registry.LOCAL_MACHINE, `HARDWARE\DESCRIPTION\System\CentralProcessor\0`, registry.QUERY_VALUE)
	if err != nil {
		return "", err
	}
	defer func() {
		err = registryKey.Close()
		if err != nil {
			log.Warnf("[SYSTEM] could not close registry key handler: %s", err.Error())
		}
	}()

	processorName, _, err := registryKey.GetStringValue("ProcessorNameString")
	if err != nil {
		return "", err
	}

	return processorName, nil
}
