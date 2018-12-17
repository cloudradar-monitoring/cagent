// +build windows

package main

import (
	"os"
	"gopkg.in/toast.v1"
	"path/filepath"
)

const urlScheme = "cagent"
const toastErrorIcon = "resources\\error.png"
const toastSuccessIcon = "resources\\success.png"
const toastAppID = "cloudradar.cagent"

func getExecutablePath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}

	return filepath.Dir(ex)
}

func sendErrorNotification(title, message string) error {
	msg := toast.Notification{
		AppID:   toastAppID,
		Title:   title,
		Message: message,
		Actions: []toast.Action{
			{"protocol", "How to fix?", "https://kb.cloudradar.io/books/configuring-hosts/page/installing-agents#bkmrk-potential-problems-a"},
		},
	}

	iconPath := getExecutablePath() + "\\" + toastErrorIcon
	if _, err := os.Stat(iconPath); err == nil {
		msg.Icon = iconPath
	}
	return msg.Push()
}

func sendSuccessNotification(title, message string) error {
	msg := toast.Notification{
		AppID:   toastAppID,
		Title:   title,
		Message: message,
		Actions: []toast.Action{},
	}

	iconPath := getExecutablePath() + "\\" + toastSuccessIcon
	if _, err := os.Stat(iconPath); err == nil {
		msg.Icon = iconPath
	}
	return msg.Push()
}

func handleToastFeedback(ca *cagent.Cagent, cfgPath string) {
	// handle URL schema arguments on windows
	if runtime.GOOS != "windows" {
		return
	}

	if len(os.Args) < 2 {
		return
	}

	switch os.Args[1] {
	case urlScheme + ":test":
		toastCmdTest(ca)
	case urlScheme + ":config":
		toastOpenConfig(cfgPath)
	}
}

func toastCmdTest(ca *cagent.Cagent) {
	handleFlagTest(true, ca)
}

func toastOpenConfig(cfgPath string) error {
	r := strings.NewReplacer("&", "^&")
	cfgPath = r.Replace(cfgPath)
	defer os.Exit(1)
	return exec.Command("cmd", "/C", "start", "", "notepad", cfgPath).Start()
}
