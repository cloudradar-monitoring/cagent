// +build !windows

package main

import (
	"errors"

	"github.com/cloudradar-monitoring/cagent"
)

func sendErrorNotification(_, _ string) error {
	return errors.New("Implemented only or Windows")
}

func sendSuccessNotification(_, _ string) error {
	return errors.New("Implemented only or Windows")
}

func handleToastFeedback(_ *cagent.Cagent, _ string) {
	// only for windows
}
