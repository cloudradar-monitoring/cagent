package cagent

import (
	"context"
	"fmt"
	"net"
	"regexp"
	"strings"
	"time"

	utilnet "github.com/shirou/gopsutil/net"
	log "github.com/sirupsen/logrus"
)

const netGetCountersTimeout = time.Second * 10

type netWatcher struct {
	cagent           *Cagent
	lastIOCounters   []utilnet.IOCountersStat
	lastIOCountersAt *time.Time

	netInterfaceExcludeRegexCompiled []*regexp.Regexp
	constantlyExcludedInterfaceCache map[string]bool
}

func (ca *Cagent) NetWatcher() *netWatcher {
	if ca.netWatcher != nil {
		return ca.netWatcher
	}

	ca.netWatcher = &netWatcher{cagent: ca, constantlyExcludedInterfaceCache: map[string]bool{}}
	return ca.netWatcher
}

// InterfaceExcludeRegexCompiled compiles and cache all the interfaces-filtering regexp's user has specified in the config
// So we don't need to compile them on each iteration of measurements
func (nw *netWatcher) InterfaceExcludeRegexCompiled() []*regexp.Regexp {
	if len(nw.netInterfaceExcludeRegexCompiled) > 0 {
		return nw.netInterfaceExcludeRegexCompiled
	}

	if len(nw.cagent.NetInterfaceExcludeRegex) > 0 {
		for _, reString := range nw.cagent.NetInterfaceExcludeRegex {
			re, err := regexp.Compile(reString)

			if err != nil {
				log.Errorf("[NET] net_interface_exclude_regex regexp '%s' compile error: %s", reString, err.Error())
				continue
			}
			nw.netInterfaceExcludeRegexCompiled = append(nw.netInterfaceExcludeRegexCompiled, re)
		}
	}

	return nw.netInterfaceExcludeRegexCompiled
}

func isInterfaceLoobpack(netIf *utilnet.InterfaceStat) bool {
	for _, flag := range netIf.Flags {
		if flag == "loopback" {
			return true
		}
	}

	return false
}

func isInterfaceDown(netIf *utilnet.InterfaceStat) bool {
	for _, flag := range netIf.Flags {
		if flag == "up" {
			return false
		}
	}

	return true
}

func (nw *netWatcher) isInterfaceExcludedByName(netIf *utilnet.InterfaceStat) bool {
	for _, excludedIf := range nw.cagent.NetInterfaceExclude {
		if strings.EqualFold(netIf.Name, excludedIf) {
			return true
		}
	}
	return false
}

func (nw *netWatcher) isInterfaceExcludedByRegexp(netIf *utilnet.InterfaceStat) bool {
	for _, re := range nw.InterfaceExcludeRegexCompiled() {
		if re.MatchString(netIf.Name) {
			return true
		}
	}
	return false
}

func (nw *netWatcher) ExcludedInterfacesByName(allInterfaces []utilnet.InterfaceStat) map[string]struct{} {
	excludedInterfaces := map[string]struct{}{}

	for _, netIf := range allInterfaces {
		// use a cache for excluded interfaces, because all the checks(except UP/DOWN state) are constant for the same interface&config
		if isExcluded, cacheExists := nw.constantlyExcludedInterfaceCache[netIf.Name]; cacheExists {
			if isExcluded ||
				nw.cagent.NetInterfaceExcludeDisconnected && isInterfaceDown(&netIf) {
				// interface is found excluded in the cache or has a DOWN state
				excludedInterfaces[netIf.Name] = struct{}{}
				log.Debugf("[NET] interface excluded: %s", netIf.Name)
				continue
			}
		} else {
			if nw.cagent.NetInterfaceExcludeLoopback && isInterfaceLoobpack(&netIf) ||
				nw.isInterfaceExcludedByName(&netIf) ||
				nw.isInterfaceExcludedByRegexp(&netIf) {
				// add the excluded interface to the cache because this checks are constant
				nw.constantlyExcludedInterfaceCache[netIf.Name] = true
			} else if nw.cagent.NetInterfaceExcludeDisconnected && isInterfaceDown(&netIf) {
				// exclude DOWN interface for now
				// lets cache it as false and then we will only check UP/DOWN status
				nw.constantlyExcludedInterfaceCache[netIf.Name] = false
			} else {
				// interface is not excluded
				nw.constantlyExcludedInterfaceCache[netIf.Name] = false
				continue
			}

			excludedInterfaces[netIf.Name] = struct{}{}
			log.Debugf("[NET] interface excluded: %s", netIf.Name)
		}
	}
	return excludedInterfaces
}

// fillEmptyMeasurements used to fill measurements with nil's for all non-excluded interfaces
// It is called in case measurements are not yet ready or some error happens while retrieving counters
func (nw *netWatcher) fillEmptyMeasurements(results MeasurementsMap, interfaces []utilnet.InterfaceStat, excludedInterfacesByName map[string]struct{}) {
	for _, netIf := range interfaces {
		if _, isExcluded := excludedInterfacesByName[netIf.Name]; isExcluded {
			continue
		}

		for _, metric := range nw.cagent.NetMetrics {
			results[metric+"."+netIf.Name] = nil
		}
	}
}

// fillCountersMeasurements used to fill measurements with nil's for all non-excluded interfaces
func (nw *netWatcher) fillCountersMeasurements(results MeasurementsMap, interfaces []utilnet.InterfaceStat, excludedInterfacesByName map[string]struct{}) error {
	ctx, _ := context.WithTimeout(context.Background(), netGetCountersTimeout)
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
		log.Debugf("[NET] IO stat is available starting from 2nd check")
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
			log.Errorf("[NET] Previous IOCounters stat not found: %s", ioCounter.Name)
			continue
		}

		secondsSinceLastMeasurement := gotIOCountersAt.Sub(*nw.lastIOCountersAt).Seconds()
		for _, metric := range nw.cagent.NetMetrics {
			switch metric {
			case "in_B_per_s":
				bytesReceivedSinceLastMeasurement := ioCounter.BytesRecv - previousIOCounter.BytesRecv
				results[metric+"."+ioCounter.Name] = floatToIntRoundUP(float64(bytesReceivedSinceLastMeasurement) / secondsSinceLastMeasurement)
			case "out_B_per_s":
				bytesSentSinceLastMeasurement := ioCounter.BytesSent - previousIOCounter.BytesSent
				results[metric+"."+ioCounter.Name] = floatToIntRoundUP(float64(bytesSentSinceLastMeasurement) / secondsSinceLastMeasurement)
			case "errors_per_s":
				errorsSinceLastMeasurement := ioCounter.Errin + ioCounter.Errout - previousIOCounter.Errin - previousIOCounter.Errout
				results[metric+"."+ioCounter.Name] = floatToIntRoundUP(float64(errorsSinceLastMeasurement) / secondsSinceLastMeasurement)
			case "dropped_per_s":
				droppedSinceLastMeasurement := ioCounter.Dropin + ioCounter.Dropout - ioCounter.Dropin - previousIOCounter.Dropout
				results[metric+"."+ioCounter.Name] = floatToIntRoundUP(float64(droppedSinceLastMeasurement) / secondsSinceLastMeasurement)
			}
		}
	}

	return nil
}

func (nw *netWatcher) Results() (MeasurementsMap, error) {
	results := MeasurementsMap{}

	interfaces, err := utilnet.Interfaces()
	if err != nil {
		log.Errorf("[NET] Failed to read interfaces: %s", err.Error())
		return nil, err
	}

	excludedInterfacesByNameMap := nw.ExcludedInterfacesByName(interfaces)
	// fill counters measurements into results
	err = nw.fillCountersMeasurements(results, interfaces, excludedInterfacesByNameMap)
	if err != nil {
		log.Errorf("[NET] Failed to collect counters: %s", err.Error())
		return results, err
	}

	return results, nil
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
