package cagent

import (
	"context"
	"errors"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	utilnet "github.com/shirou/gopsutil/net"
	log "github.com/sirupsen/logrus"
)

const fsGetNetInterfacesTimeout = time.Second * 10

type netWatcher struct {
	cagent           *Cagent
	lastIOCounters   []utilnet.IOCountersStat
	lastIOCountersAt *time.Time

	ExcludedInterfaceCache map[string]bool
}

func (ca *Cagent) NetWatcher() *netWatcher {
	return &netWatcher{cagent: ca, ExcludedInterfaceCache: map[string]bool{}}
}

func (nw *netWatcher) Results() (MeasurementsMap, error) {
	results := MeasurementsMap{}

	var errs []string
	ctx, cancel := context.WithTimeout(context.Background(), fsGetNetInterfacesTimeout)
	defer cancel()

	interfaces, err := utilnet.Interfaces()

	if err != nil {
		log.Errorf("[NET] Failed to read interfaces: %s", err.Error())
		errs = append(errs, err.Error())
	}
		
	netInterfaceExcludeRegexCompiled := []*regexp.Regexp{}

	if len(nw.cagent.NetInterfaceExcludeRegex) > 0 {
		for _, reString := range nw.cagent.NetInterfaceExcludeRegex {
			re, err := regexp.Compile(reString)

			if err != nil {
				log.Errorf("[NET] net_interface_exclude_regex regexp '%s' compile error: %s", reString, err.Error())
				continue
			}
			netInterfaceExcludeRegexCompiled = append(netInterfaceExcludeRegexCompiled, re)
		}
	}

	for _, netIf := range interfaces {
		if isExcluded, cacheExists := nw.ExcludedInterfaceCache[netIf.Name]; cacheExists {
			if isExcluded {
				log.Debugf("[NET] interface excluded: %s", netIf.Name)
				continue
			}
		} else {
			isExcluded := false

			if nw.cagent.NetInterfaceExcludeLoopback == true {
				loopback := false
				for _, flag := range netIf.Flags {
					if flag == "loopback" {
						loopback = true
						break
					}
				}

				if loopback {
					isExcluded = true
				}
			}

			if !isExcluded && nw.cagent.NetInterfaceExcludeDisconnected {
				up := false
				for _, flag := range netIf.Flags {
					if flag == "up" {
						up = true
						break
					}
				}
				if !up {
					isExcluded = true
				}
			}

			if !isExcluded {
				for _, excludedIf := range nw.cagent.NetInterfaceExclude {
					if strings.EqualFold(netIf.Name, excludedIf) {
						isExcluded = true
						break
					}
				}
			}

			if !isExcluded {
				for _, re := range netInterfaceExcludeRegexCompiled {
					if re.MatchString(netIf.Name) {
						isExcluded = true
						break
					}
				}
			}

			nw.ExcludedInterfaceCache[netIf.Name] = isExcluded

			if isExcluded {
				log.Debugf("[NET] interface excluded: %s", netIf.Name)
				continue
			}
		}
	}

	ctx, _ = context.WithTimeout(context.Background(), fsGetUsageTimeout)
	counters, err := utilnet.IOCountersWithContext(ctx, true)

	if err != nil {
		log.Errorf("[NET] Failed to read IOCounters: %s", err.Error())
		errs = append(errs, err.Error())
		for _, netIf := range interfaces {
			for _, metric := range nw.cagent.NetMetrics {
				results[metric+"."+netIf.Name] = nil
			}
		}
	} else {
		gotIOCountersAt := time.Now()
		if nw.lastIOCounters != nil {
			for _, counter := range counters {
				if ifIncluded, exists := nw.ExcludedInterfaceCache[counter.Name]; exists && !ifIncluded {
					var previousIOCounters *utilnet.IOCountersStat
					for _, lastIOCounter := range nw.lastIOCounters {
						if lastIOCounter.Name == counter.Name {
							previousIOCounters = &lastIOCounter
							break
						}
					}

					if previousIOCounters == nil {
						log.Errorf("[NET] Previous IOCounters stat not found: %s", counter.Name)
						continue
					}

					for _, metric := range nw.cagent.NetMetrics {
						switch metric {
						case "in_B_per_s":
							results[metric+"."+counter.Name] = int64(float64(counter.BytesRecv-previousIOCounters.BytesRecv)/gotIOCountersAt.Sub(*nw.lastIOCountersAt).Seconds() + 0.5)
						case "out_B_per_s":
							results[metric+"."+counter.Name] = int64(float64(counter.BytesSent-previousIOCounters.BytesSent)/gotIOCountersAt.Sub(*nw.lastIOCountersAt).Seconds() + 0.5)
						case "errors_per_s":
							results[metric+"."+counter.Name] = int64(float64(counter.Errin+counter.Errout-previousIOCounters.Errin-previousIOCounters.Errout)/gotIOCountersAt.Sub(*nw.lastIOCountersAt).Seconds() + 0.5)
						case "dropped_per_s":
							results[metric+"."+counter.Name] = int64(float64(counter.Dropin+counter.Dropout-previousIOCounters.Dropin-previousIOCounters.Dropout)/gotIOCountersAt.Sub(*nw.lastIOCountersAt).Seconds() + 0.5)
						}
					}
				}
			}
		} else {
			log.Debugf("[NET] IO stat is available starting at 2nd check")
			for _, netIf := range interfaces {
				if isExcluded, exists := nw.ExcludedInterfaceCache[netIf.Name]; exists && isExcluded {
					continue
				}
				for _, metric := range nw.cagent.NetMetrics {
					results[metric+"."+netIf.Name] = nil
				}
			}
		}

		nw.lastIOCounters = counters
		nw.lastIOCountersAt = &gotIOCountersAt
	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errors.New("NET: " + strings.Join(errs, "; "))
}

func IPAddresses() (MeasurementsMap, error) {
	var addresses []string

	// Fetch all interfaces
	interfaces, err := utilnet.Interfaces()
	if err != nil {
		return nil, err
	}

	// Check all interfaces for their addresses
INFLOOP:
	for _, inf := range interfaces {
		for _, flag := range inf.Flags {
			// Ignore loopback addresses
			if flag == "loopback" {
				continue INFLOOP
			}
		}

		// Append all addresses to our slice
		for _, addr := range inf.Addrs {
			addresses = append(addresses, addr.Addr)
		}
	}

	result := make(MeasurementsMap)
	v4Count := uint32(1)
	v6Count := uint32(1)

	for _, address := range addresses {
		ipAddr, _, err := net.ParseCIDR(address)
		if err != nil {
			log.Warnf("Failed to parse IP address %s: %s", address, err)
			continue
		}
		// Check if ip is v4
		if v4 := ipAddr.To4(); v4 != nil {
			result[fmt.Sprintf("ipv4.%d", v4Count)] = v4.String()
			v4Count++
			continue
		}
		// Check if ip is v6
		if v6 := ipAddr.To16(); v6 != nil {
			result[fmt.Sprintf("ipv6.%d", v6Count)] = v6.String()
			v6Count++
			continue
		}

		log.Warnf("Could not determine if IP is v4 or v6: %s", address)
	}

	return result, nil
}
