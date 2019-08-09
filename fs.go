package cagent

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

const fsInfoRequestTimeout = time.Second * 10

type ioCountersMeasurement struct {
	timestamp time.Time
	counters  *disk.IOCountersStat
}

type ioUsageInfo struct {
	readBytesPerSecond       float64
	writeBytesPerSecond      float64
	readOperationsPerSecond  float64
	writeOperationsPerSecond float64
}

type FSWatcher struct {
	AllowedTypes      map[string]struct{}
	ExcludePath       map[string]struct{}
	ExcludedPathCache map[string]bool
	cagent            *Cagent
	prevIOCounters    map[string]*ioCountersMeasurement
}

func (ca *Cagent) FSWatcher() *FSWatcher {
	if ca.fsWatcher != nil {
		return ca.fsWatcher
	}

	ca.fsWatcher = &FSWatcher{
		AllowedTypes:      map[string]struct{}{},
		ExcludePath:       make(map[string]struct{}),
		ExcludedPathCache: map[string]bool{},
		cagent:            ca,
		prevIOCounters:    make(map[string]*ioCountersMeasurement),
	}

	for _, t := range ca.Config.FSTypeInclude {
		ca.fsWatcher.AllowedTypes[strings.ToLower(t)] = struct{}{}
	}

	for _, t := range ca.Config.FSPathExclude {
		ca.fsWatcher.ExcludePath[t] = struct{}{}
	}

	return ca.fsWatcher
}

func (fw *FSWatcher) Results() (common.MeasurementsMap, error) {
	results := common.MeasurementsMap{}

	var errs []string
	ctx, cancel := context.WithTimeout(context.Background(), fsInfoRequestTimeout)
	defer cancel()

	partitions, err := disk.PartitionsWithContext(ctx, true)

	if err != nil {
		log.WithError(err).Errorf("[FS] Failed to read partitions")
		errs = append(errs, err.Error())
	}

	partitionIOUsage := map[string]*ioUsageInfo{}
	for _, partition := range partitions {
		if _, typeAllowed := fw.AllowedTypes[strings.ToLower(partition.Fstype)]; !typeAllowed {
			log.Debugf("[FS] fstype excluded: %s", partition.Fstype)
			continue
		}

		pathExcluded := false

		if fw.cagent.Config.FSPathExcludeRecurse {
			for path := range fw.ExcludePath {
				if strings.HasPrefix(partition.Mountpoint, path) {
					log.Debugf("[FS] mountpoint excluded: %s", partition.Mountpoint)
					pathExcluded = true
					break
				}
			}
		}

		if pathExcluded {
			continue
		}

		partitionMountPoint := strings.ToLower(partition.Mountpoint)

		cacheExists := false
		if pathExcluded, cacheExists = fw.ExcludedPathCache[partitionMountPoint]; cacheExists {
			if pathExcluded {
				log.Debugf("[FS] mountpoint excluded: %s", partition.Fstype)
				continue
			}
		} else {
			pathExcluded = false
			for _, glob := range fw.cagent.Config.FSPathExclude {
				pathExcluded, _ = filepath.Match(glob, partition.Mountpoint)
				if pathExcluded {
					break
				}
			}
			fw.ExcludedPathCache[partitionMountPoint] = pathExcluded

			if pathExcluded {
				log.Debugf("[FS] mountpoint excluded: %s", partition.Mountpoint)
				continue
			}
		}

		usage, err := getFsPartitionUsageInfo(partition.Mountpoint)
		if err != nil {
			log.WithError(err).Errorf("[FS] Failed to get usage info for '%s'(%s)", partition.Mountpoint, partition.Device)
			errs = append(errs, err.Error())
			continue
		}

		for _, metric := range fw.cagent.Config.FSMetrics {
			switch strings.ToLower(metric) {
			case "free_b":
				results[metric+"."+partition.Mountpoint] = float64(usage.Free)
			case "free_percent":
				results[metric+"."+partition.Mountpoint] = float64(int64((100-usage.UsedPercent)*100+0.5)) / 100
			case "used_percent":
				results[metric+"."+partition.Mountpoint] = float64(int64(usage.UsedPercent*100+0.5)) / 100
			case "total_b":
				results[metric+"."+partition.Mountpoint] = usage.Total
			case "inodes_total":
				results[metric+"."+partition.Mountpoint] = usage.InodesTotal
			case "inodes_free":
				results[metric+"."+partition.Mountpoint] = usage.InodesFree
			case "inodes_used":
				results[metric+"."+partition.Mountpoint] = usage.InodesUsed
			case "inodes_used_percent":
				results[metric+"."+partition.Mountpoint] = float64(int64(usage.InodesUsedPercent*100+0.5)) / 100
			}
		}

		ioCounters, err := getPartitionIOCounters(partition.Device)
		if err != nil {
			log.WithError(err).Errorf("[FS] Failed to get IO counters for '%s'(%s)", partition.Mountpoint, partition.Device)
			errs = append(errs, err.Error())
			continue
		}
		currTimestamp := time.Now()
		var prevIOCountersMeasurementTimestamp time.Time
		var prevIOCounters *disk.IOCountersStat
		prevIOCountersMeasurement, prevMeasurementExists := fw.prevIOCounters[partitionMountPoint]
		fw.prevIOCounters[partitionMountPoint] = &ioCountersMeasurement{currTimestamp, ioCounters}
		if prevMeasurementExists {
			prevIOCountersMeasurementTimestamp = prevIOCountersMeasurement.timestamp
			prevIOCounters = prevIOCountersMeasurement.counters

			ioUsage := calcIOCountersUsage(prevIOCounters, ioCounters, currTimestamp.Sub(prevIOCountersMeasurementTimestamp))
			for _, metric := range fw.cagent.Config.FSMetrics {
				switch strings.ToLower(metric) {
				case "read_b_per_s":
					results[metric+"."+partition.Mountpoint] = common.RoundToTwoDecimalPlaces(ioUsage.readBytesPerSecond)
				case "write_b_per_s":
					results[metric+"."+partition.Mountpoint] = common.RoundToTwoDecimalPlaces(ioUsage.writeBytesPerSecond)
				case "read_ops_per_s":
					results[metric+"."+partition.Mountpoint] = common.RoundToTwoDecimalPlaces(ioUsage.readOperationsPerSecond)
				case "write_ops_per_s":
					results[metric+"."+partition.Mountpoint] = common.RoundToTwoDecimalPlaces(ioUsage.writeOperationsPerSecond)
				}
			}
			partitionIOUsage[partitionMountPoint] = ioUsage
		} else {
			log.Debugf("[FS] skipping IO usage metrics for %s as it will be available starting from second check", partition.Mountpoint)
		}
	}

	totalIOUsage := calcTotalIOUsage(partitionIOUsage)
	for _, metric := range fw.cagent.Config.FSMetrics {
		switch strings.ToLower(metric) {
		case "read_b_per_s":
			results["total_read_B_per_s"] = common.RoundToTwoDecimalPlaces(totalIOUsage.readBytesPerSecond)
		case "write_b_per_s":
			results["total_write_B_per_s"] = common.RoundToTwoDecimalPlaces(totalIOUsage.writeBytesPerSecond)
		case "read_ops_per_s":
			results["total_read_ops_per_s"] = common.RoundToTwoDecimalPlaces(totalIOUsage.readOperationsPerSecond)
		case "write_ops_per_s":
			results["total_write_ops_per_s"] = common.RoundToTwoDecimalPlaces(totalIOUsage.writeOperationsPerSecond)
		}
	}

	if len(errs) != 0 {
		return results, errors.New("FS: " + strings.Join(errs, "; "))
	}

	return results, nil
}

func getFsPartitionUsageInfo(mountPoint string) (*disk.UsageStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fsInfoRequestTimeout)
	defer cancel()
	return disk.UsageWithContext(ctx, mountPoint)
}

func getPartitionIOCounters(deviceName string) (*disk.IOCountersStat, error) {
	ctx, cancel := context.WithTimeout(context.Background(), fsInfoRequestTimeout)
	defer cancel()
	name := filepath.Base(deviceName)
	result, err := disk.IOCountersWithContext(ctx, name)
	if err != nil {
		return nil, err
	}
	ret := result[name]
	return &ret, nil
}

func calcIOCountersUsage(prev, curr *disk.IOCountersStat, timeDelta time.Duration) *ioUsageInfo {
	deltaSeconds := timeDelta.Seconds()
	return &ioUsageInfo{
		readBytesPerSecond:       float64(curr.ReadBytes-prev.ReadBytes) / deltaSeconds,
		writeBytesPerSecond:      float64(curr.WriteBytes-prev.WriteBytes) / deltaSeconds,
		readOperationsPerSecond:  float64(curr.ReadCount-prev.ReadCount) / deltaSeconds,
		writeOperationsPerSecond: float64(curr.WriteCount-prev.WriteCount) / deltaSeconds,
	}
}

func calcTotalIOUsage(partitionsUsageInfo map[string]*ioUsageInfo) *ioUsageInfo {
	result := &ioUsageInfo{}
	for _, info := range partitionsUsageInfo {
		result.readBytesPerSecond += info.readBytesPerSecond
		result.writeBytesPerSecond += info.writeBytesPerSecond
		result.readOperationsPerSecond += info.readOperationsPerSecond
		result.writeOperationsPerSecond += info.writeOperationsPerSecond
	}
	return result
}
