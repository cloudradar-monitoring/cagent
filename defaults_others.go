// +build !windows,!darwin

package cagent

func init() {
	DefaultCfgPath = "/etc/cagent/cagent.conf"
	rootCertsPath = "/etc/cagent/cacert.pem"
	defaultLogPath = "/var/log/cagent/cagent.log"
}
