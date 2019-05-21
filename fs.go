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

const fsGetUsageTimeout = time.Second * 10
const fsGetPartitionsTimeout = time.Second * 10

type FSWatcher struct {
	AllowedTypes      map[string]struct{}
	ExcludePath       map[string]struct{}
	ExcludedPathCache map[string]bool
	cagent            *Cagent
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
	ctx, cancel := context.WithTimeout(context.Background(), fsGetPartitionsTimeout)
	defer cancel()

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

		cacheExists := false
		if pathExcluded, cacheExists = fw.ExcludedPathCache[partition.Mountpoint]; cacheExists {
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
			fw.ExcludedPathCache[strings.ToLower(partition.Mountpoint)] = pathExcluded

			if pathExcluded {
				log.Debugf("[FS] mountpoint excluded: %s", partition.Mountpoint)
				continue
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), fsGetUsageTimeout)
		// FIXME: inspect for possible resource leak
		defer cancel()

		usage, err := disk.UsageWithContext(ctx, partition.Mountpoint)

		if err != nil {
			log.Errorf("[FS] Failed to read '%s'(%s): %s", partition.Mountpoint, partition.Device, err.Error())
			errs = append(errs, err.Error())
			for _, metric := range fw.cagent.Config.FSMetrics {
				results[metric+"."+partition.Mountpoint] = nil
			}
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

	}

	if len(errs) != 0 {
		return results, errors.New("FS: " + strings.Join(errs, "; "))
	}

	return results, nil
}
