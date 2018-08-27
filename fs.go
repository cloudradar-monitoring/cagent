package cagent

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
	"time"

	"github.com/shirou/gopsutil/disk"
	log "github.com/sirupsen/logrus"
)

const fsGetUsageTimeout = time.Second * 10
const fsGetPartitionsTimeout = time.Second * 10

type fsWatcher struct {
	AllowedTypes      map[string]struct{}
	ExcludedPathGlob  []string
	ExcludedPathCache map[string]bool
	Metrics           []string
}

func (ca *Cagent) FSWatcher() *fsWatcher {
	fsWatcher := fsWatcher{AllowedTypes: map[string]struct{}{}, ExcludedPathCache: map[string]bool{}}
	for _, t := range ca.FSTypeInclude {
		fsWatcher.AllowedTypes[strings.ToLower(t)] = struct{}{}
	}

	fsWatcher.ExcludedPathGlob = ca.FSPathExclude
	fsWatcher.Metrics = ca.FSMetrics

	return &fsWatcher
}

func (fw *fsWatcher) Results() (MeasurementsMap, error) {
	results := MeasurementsMap{}

	var errs []string
	ctx, _ := context.WithTimeout(context.Background(), fsGetPartitionsTimeout)
	partitions, err := disk.PartitionsWithContext(ctx, true)

	if err != nil {

		log.Errorf("[FS] Failed to read partitions: %s", err.Error())
		errs = append(errs, err.Error())
	}
	for _, partition := range partitions {
		if _, typeAllowed := fw.AllowedTypes[strings.ToLower(partition.Fstype)]; !typeAllowed {
			log.Debugf("[FS] fstype excluded: %s", partition.Fstype)

			continue
		}

		if pathExcluded, cacheExists := fw.ExcludedPathCache[strings.ToLower(partition.Mountpoint)]; cacheExists {
			if pathExcluded {
				log.Debugf("[FS] mountpoint excluded: %s", partition.Fstype)

				continue
			}
		} else {
			pathExcluded := false
			for _, glob := range fw.ExcludedPathGlob {
				pathExcluded, _ = filepath.Match(glob, partition.Mountpoint)
				if pathExcluded {
					break
				}
			}
			fw.ExcludedPathCache[strings.ToLower(partition.Mountpoint)] = pathExcluded

			if pathExcluded {
				log.Debugf("[FS] mountpoint excluded: %s", partition.Mountpoint)
				continue
			}
		}

		ctx, _ := context.WithTimeout(context.Background(), fsGetUsageTimeout)
		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)

		if err != nil {
			log.Errorf("[FS] Failed to read '%s'(%s): %s", partition.Mountpoint, partition.Device, err.Error())
			errs = append(errs, err.Error())
			for _, metric := range fw.Metrics {
				results[metric+"."+partition.Mountpoint] = nil
			}
			continue
		}

		for _, metric := range fw.Metrics {
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

	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errors.New("FS: " + strings.Join(errs, "; "))
}
