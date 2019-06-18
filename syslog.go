// +build !windows,!nacl,!plan9
// OS list copied from log/syslog

package cagent

import (
	"fmt"
	"log/syslog"
	"net/url"

	"github.com/sirupsen/logrus"
	lSyslog "github.com/sirupsen/logrus/hooks/syslog"
)

const udpPort = "514"

func addSyslogHook(syslogURL string) error {

	var network, raddr string

	if syslogURL != "local" {
		u, err := url.Parse(syslogURL)
		if err != nil {
			return fmt.Errorf("Wrong format of syslogURL: %s", err.Error())
		}
		network = u.Scheme
		raddr = u.Host

		if u.Port() == "" {
			raddr += ":;" + udpPort
		}
	}

	hook, err := lSyslog.NewSyslogHook(network, raddr, syslog.LOG_WARNING, "frontman")

	if err != nil {
		return err
	}

	logrus.AddHook(hook)

	return nil
}
