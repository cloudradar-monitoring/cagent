package cagent

import (
	"os"
	"os/signal"
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
		return ca.cpuUtilisationAnalyser
	}

	go func() {
		for {
			select {
			case x := <-thresholdChan:
				log.Debugf("[CPU_ANALYSIS] CPU threshold signal(%.2f) received from chan", x)
				if !cuan.topIsRunning {
					go cuan.top.Run()
					cuan.topIsRunning = true
				}
				break
			case <-time.After(time.Duration(cfg.TrailingProcessAnalysisMinutes) * time.Minute):
				if cuan.topIsRunning {
					log.Debugf("[CPU_ANALYSIS] TrailingRecoveryTime reached")
					cuan.hasUnclaimedResults = true
					cuan.topIsRunning = false
					cuan.top.Stop()
				}
				break
			case <-sigc:
				log.Debugf("[CPU_ANALYSIS] got interrupt signal")
				return
			}
		}
	}()

	return ca.cpuUtilisationAnalyser
}

func (cuan *CPUUtilisationAnalyser) Results() (MeasurementsMap, bool, error) {
	if cuan.top == nil || !cuan.hasUnclaimedResults && !cuan.topIsRunning {
		return nil, false, nil
	}

	cuan.hasUnclaimedResults = false
	topProcs := cuan.top.HighestNLoad(cuan.NumberOfProcesses)

	return MeasurementsMap{"top": topProcs}, true, nil
}
