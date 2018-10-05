package cagent

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/shirou/gopsutil/net"
	log "github.com/sirupsen/logrus"
)

const fsGetNetInterfacesTimeout = time.Second * 10

type netWatcher struct {
	cagent           *Cagent
	lastIOCounters   []net.IOCountersStat
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

	interfaces, err := net.Interfaces()

	if err != nil {
		log.Errorf("[NET] Failed to read interfaces: %s", err.Error())
		errs = append(errs, err.Error())
	}
	var netInterfaceExcludeRegexCompiled []*regexp.Regexp

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
		if nw.cagent.NetInterfaceExcludeDisconnected {
			up := false
			for _, flag := range netIf.Flags {
				if flag == "up" {
					up = true
					break
				}
			}
			if !up {
				log.Debugf("[NET] down interface excluded: %s", netIf.Name)
				continue
			}
		}

		if ifExcluded, cacheExists := nw.ExcludedInterfaceCache[netIf.Name]; cacheExists {
			if ifExcluded {
				log.Debugf("[NET] interface excluded: %s", netIf.Name)
				continue
			}
		} else {
			if nw.cagent.NetInterfaceExcludeLoopback {
				loopback := false
				for _, flag := range netIf.Flags {
					if flag == "loopback" {
						loopback = true
						break
					}
				}
				nw.ExcludedInterfaceCache[netIf.Name] = loopback
				if loopback {
					log.Debugf("[NET] loopback interface excluded: %s", netIf.Name)
					continue
				}
			}
			ifExcluded := false
			for _, excludedIf := range nw.cagent.NetInterfaceExclude {
				if strings.Contains(strings.ToLower(netIf.Name), strings.ToLower(excludedIf)) {
					ifExcluded = true
					break
				}
			}

			if !ifExcluded {
				for _, re := range netInterfaceExcludeRegexCompiled {
					if re.MatchString(netIf.Name) {
						ifExcluded = true
						break
					}
				}
			}

			nw.ExcludedInterfaceCache[netIf.Name] = ifExcluded

			if ifExcluded {
				log.Debugf("[NET] interface excluded: %s", netIf.Name)
				continue
			}
		}
	}

	ctx, _ = context.WithTimeout(context.Background(), fsGetUsageTimeout)
	counters, err := net.IOCountersWithContext(ctx, true)

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
					var previousIOCounters *net.IOCountersStat
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
