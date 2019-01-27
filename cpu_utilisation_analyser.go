package cagent

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent/pkg/monitoring/top"
)

type CPUUtilisationAnalyser struct {
	NumberOfProcesses int

	top                 *top.Top
	topIsRunning        bool
	hasUnclaimedResults bool
}

func (ca *Cagent) CPUUtilisationAnalyser() *CPUUtilisationAnalyser {

	if ca.cpuUtilisationAnalyser != nil {
		return ca.cpuUtilisationAnalyser
	}

	cfg := ca.Config.CPUUtilisationAnalysis
	if cfg.Threshold < 0 || cfg.Metric == "" || cfg.Function == "" || cfg.GatheringMode == "" || cfg.ReportProcesses == 0 {
		return &CPUUtilisationAnalyser{}
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)

	cuan := CPUUtilisationAnalyser{NumberOfProcesses: cfg.ReportProcesses}
	ca.cpuUtilisationAnalyser = &cuan
	cuan.top = top.New()

	thresholdChan := make(chan float64)
	err := ca.cpuWatcher.AddThresholdNotifier(cfg.Threshold, cfg.Metric, cfg.Function, cfg.GatheringMode, thresholdChan)
	if err != nil {
		log.Error("[CPU_ANALYSIS] addThresholdNotifier error", err.Error())
	} else {
		go func() {
			for {
				select {
				case <-thresholdChan:
					log.Debugf("[CPU_ANALYTICS] CPU threshold signal received from chan")
					if !cuan.topIsRunning {
						go cuan.top.Run()
						cuan.topIsRunning = true
					}
					break
				case <-time.After(time.Duration(cfg.TrailingRecoveryMinutes) * time.Minute):
					if cuan.topIsRunning {
						log.Debugf("[CPU_ANALYTICS] TrailingRecoveryTime reached")
						cuan.hasUnclaimedResults = true
						cuan.topIsRunning = false
						cuan.top.Stop()
					}
					break
				case <-sigc:
					log.Debugf("[CPU_ANALYTICS] got interrupt signal")
					return
				}
			}
		}()
	}
	return ca.cpuUtilisationAnalyser

}

func (cuan *CPUUtilisationAnalyser) Results() (MeasurementsMap, error) {
	if cuan.top == nil || !cuan.hasUnclaimedResults && !cuan.topIsRunning {
		return nil, nil
	}

	cuan.hasUnclaimedResults = false
	topProcs, err := cuan.top.GetProcesses()
	if err != nil {
		log.Errorf("[CPU_ANALYSIS] top.GetProcesses() error: %s", err.Error())
		return nil, fmt.Errorf("CPU TOP PROCS: %s", err.Error())
	}

	sort.Slice(topProcs, func(i, j int) bool {
		return topProcs[i].Load > topProcs[j].Load
	})

	if len(topProcs) > cuan.NumberOfProcesses {
		topProcs = topProcs[0:cuan.NumberOfProcesses]
	}

	return MeasurementsMap{"top": topProcs}, nil
}
