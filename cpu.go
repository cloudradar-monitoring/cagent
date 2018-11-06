package cagent

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"math"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	log "github.com/sirupsen/logrus"
)

const measureInterval = time.Second * 10
const cpuGetUtilisationTimeout = time.Second * 10

var errMetricsAreNotCollectedYet = errors.New("metrics are not collected yet")
var utilisationMetricsByOS = map[string][]string{
	"windows": {"system", "user", "idle", "irq"},
	"linux":   {"system", "user", "nice", "iowait", "idle", "softirq", "irq"},
	"freebsd": {"system", "user", "nice", "idle", "irq"},
	"solaris": {},
	"openbsd": {"system", "user", "nice", "idle", "irq"},
	"darwin":  {"system", "user", "nice", "idle"},
}

type ValuesMap map[string]float64
type ValuesCount map[string]int

type TimeValue struct {
	Time   time.Time
	Values ValuesMap
}

type TimeSeriesAverage struct {
	TimeSeries         []TimeValue
	mu                 sync.Mutex
	_DurationInMinutes []int // do not set directly, use SetDurationsMinutes
}

type CPUWatcher struct {
	LoadAvg1  bool
	LoadAvg5  bool
	LoadAvg15 bool

	UtilAvg   TimeSeriesAverage
	UtilTypes []string
}

var utilisationMetricsByOSMap = make(map[string]map[string]struct{})

func (tsa *TimeSeriesAverage) SetDurationsMinutes(durations ...int) {
	tsa._DurationInMinutes = durations
	sort.Ints(durations)
}

func init() {
	for osName, metrics := range utilisationMetricsByOS {
		utilisationMetricsByOSMap[osName] = make(map[string]struct{})
		for _, metric := range metrics {
			utilisationMetricsByOSMap[osName][metric] = struct{}{}
		}
	}
}

func minutes(mins int) time.Duration {
	return time.Duration(time.Minute * time.Duration(mins))
}

func (tsa *TimeSeriesAverage) Add(t time.Time, valuesMap ValuesMap) {
	for {
		// remove outdated measurements from the time series
		if len(tsa.TimeSeries) > 0 && time.Since(tsa.TimeSeries[0].Time) > minutes(tsa._DurationInMinutes[len(tsa._DurationInMinutes)-1]+1) {
			tsa.TimeSeries = tsa.TimeSeries[1:]
		} else {
			break
		}
	}
	tsa.TimeSeries = append(tsa.TimeSeries, TimeValue{t, valuesMap})
}

func (tsa *TimeSeriesAverage) Average() map[int]ValuesMap {
	sum := make(map[int]ValuesMap)
	count := make(map[int]ValuesCount)

	for _, d := range tsa._DurationInMinutes {
		sum[d] = make(ValuesMap)
		count[d] = make(ValuesCount)
	}
	for _, ts := range tsa.TimeSeries {
		n := time.Now()

		for _, d := range tsa._DurationInMinutes {
			if n.Sub(ts.Time) < minutes(d) {
				for key, val := range ts.Values {
					sum[d][key] += val
					count[d][key]++
				}
			}
		}
	}

	for _, d := range tsa._DurationInMinutes {
		for key, val := range sum[d] {
			sum[d][key] = val / float64(count[d][key])
		}
	}

	return sum
}

func roundUpWithPrecision(p float64, precision int) float64 {
	k := math.Pow10(precision)
	return float64(int64(p*k+0.5)) / k
}

func (tsa *TimeSeriesAverage) Percentage() (map[int]ValuesMap, error) {
	sum := make(map[int]ValuesMap)

	tsa.mu.Lock()
	defer tsa.mu.Unlock()

	if len(tsa.TimeSeries) == 0 {
		return nil, errMetricsAreNotCollectedYet
	}

	last := tsa.TimeSeries[len(tsa.TimeSeries)-1]
	for _, d := range tsa._DurationInMinutes {
		sum[d] = make(ValuesMap)
		// found minimal index of the first measurement in this period
		keyInt := len(tsa.TimeSeries) - int(int64(d)*int64(time.Minute)/int64(measureInterval)) - 1

		if keyInt < 0 {
			log.Debugf("cpu.util metrics for %d min avg calculation are not collected yet", d)
		}

		for key, lastVal := range last.Values {
			if keyInt < 0 {
				sum[d][key] = -1
				continue
			}
			// On windows perCPU measurements returned as Percentage, not CPU time
			// see https://github.com/shirou/gopsutil/issues/600
			if runtime.GOOS == "windows" {
				var valSum float64
				var count int
				for i := keyInt; i < len(tsa.TimeSeries); i++ {
					t := tsa.TimeSeries[i]
					// skip measurements collected more than d minutes ago (e.g. 5 minutes for avg5)
					if time.Since(t.Time).Minutes() > float64(d) {
						continue
					}

					valSum += t.Values[key]
					count++
				}

				// found the average for all collected measurements for the last d minutes
				sum[d][key] = roundUpWithPrecision(valSum/float64(count), 2)
			} else {

				var hasMetrics bool
				// filter out measurements collected more than d minutes ago (e.g. in case some of them were timeouted)
				for i := keyInt; i < len(tsa.TimeSeries); i++ {
					// allow 3 seconds(0.05min) outage to include metrics query time
					if time.Since(tsa.TimeSeries[i].Time).Minutes() <= float64(d)+0.05 {
						keyInt = i
						hasMetrics = true
						break
					}
				}
				if !hasMetrics || keyInt == len(tsa.TimeSeries)-1 {
					// looks like some problem happen and we don't have enough(more than 1) measurements
					// this could happen if all CPU queries for the last d minutes were timeouted
					sum[d][key] = -1
					continue
				}

				secondsSpentOnThisTypeOfLoad := lastVal - tsa.TimeSeries[keyInt].Values[key]
				secondsBetweenFirstAndLastMeasurementInTheRange := last.Time.Sub(tsa.TimeSeries[keyInt].Time).Seconds()

				// divide CPU times with seconds to found the percentage
				sum[d][key] = roundUpWithPrecision((secondsSpentOnThisTypeOfLoad/secondsBetweenFirstAndLastMeasurementInTheRange)*100, 2)
			}
		}
	}

	return sum, nil
}

func (ca *Cagent) CPUWatcher() *CPUWatcher {
	stat := CPUWatcher{}
	stat.UtilAvg.mu.Lock()

	if len(ca.CPULoadDataGather) > 0 {
		_, err := load.Avg()

		if err != nil && err.Error() == "not implemented yet" {
			log.Errorf("[CPU] load_avg metric unavailable on %s", runtime.GOOS)
		} else {
			for _, d := range ca.CPULoadDataGather {
				if strings.HasPrefix(d, "avg") {
					v, _ := strconv.Atoi(d[3:])

					switch v {
					case 1:
						stat.LoadAvg1 = true
					case 5:
						stat.LoadAvg5 = true
					case 15:
						stat.LoadAvg15 = true
					default:
						log.Errorf("[CPU] wrong cpu_load_data_gathering_mode. Supported values: avg1, avg5, avg15")
					}
				}
			}
		}
	}

	durations := []int{}
	for _, d := range ca.CPUUtilDataGather {
		if strings.HasPrefix(d, "avg") {
			v, err := strconv.Atoi(d[3:])
			if err != nil {
				log.Errorf("[CPU] failed to parse cpu_load_data_gathering_mode '%s': %s", d, err.Error())
				continue
			}
			durations = append(durations, v)
		}
	}

	for _, t := range ca.CPUUtilTypes {
		found := false

		for _, metric := range utilisationMetricsByOS[runtime.GOOS] {
			if metric == t {
				found = true
				break
			}
		}

		if !found {
			log.Errorf("[CPU] utilisation metric '%s' not implemented on %s", t, runtime.GOOS)
		} else {
			stat.UtilTypes = append(stat.UtilTypes, t)
		}
	}

	stat.UtilAvg.SetDurationsMinutes(durations...)
	stat.UtilAvg.mu.Unlock()

	return &stat
}

func (stat *CPUWatcher) Once() error {

	stat.UtilAvg.mu.Lock()
	defer stat.UtilAvg.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), cpuGetUtilisationTimeout)
	defer cancel()
	times, err := cpu.TimesWithContext(ctx, true)

	if err != nil {
		return err
	}

	values := ValuesMap{}

	cpuStatPerCpuPerCorePerType := make(map[int]map[int]map[string]float64)

	for _, cputime := range times {
		for _, utype := range stat.UtilTypes {
			utype = strings.ToLower(utype)
			var value float64
			switch utype {
			case "system":
				value = cputime.System
			case "user":
				value = cputime.User
			case "nice":
				value = cputime.Nice
			case "idle":
				value = cputime.Idle
			case "iowait":
				value = cputime.Iowait
			case "irq":
				value = cputime.Irq
			case "softirq":
				value = cputime.Softirq
			case "steal":
				value = cputime.Steal
			default:
				continue
			}

			if runtime.GOOS == "windows" {
				// calculate the logical CPU index from "cpuIndex,coreIndex" format
				cpuIndexParts := strings.Split(cputime.CPU, ",")

				cpuIndex, _ := strconv.Atoi(cpuIndexParts[0])
				coreIndex, _ := strconv.Atoi(cpuIndexParts[1])

				if _, exists := cpuStatPerCpuPerCorePerType[cpuIndex]; !exists {
					cpuStatPerCpuPerCorePerType[cpuIndex] = make(map[int]map[string]float64)
				}

				if _, exists := cpuStatPerCpuPerCorePerType[cpuIndex][coreIndex]; !exists {
					cpuStatPerCpuPerCorePerType[cpuIndex][coreIndex] = make(map[string]float64)
				}

				// store by indexes to iterate in the right order later
				cpuStatPerCpuPerCorePerType[cpuIndex][coreIndex][utype] = value
				values[fmt.Sprintf("%s.%%d.total", utype)] += value / float64(len(times))

			} else {
				values[fmt.Sprintf("%s.%%d.%s", utype, cputime.CPU)] = value
				values[fmt.Sprintf("%s.%%d.total", utype)] += value / float64(len(times))
			}
		}
	}

	if runtime.GOOS == "windows" {
		// calculate CPU logical index on Windows
		logicalCPUIndex := 0
		for cpuIndex := 0; cpuIndex < len(cpuStatPerCpuPerCorePerType); cpuIndex++ {
			for coreIndex := 0; coreIndex < len(cpuStatPerCpuPerCorePerType[cpuIndex]); coreIndex++ {
				for utype, value := range cpuStatPerCpuPerCorePerType[cpuIndex][coreIndex] {
					values[fmt.Sprintf("%s.%%d.cpu%d", utype, logicalCPUIndex)] = value
				}
				logicalCPUIndex++
			}
		}
	}

	stat.UtilAvg.Add(time.Now(), values)
	return nil
}

func (stat *CPUWatcher) Run() {
	for {
		start := time.Now()
		err := stat.Once()
		if err != nil {
			log.Errorf("[CPU] Failed to read utilisation metrics: " + err.Error())
		}

		spent := time.Since(start)

		// Sleep if we spent less than measureInterval on measurement
		if spent < measureInterval {
			time.Sleep(measureInterval - spent)
		}
	}
}

func (cs *CPUWatcher) Results() (MeasurementsMap, error) {
	var errs []string
	util, err := cs.UtilAvg.Percentage()
	if err != nil {
		log.Errorf("[CPU] Failed to calculate utilisation metrics: " + err.Error())
		errs = append(errs, err.Error())
	}
	results := MeasurementsMap{}
	for d, m := range util {
		for k, v := range m {
			if v == -1 {
				results["util."+fmt.Sprintf(k, d)] = nil
			} else {
				results["util."+fmt.Sprintf(k, d)] = v
			}
		}
	}
	var loadAvg *load.AvgStat
	if cs.LoadAvg1 || cs.LoadAvg5 || cs.LoadAvg15 {
		loadAvg, err = load.Avg()
		if err != nil {
			log.Error("[CPU] Failed to read load_avg: ", err.Error())
			errs = append(errs, err.Error())
		} else {
			if cs.LoadAvg1 {
				results["load.avg.1"] = loadAvg.Load1
			}

			if cs.LoadAvg5 {
				results["load.avg.5"] = loadAvg.Load5
			}

			if cs.LoadAvg15 {
				results["load.avg.15"] = loadAvg.Load15
			}
		}
	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errors.New("CPU: " + strings.Join(errs, "; "))

}
