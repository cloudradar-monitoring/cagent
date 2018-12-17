// +build !windows

package main

import (
	"errors"
	"github.com/cloudradar-monitoring/cagent"
)

func sendErrorNotification(title, message string) error {
	return errors.New("Implemented only or Windows")
}

func sendSuccessNotification(title, message string) error {
	return errors.New("Implemented only or Windows")
}

func handleToastFeedback(ca *cagent.Cagent, cfgPath string) {
	// only for windows
	return
}
