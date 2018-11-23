// +build darwin

package cagent

import (
	"os"
)

func init() {
	DefaultCfgPath = os.Getenv("HOME") + "/.cagent/cagent.conf"
	defaultLogPath = os.Getenv("HOME") + "/.cagent/cagent.log"
}
