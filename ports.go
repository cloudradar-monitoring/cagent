package cagent

import (
	"fmt"
	"github.com/shirou/gopsutil/net"
	log "github.com/sirupsen/logrus"
	"syscall"
)

type PortStat struct {
	Protocol     string `json:"proto"`
	LocalAddress string `json:"addr"`
	State        string `json:"state,omitempty"`
	PID          int32  `json:"pid,omitempty"`
	ProgramName  string `json:"program,omitempty"`
}

func (ca *Cagent) PortsResult(processList []ProcStat) (m MeasurementsMap, err error) {
	connections, err := net.Connections("inet")
	if err != nil {
		log.Error("[PORTS] error: ", err.Error())
		return nil, err
	}

	var ports []PortStat
	for _, conn := range connections {
		state := conn.Status
		if state == "NONE" {
			state = ""
		}
		if state != "LISTEN" && state != "" {
			continue
		}

		var programName string
		if conn.Pid != 0 {
			for _, proc := range processList {
				if int32(proc.PID) == conn.Pid {
					programName = proc.Name
					break
				}
			}
		}

		ports = append(ports, PortStat{
			Protocol:     formatConnectionProtocol(conn.Family, conn.Type),
			LocalAddress: formatNetAddr(&conn.Laddr),
			State:        state,
			PID:          conn.Pid,
			ProgramName:  programName,
		})
	}

	log.Info("[PORTS] results: ", len(ports))

	m = MeasurementsMap{"list": ports}
	return
}

func formatNetAddr(addr *net.Addr) string {
	return fmt.Sprintf("%s:%d", addr.IP, addr.Port)
}

func formatConnectionProtocol(family, socketType uint32) string {
	var version string
	if family == syscall.AF_INET6 {
		version = "6"
	}

	var baseProto string
	if socketType == syscall.SOCK_STREAM {
		baseProto = "tcp"
	} else {
		baseProto = "udp"
	}

	return baseProto + version
}
