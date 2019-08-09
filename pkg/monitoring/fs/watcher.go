package fs

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/shirou/gopsutil/disk"
	"github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/common"
)

type FileSystemWatcherConfig struct {
	TypeInclude        []string
	PathExclude        []string
	PathExcludeRecurse bool
	Metrics            []string
}

type FileSystemWatcher struct {
	AllowedTypes      map[string]struct{}
	ExcludePath       map[string]struct{}
	ExcludedPathCache map[string]bool
	config            *FileSystemWatcherConfig
	prevIOCounters    map[string]*ioCountersMeasurement
}

func NewWatcher(config FileSystemWatcherConfig) *FileSystemWatcher {
	fsWatcher := &FileSystemWatcher{
		AllowedTypes:      map[string]struct{}{},
		ExcludePath:       make(map[string]struct{}),
		ExcludedPathCache: map[string]bool{},
		config:            &config,
		prevIOCounters:    make(map[string]*ioCountersMeasurement),
	}

	for _, t := range config.TypeInclude {
		fsWatcher.AllowedTypes[strings.ToLower(t)] = struct{}{}
	}

	for _, t := range config.PathExclude {
		fsWatcher.ExcludePath[t] = struct{}{}
	}

	return fsWatcher
}

func (fw *FileSystemWatcher) Results() (common.MeasurementsMap, error) {
	results := common.MeasurementsMap{}

	var errs []string
	ctx, cancel := context.WithTimeout(context.Background(), fsInfoRequestTimeout)
	defer cancel()

	partitions, err := disk.PartitionsWithContext(ctx, true)

	if err != nil {
		logrus.WithError(err).Errorf("[FS] Failed to read partitions")
		errs = append(errs, err.Error())
	}

	partitionIOUsage := map[string]*ioUsageInfo{}
	for _, partition := range partitions {
		if _, typeAllowed := fw.AllowedTypes[strings.ToLower(partition.Fstype)]; !typeAllowed {
			logrus.Debugf("[FS] fstype excluded: %s", partition.Fstype)
			continue
		}

		pathExcluded := false

		if fw.config.PathExcludeRecurse {
			for path := range fw.ExcludePath {
				if strings.HasPrefix(partition.Mountpoint, path) {
					logrus.Debugf("[FS] mountpoint excluded: %s", partition.Mountpoint)
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
				logrus.Debugf("[FS] mountpoint excluded: %s", partition.Fstype)
				continue
			}
		} else {
			pathExcluded = false
			for _, glob := range fw.config.PathExclude {
				pathExcluded, _ = filepath.Match(glob, partition.Mountpoint)
				if pathExcluded {
					break
				}
			}
			fw.ExcludedPathCache[partitionMountPoint] = pathExcluded

			if pathExcluded {
				logrus.Debugf("[FS] mountpoint excluded: %s", partition.Mountpoint)
				continue
			}
		}

		usage, err := getFsPartitionUsageInfo(partition.Mountpoint)
		if err != nil {
			logrus.WithError(err).Errorf("[FS] Failed to get usage info for '%s'(%s)", partition.Mountpoint, partition.Device)
			errs = append(errs, err.Error())
			continue
		}

		for _, metric := range fw.config.Metrics {
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
			logrus.WithError(err).Errorf("[FS] Failed to get IO counters for '%s' (device %s)", partition.Mountpoint, partition.Device)
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
			for _, metric := range fw.config.Metrics {
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
			logrus.Debugf("[FS] skipping IO usage metrics for %s as it will be available starting from second check", partition.Mountpoint)
		}
	}

	totalIOUsage := calcTotalIOUsage(partitionIOUsage)
	for _, metric := range fw.config.Metrics {
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
