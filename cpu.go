package cagent

import (
	"context"
	"errors"
	"fmt"
	"math"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

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

type thresholdNotifier struct {
	Percentage           float64
	Metric               string // possible values: system, user, nice, idle, iowait, irq, softirq, steal
	Function             func(current, threshold float64) (notify bool)
	GatheringModeMinutes int // supported values: 1, 5, 15
	Chan                 chan float64
}

type CPUWatcher struct {
	LoadAvg1  bool
	LoadAvg5  bool
	LoadAvg15 bool

	UtilAvg   TimeSeriesAverage
	UtilTypes []string

	ThresholdNotifiers []thresholdNotifier
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
	if ca.cpuWatcher != nil {
		return ca.cpuWatcher
	}

	cw := CPUWatcher{}
	cw.UtilAvg.mu.Lock()

	if len(ca.Config.CPULoadDataGather) > 0 {
		_, err := load.Avg()

		if err != nil && err.Error() == "not implemented yet" {
			log.Errorf("[CPU] load_avg metric unavailable on %s", runtime.GOOS)
		} else {
			for _, d := range ca.Config.CPULoadDataGather {
				if strings.HasPrefix(d, "avg") {
					v, _ := strconv.Atoi(d[3:])

					switch v {
					case 1:
						cw.LoadAvg1 = true
					case 5:
						cw.LoadAvg5 = true
					case 15:
						cw.LoadAvg15 = true
					default:
						log.Errorf("[CPU] wrong cpu_load_data_gathering_mode. Supported values: avg1, avg5, avg15")
					}
				}
			}
		}
	}

	durations := []int{}
	for _, d := range ca.Config.CPUUtilDataGather {
		if strings.HasPrefix(d, "avg") {
			v, err := strconv.Atoi(d[3:])
			if err != nil {
				log.Errorf("[CPU] failed to parse cpu_load_data_gathering_mode '%s': %s", d, err.Error())
				continue
			}
			durations = append(durations, v)
		}
	}

	for _, t := range ca.Config.CPUUtilTypes {
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
			cw.UtilTypes = append(cw.UtilTypes, t)
		}
	}

	cw.UtilAvg.SetDurationsMinutes(durations...)
	cw.UtilAvg.mu.Unlock()
	ca.cpuWatcher = &cw

	// optimization to prevent CPU watcher to run in case CPU util metrics not are not needed
	if len(ca.Config.CPUUtilTypes) > 0 && len(ca.Config.CPUUtilDataGather) > 0 || len(ca.Config.CPULoadDataGather) > 0 {
		err := cw.Once()
		if err != nil {
			log.Error("[CPU] Failed to read utilisation metrics: " + err.Error())
		} else {
			go cw.Run()
		}
	}

	return ca.cpuWatcher
}

func (cw *CPUWatcher) Once() error {

	cw.UtilAvg.mu.Lock()

	ctx, cancel := context.WithTimeout(context.Background(), cpuGetUtilisationTimeout)
	defer cancel()
	times, err := cpu.TimesWithContext(ctx, true)

	if err != nil {
		cw.UtilAvg.mu.Unlock()
		return err
	}

	values := ValuesMap{}

	cpuStatPerCPUPerCorePerType := make(map[int]map[int]map[string]float64)

	for _, cputime := range times {
		for _, utype := range cw.UtilTypes {
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

				if _, exists := cpuStatPerCPUPerCorePerType[cpuIndex]; !exists {
					cpuStatPerCPUPerCorePerType[cpuIndex] = make(map[int]map[string]float64)
				}

				if _, exists := cpuStatPerCPUPerCorePerType[cpuIndex][coreIndex]; !exists {
					cpuStatPerCPUPerCorePerType[cpuIndex][coreIndex] = make(map[string]float64)
				}

				// store by indexes to iterate in the right order later
				cpuStatPerCPUPerCorePerType[cpuIndex][coreIndex][utype] = value
				values[fmt.Sprintf("%s.%%d.total", utype)] += value / float64(len(times))

			} else {
				values[fmt.Sprintf("%s.%%d.%s", utype, cputime.CPU)] = value
				values[fmt.Sprintf("%s.%%d.total", utype)] += value / float64(len(times))
			}
		}
	}

	if runtime.GOOS == "windows" {
		// calculate persistent CPU logical indexes from the "cpuIndex,coreIndex" on Windows
		// iterate on CPUs then on their cores
		// Result will be like this: 0,0 -> 0; 1,1 -> 3
		logicalCPUIndex := 0
		for cpuIndex := 0; cpuIndex < len(cpuStatPerCPUPerCorePerType); cpuIndex++ {
			for coreIndex := 0; coreIndex < len(cpuStatPerCPUPerCorePerType[cpuIndex]); coreIndex++ {
				for utype, value := range cpuStatPerCPUPerCorePerType[cpuIndex][coreIndex] {
					values[fmt.Sprintf("%s.%%d.cpu%d", utype, logicalCPUIndex)] = value
				}
				logicalCPUIndex++
			}
		}
	}

	cw.UtilAvg.Add(time.Now(), values)
	cw.UtilAvg.mu.Unlock()
	if cw.ThresholdNotifiers != nil {
		avg, _ := cw.UtilAvg.Percentage()

		for _, tm := range cw.ThresholdNotifiers {
			var values ValuesMap
			var exists bool
			if values, exists = avg[tm.GatheringModeMinutes]; !exists {
				continue
			}

			if val, exists := values[tm.Metric+".%d.total"]; exists && tm.Function(val, tm.Percentage) {
				tm.Chan <- val
			}
		}
	}
	return nil
}

func (cw *CPUWatcher) Run() {
	for {
		start := time.Now()
		err := cw.Once()
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

func (cw *CPUWatcher) Results() (MeasurementsMap, error) {
	var errs []string
	util, err := cw.UtilAvg.Percentage()
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
	if cw.LoadAvg1 || cw.LoadAvg5 || cw.LoadAvg15 {
		loadAvg, err = load.Avg()
		if err != nil {
			log.Error("[CPU] Failed to read load_avg: ", err.Error())
			errs = append(errs, err.Error())
		} else {
			if cw.LoadAvg1 {
				results["load.avg.1"] = loadAvg.Load1
			}

			if cw.LoadAvg5 {
				results["load.avg.5"] = loadAvg.Load5
			}

			if cw.LoadAvg15 {
				results["load.avg.15"] = loadAvg.Load15
			}
		}
	}

	if len(errs) == 0 {
		return results, nil
	}

	return results, errors.New("CPU: " + strings.Join(errs, "; "))

}

func (cw *CPUWatcher) AddThresholdNotifier(percentage float64, metric string, operator string, gatheringMode string, ch chan float64) error {

	if ch == nil {
		return fmt.Errorf("ch should be non-nil chan")
	}

	if percentage <= 0 || percentage > 100 {
		return fmt.Errorf("percentage should be more >0 and <=100")
	}
	tn := thresholdNotifier{Percentage: percentage, Chan: ch}

	tn.Percentage = percentage

	switch metric {
	case "system", "user", "nice", "idle", "iowait", "irq", "softirq", "steal":
		tn.Metric = metric
	default:
		return fmt.Errorf("wrong metric: should be one of: system, user, nice, idle, iowait, irq, softirq, steal")
	}

	switch operator {
	case "lt":
		tn.Function = func(current, threshold float64) bool {
			return current < threshold
		}
	case "lte":
		tn.Function = func(current, threshold float64) bool {
			return current <= threshold
		}
	case "gt":
		tn.Function = func(current, threshold float64) bool {
			return current > threshold
		}
	case "gte":
		tn.Function = func(current, threshold float64) bool {
			return current > threshold
		}
	default:
		return fmt.Errorf("wrong operator: should be one of: lt, lte, gt, gte")
	}

	switch gatheringMode {
	case "avg1":
		tn.GatheringModeMinutes = 1
	case "avg5":
		tn.GatheringModeMinutes = 5
	case "avg15":
		tn.GatheringModeMinutes = 15
	default:
		return fmt.Errorf("wrong gathering mode: should be one of: avg1, avg5, avg15")
	}

	cw.ThresholdNotifiers = append(cw.ThresholdNotifiers, tn)

	return nil
}
