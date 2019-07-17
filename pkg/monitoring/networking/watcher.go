package networking

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	utilnet "github.com/shirou/gopsutil/net"
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

const netGetCountersTimeout = time.Second * 10

type NetWatcherConfig struct {
	NetInterfaceExclude             []string
	NetInterfaceExcludeRegex        []string
	NetInterfaceExcludeDisconnected bool
	NetInterfaceExcludeLoopback     bool
	NetMetrics                      []string
}

type NetWatcher struct {
	config NetWatcherConfig

	lastIOCounters   []utilnet.IOCountersStat
	lastIOCountersAt *time.Time

	netInterfaceExcludeRegexCompiled []*regexp.Regexp
	constantlyExcludedInterfaceCache map[string]bool
}

func NewWatcher(cfg NetWatcherConfig) *NetWatcher {
	return &NetWatcher{
		config:                           cfg,
		constantlyExcludedInterfaceCache: map[string]bool{},
	}
}

// InterfaceExcludeRegexCompiled compiles and cache all the interfaces-filtering regexp's user has specified in the Config
// So we don't need to compile them on each iteration of measurements
func (nw *NetWatcher) InterfaceExcludeRegexCompiled() []*regexp.Regexp {
	if len(nw.netInterfaceExcludeRegexCompiled) > 0 {
		return nw.netInterfaceExcludeRegexCompiled
	}

	if len(nw.config.NetInterfaceExcludeRegex) > 0 {
		for _, reString := range nw.config.NetInterfaceExcludeRegex {
			re, err := regexp.Compile(reString)

			if err != nil {
				logrus.Errorf("[NET] net_interface_exclude_regex regexp '%s' compile error: %s", reString, err.Error())
				continue
			}
			nw.netInterfaceExcludeRegexCompiled = append(nw.netInterfaceExcludeRegexCompiled, re)
		}
	}

	return nw.netInterfaceExcludeRegexCompiled
}

func (nw *NetWatcher) isInterfaceExcludedByName(netIf *utilnet.InterfaceStat) bool {
	for _, excludedIf := range nw.config.NetInterfaceExclude {
		if strings.EqualFold(netIf.Name, excludedIf) {
			return true
		}
	}
	return false
}

func (nw *NetWatcher) isInterfaceExcludedByRegexp(netIf *utilnet.InterfaceStat) bool {
	for _, re := range nw.InterfaceExcludeRegexCompiled() {
		if re.MatchString(netIf.Name) {
			return true
		}
	}
	return false
}

func (nw *NetWatcher) ExcludedInterfacesByName(allInterfaces []utilnet.InterfaceStat) map[string]struct{} {
	excludedInterfaces := map[string]struct{}{}

	for _, netIf := range allInterfaces {
		// use a cache for excluded interfaces, because all the checks(except UP/DOWN state) are constant for the same interface&Config
		if isExcluded, cacheExists := nw.constantlyExcludedInterfaceCache[netIf.Name]; cacheExists {
			if isExcluded ||
				nw.config.NetInterfaceExcludeDisconnected && isInterfaceDown(&netIf) {
				// interface is found excluded in the cache or has a DOWN state
				excludedInterfaces[netIf.Name] = struct{}{}
				logrus.Debugf("[NET] interface excluded: %s", netIf.Name)
				continue
			}
		} else {
			if nw.config.NetInterfaceExcludeLoopback && isInterfaceLoobpack(&netIf) ||
				nw.isInterfaceExcludedByName(&netIf) ||
				nw.isInterfaceExcludedByRegexp(&netIf) {
				// add the excluded interface to the cache because this checks are constant
				nw.constantlyExcludedInterfaceCache[netIf.Name] = true
			} else if nw.config.NetInterfaceExcludeDisconnected && isInterfaceDown(&netIf) {
				// exclude DOWN interface for now
				// lets cache it as false and then we will only check UP/DOWN status
				nw.constantlyExcludedInterfaceCache[netIf.Name] = false
			} else {
				// interface is not excluded
				nw.constantlyExcludedInterfaceCache[netIf.Name] = false
				continue
			}

			excludedInterfaces[netIf.Name] = struct{}{}
			logrus.Debugf("[NET] interface excluded: %s", netIf.Name)
		}
	}
	return excludedInterfaces
}

// fillEmptyMeasurements used to fill measurements with nil's for all non-excluded interfaces
// It is called in case measurements are not yet ready or some error happens while retrieving counters
func (nw *NetWatcher) fillEmptyMeasurements(results common.MeasurementsMap, interfaces []utilnet.InterfaceStat, excludedInterfacesByName map[string]struct{}) {
	for _, netIf := range interfaces {
		if _, isExcluded := excludedInterfacesByName[netIf.Name]; isExcluded {
			continue
		}

		for _, metric := range nw.config.NetMetrics {
			results[metric+"."+netIf.Name] = nil
		}
	}
}

// fillCountersMeasurements used to fill measurements with nil's for all non-excluded interfaces
func (nw *NetWatcher) fillCountersMeasurements(results common.MeasurementsMap, interfaces []utilnet.InterfaceStat, excludedInterfacesByName map[string]struct{}) error {
	ctx, cancelFn := context.WithTimeout(context.Background(), netGetCountersTimeout)
	defer cancelFn()

	counters, err := utilnet.IOCountersWithContext(ctx, true)
	if err != nil {
		// fill empty measurements for not-excluded interfaces
		nw.fillEmptyMeasurements(results, interfaces, excludedInterfacesByName)
		return fmt.Errorf("Failed to read IOCounters: %s", err.Error())
	}

	gotIOCountersAt := time.Now()
	defer func() {
		nw.lastIOCountersAt = &gotIOCountersAt
		nw.lastIOCounters = counters
	}()

	if nw.lastIOCounters == nil {
		logrus.Debugf("[NET] IO stat is available starting from 2nd check")
		nw.fillEmptyMeasurements(results, interfaces, excludedInterfacesByName)
		// do not need to return the error here, because this is normal behavior
		return nil
	}

	lastIOCounterByName := map[string]utilnet.IOCountersStat{}
	for _, lastIOCounter := range nw.lastIOCounters {
		lastIOCounterByName[lastIOCounter.Name] = lastIOCounter
	}

	for _, ioCounter := range counters {
		// iterate over all counters
		// each ioCounter corresponds to the specific interface with name ioCounter.Name
		if _, isExcluded := excludedInterfacesByName[ioCounter.Name]; isExcluded {
			continue
		}

		var previousIOCounter utilnet.IOCountersStat
		var exists bool
		// found prev counter data
		if previousIOCounter, exists = lastIOCounterByName[ioCounter.Name]; !exists {
			logrus.Errorf("[NET] Previous IOCounters stat not found: %s", ioCounter.Name)
			continue
		}

		secondsSinceLastMeasurement := gotIOCountersAt.Sub(*nw.lastIOCountersAt).Seconds()
		for _, metric := range nw.config.NetMetrics {
			switch metric {
			case "in_B_per_s":
				bytesReceivedSinceLastMeasurement := ioCounter.BytesRecv - previousIOCounter.BytesRecv
				results[metric+"."+ioCounter.Name] = common.FloatToIntRoundUP(float64(bytesReceivedSinceLastMeasurement) / secondsSinceLastMeasurement)
			case "out_B_per_s":
				bytesSentSinceLastMeasurement := ioCounter.BytesSent - previousIOCounter.BytesSent
				results[metric+"."+ioCounter.Name] = common.FloatToIntRoundUP(float64(bytesSentSinceLastMeasurement) / secondsSinceLastMeasurement)
			case "errors_per_s":
				errorsSinceLastMeasurement := ioCounter.Errin + ioCounter.Errout - previousIOCounter.Errin - previousIOCounter.Errout
				results[metric+"."+ioCounter.Name] = common.FloatToIntRoundUP(float64(errorsSinceLastMeasurement) / secondsSinceLastMeasurement)
			case "dropped_per_s":
				droppedSinceLastMeasurement := ioCounter.Dropin + ioCounter.Dropout - ioCounter.Dropin - previousIOCounter.Dropout
				results[metric+"."+ioCounter.Name] = common.FloatToIntRoundUP(float64(droppedSinceLastMeasurement) / secondsSinceLastMeasurement)
			}
		}
	}

	return nil
}

func (nw *NetWatcher) Results() (common.MeasurementsMap, error) {
	results := common.MeasurementsMap{}

	interfaces, err := utilnet.Interfaces()
	if err != nil {
		logrus.Errorf("[NET] Failed to read interfaces: %s", err.Error())
		return nil, err
	}

	excludedInterfacesByNameMap := nw.ExcludedInterfacesByName(interfaces)
	// fill counters measurements into results
	err = nw.fillCountersMeasurements(results, interfaces, excludedInterfacesByNameMap)
	if err != nil {
		logrus.Errorf("[NET] Failed to collect counters: %s", err.Error())
		return results, err
	}

	return results, nil
}
