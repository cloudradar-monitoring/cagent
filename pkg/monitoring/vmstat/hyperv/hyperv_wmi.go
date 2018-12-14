// +build windows

package hyperv

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/StackExchange/wmi"

	log "github.com/sirupsen/logrus"
)

var guestNetworkRegexp *regexp.Regexp

func init() {
	guestNetworkRegexp = regexp.MustCompile(`^Microsoft:GuestNetwork\\(\w{8}-\w{4}-\w{4}-\w{4}-\w{12})\\(\w{8}-\w{4}-\w{4}-\w{4}-\w{12})$`)
}

type vmStat struct {
	InstanceName           string
	CPUWaitTimePerDispatch float64
	State                  string
	CPUUsagePercent        float32
	AssignedMemory         string
	Uptime                 string
	IPAddresses            []string
	Heartbeat              string
	GUID                   string
}

type EnabledState uint16
type HealthState uint16
type Heartbeat uint16

type msvm_ComputerSystem struct {
	ElementName           string
	InstallDate           string
	HealthState           HealthState
	EnabledState          EnabledState
	OtherEnabledState     string
	RequestedState        EnabledState
	TimeOfLastStateChange string
	Name                  string
	ProcessID             uint32
}

type msvm_SummaryInformation struct {
	Name                 string
	ElementName          string
	GuestOperatingSystem *string
	Version              string
	EnabledState         EnabledState
	HealthState          HealthState
	CreationTime         time.Time
	MemoryUsage          *uint64
	Uptime               uint64
	ProcessorLoad        *uint16
	Heartbeat            *Heartbeat
	NumberOfProcessors   uint16
}

type msvm_GuestNetworkAdapterConfiguration struct {
	InstanceID  string
	IPAddresses []string
}

type hypervStat struct {
	TotalCPUWaitTimePerDispatch float64
	List                        map[string]vmStat
}

var hypervPath = "\\Hyper-V Hypervisor Virtual processor(*)\\CPU Wait Time Per Dispatch"

func (im *impl) GetMeasurements() (map[string]interface{}, error) {
	meas := make(map[string]interface{})

	stat := &hypervStat{
		List: make(map[string]vmStat),
	}

	var countersErr error
	res, countersErr := im.watcher.GetFormattedQueryData(hypervPath)
	if countersErr == nil {
		// Iterate over the reported counter values
		for _, c := range res {
			// Some filtering...
			switch c.InstanceName {
			case "_Total":
				meas["cpu_wait_time_per_dispatch_total_ns"] = int64(c.Value)
			default:
				name := strings.Split(c.InstanceName, ":")[0]
				stat.List[name] = vmStat{
					InstanceName:           name,
					CPUWaitTimePerDispatch: c.Value,
				}
			}
		}
	} else {
		fmt.Println(countersErr.Error())
		log.WithError(countersErr)
	}

	var list []interface{}

	ipMap := make(map[string][]string)

	var q string
	var ips []msvm_GuestNetworkAdapterConfiguration
	q = wmi.CreateQuery(&ips, "")
	if err := wmi.QueryNamespace(q, &ips, `root\virtualization\v2`); err == nil {
		for i := range ips {
			// InstanceID is in form of Microsoft:GuestNetwork\GUID\GUID
			// First GUID matches field Name in struct msvm_SummaryInformation
			if !guestNetworkRegexp.Match([]byte(ips[i].InstanceID)) {
				log.Warn("invalid InstanceID in hyper-v response. expected form:"+
					"[Microsoft:GuestNetwork\\<GUID>\\<GUID>] actual form: [%s]", ips[i].InstanceID)
			} else {
				id := guestNetworkRegexp.FindAllStringSubmatch(ips[i].InstanceID, -1)
				ipMap[id[0][1]] = ips[i].IPAddresses
			}
		}
	} else {
		log.WithError(err)
	}

	var dst []msvm_SummaryInformation
	q = wmi.CreateQuery(&dst, "")

	if err := wmi.QueryNamespace(q, &dst, `root\virtualization\v2`); err != nil {
		log.WithError(err)
	} else {
		for i := range dst {
			vmEntry := make(map[string]interface{})
			// include perfcounters only when perf has data
			if countersErr == nil {
				vmEntry["cpu_wait_time_per_dispatch_ns"] = int64(stat.List[dst[i].ElementName].CPUWaitTimePerDispatch)
			}
			vmEntry["name"] = dst[i].ElementName
			vmEntry["guid"] = dst[i].Name
			vmEntry["guest_operating_system"] = dst[i].GuestOperatingSystem
			vmEntry["version"] = dst[i].Version
			vmEntry["enabled_state"] = dst[i].EnabledState.String()
			vmEntry["health_state"] = dst[i].HealthState.String()
			vmEntry["creation_time"] = dst[i].CreationTime
			// dst[i].MemoryUsage is in megabytes
			vmEntry["assigned_memory_B"] = *dst[i].MemoryUsage * 0x100000
			vmEntry["uptime_s"] = dst[i].Uptime
			vmEntry["processor_load_percent"] = dst[i].ProcessorLoad
			vmEntry["heartbeat"] = dst[i].Heartbeat.String()
			vmEntry["number_of_processors"] = dst[i].NumberOfProcessors

			ipv4Count := 1
			ipv6Count := 1
			for j := range ipMap[dst[i].Name] {
				ip := net.ParseIP(ipMap[dst[i].Name][j])
				if p4 := ip.To4(); len(p4) == net.IPv4len {
					vmEntry["ipv4."+strconv.Itoa(ipv4Count)] = ipMap[dst[i].Name][j]
					ipv4Count++
				} else {
					vmEntry["ipv6."+strconv.Itoa(ipv6Count)] = ipMap[dst[i].Name][j]
					ipv6Count++
				}
			}

			list = append(list, vmEntry)
		}
	}

	meas["list"] = list

	return meas, nil
}

func (st EnabledState) String() string {
	switch st {
	case 2:
		return "running"
	case 3:
		return "shutdown"
	case 4:
		return "shutting down"
	case 6:
		return "offline"
	case 7:
		return "in test"
	case 8:
		return "deferred"
	case 9:
		return "quiesce"
	case 10:
		return "starting"
	default:
		return "<unknown>"
	}
}

func (st HealthState) String() string {
	switch st {
	case 5:
		return "ok"
	case 20:
		return "major failure"
	case 25:
		return "critical failure"
	default:
		return "<unknown>"
	}
}

func (st Heartbeat) String() string {
	switch st {
	case 2:
		return "ok"
	case 6:
		return "error"
	case 12:
		return "no contact"
	case 13:
		return "lost communication"
	default:
		return "unknown"
	}
}
