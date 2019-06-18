package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/kardianos/service"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"

	"github.com/cloudradar-monitoring/cagent"
)

var (
	// set on build:
	// go build -o cagent -ldflags="-X main.version=$(git --git-dir=src/github.com/cloudradar-monitoring/cagent/.git describe --always --long --dirty --tag)" github.com/cloudradar-monitoring/cagent/cmd/cagent
	version string
)

var svcConfig = &service.Config{
	Name:        "cagent",
	DisplayName: "Cagent",
	Description: "Monitoring agent for system metrics",
}

func askForConfirmation(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/n]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("")
			log.WithError(err).Fatalln("Failed to read confirmation")
		}

		response = strings.ToLower(strings.TrimSpace(response))

		if response == "y" || response == "yes" {
			return true
		} else if response == "n" || response == "no" {
			return false
		}
	}
}

func main() {
	systemManager := service.ChosenSystem()

	var serviceInstallUserPtr *string
	var serviceInstallPtr *bool
	var settingsPtr *bool

	// Setup flag pointers
	outputFilePtr := flag.String("o", "", "file to write the results (default ./results.out)")
	cfgPathPtr := flag.String("c", cagent.DefaultCfgPath, "config file path")
	logLevelPtr := flag.String("v", "", "log level – overrides the level in config file (values \"error\",\"info\",\"debug\")")
	daemonizeModePtr := flag.Bool("d", false, "daemonize – run the process in background")
	oneRunOnlyModePtr := flag.Bool("r", false, "one run only – perform checks once and exit. Overwrites output file")
	serviceUninstallPtr := flag.Bool("u", false, fmt.Sprintf("stop and uninstall the system service(%s)", systemManager.String()))
	printConfigPtr := flag.Bool("p", false, "print the active config")
	testConfigPtr := flag.Bool("t", false, "test the HUB config")
	assumeYesPtr := flag.Bool("y", false, "automatic yes to prompts. Assume 'yes' as answer to all prompts and run non-interactively")
	flagServiceStatusPtr := flag.Bool("service_status", false, "check status of cagent within system service")
	flagServiceStartPtr := flag.Bool("service_start", false, "start cagent as system service")
	flagServiceStopPtr := flag.Bool("service_stop", false, "stop cagent if running as system service")
	flagServiceRestartPtr := flag.Bool("service_restart", false, "restart cagent within system service")
	flagServiceUpgradePtr := flag.Bool("service_upgrade", false, "upgrade cagent service unit configuration")

	if runtime.GOOS == "windows" {
		settingsPtr = flag.Bool("x", false, "open the settings UI")
	}

	versionPtr := flag.Bool("version", false, "show the cagent version")

	// some OS specific flags
	if runtime.GOOS == "windows" {
		serviceInstallPtr = flag.Bool("s", false, fmt.Sprintf("install and start the system service(%s)", systemManager.String()))
	} else {
		serviceInstallUserPtr = flag.String("s", "", fmt.Sprintf("username to install and start the system service(%s)", systemManager.String()))
	}

	flag.Parse()

	// version should be handled first to ensure it will be accessible in case of fatal errors before
	handleFlagVersion(*versionPtr)

	// check some incompatible flags
	if serviceInstallUserPtr != nil && *serviceInstallUserPtr != "" ||
		serviceInstallPtr != nil && *serviceInstallPtr {
		if *outputFilePtr != "" {
			log.Fatalln("Output file(-o) flag can't be used together with service install(-s) flag")
		}

		if *serviceUninstallPtr {
			log.Fatalln("Service uninstall(-u) flag can't be used together with service install(-s) flag")
		}
	}

	cfg, err := cagent.HandleAllConfigSetup(*cfgPathPtr)
	if err != nil {
		log.WithError(err).Fatalln("Failed to handle Cagent configuration")
	}

	ca := cagent.New(cfg, *cfgPathPtr, version)

	handleFlagPrintConfig(*printConfigPtr, cfg)

	if ((serviceInstallPtr == nil) || ((serviceInstallPtr != nil) && (!*serviceInstallPtr))) &&
		((serviceInstallUserPtr == nil) || ((serviceInstallUserPtr != nil) && len(*serviceInstallUserPtr) == 0)) &&
		!*serviceUninstallPtr {
		handleServiceCommand(ca, *flagServiceStatusPtr, *flagServiceStartPtr, *flagServiceStopPtr, *flagServiceRestartPtr)
	}

	handleFlagTest(*testConfigPtr, ca)
	handleFlagSettings(settingsPtr, ca)

	if len(*outputFilePtr) == 0 && cfg.IOMode == cagent.IOModeFile {
		*outputFilePtr = cfg.OutFile
	}

	// log level set in flag has a precedence. If specified we need to set it ASAP
	handleFlagLogLevel(ca, *logLevelPtr)

	writePidFileIfNeeded(ca, oneRunOnlyModePtr)
	defer removePidFileIfNeeded(ca, oneRunOnlyModePtr)

	handleToastFeedback(ca, *cfgPathPtr)

	if !service.Interactive() {
		runUnderOsServiceManager(ca)
	}

	handleFlagServiceUpgrade(ca, *cfgPathPtr, flagServiceUpgradePtr, serviceInstallUserPtr)
	handleFlagServiceUninstall(ca, *serviceUninstallPtr)
	handleFlagServiceInstall(ca, serviceInstallUserPtr, serviceInstallPtr, *cfgPathPtr, assumeYesPtr)
	handleFlagDaemonizeMode(*daemonizeModePtr)

	output := handleFlagOutput(*outputFilePtr, *oneRunOnlyModePtr)
	if output != nil {
		defer output.Close()
	}

	handleFlagOneRunOnlyMode(ca, *oneRunOnlyModePtr, output)

	// nothing resulted in os.Exit
	// so lets use the default continuous run mode and wait for interrupt
	// setup interrupt handler
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM)
	interruptChan := make(chan struct{})
	doneChan := make(chan struct{})
	go func() {
		ca.Run(output, interruptChan, cfg)
		doneChan <- struct{}{}
	}()

	//  Handle interrupts
	select {
	case sig := <-sigc:
		log.WithFields(log.Fields{
			"signal": sig.String(),
		}).Infoln("Finishing the batch and exit...")
		interruptChan <- struct{}{}
		os.Exit(0)
	case <-doneChan:
		os.Exit(0)
	}
}

func handleFlagVersion(versionFlag bool) {
	if versionFlag {
		fmt.Printf("cagent v%s released under MIT license. https://github.com/cloudradar-monitoring/cagent/\n", version)
		os.Exit(0)
	}
}

func handleServiceCommand(ca *cagent.Cagent, check, start, stop, restart bool) {
	if !check && !start && !stop && !restart {
		return
	}

	svc, err := getServiceFromFlags(ca, "", "")
	if err != nil {
		log.WithError(err).Fatalln("can't find service")
	}

	var status service.Status
	if status, err = svc.Status(); err != nil {
		log.WithError(err).Fatalln("can't get service status")
	}

	if check {
		switch status {
		case service.StatusRunning:
			fmt.Println("running")
		case service.StatusStopped:
			fmt.Println("stopped")
		case service.StatusUnknown:
			fmt.Println("unknown")
		}

		os.Exit(0)
	}

	if stop && (status == service.StatusRunning) {
		if err = svc.Stop(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("stopped")
		os.Exit(0)
	} else if stop {
		fmt.Println("service is not running")
		os.Exit(0)
	}

	if start {
		if status == service.StatusRunning {
			fmt.Println("already")
			os.Exit(1)
		}

		if err = svc.Start(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("started")
		os.Exit(0)
	}

	if restart {
		if err = svc.Restart(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		fmt.Println("restarted")
		os.Exit(0)
	}
}

func handleFlagPrintConfig(printConfig bool, cfg *cagent.Config) {
	if printConfig {
		fmt.Println(cfg.DumpToml())
		os.Exit(0)
	}
}

func handleFlagSettings(settingsUI *bool, ca *cagent.Cagent) {
	if settingsUI != nil && *settingsUI {
		windowsShowSettingsUI(ca, false)
		os.Exit(0)
	}
}

func handleFlagLogLevel(ca *cagent.Cagent, logLevel string) {
	// Check loglevel and if needed warn user and set to default
	switch cagent.LogLevel(logLevel) {
	case cagent.LogLevelError, cagent.LogLevelInfo, cagent.LogLevelDebug:
		ca.SetLogLevel(cagent.LogLevel(logLevel))
	default:
		if len(logLevel) > 0 {
			log.WithFields(log.Fields{
				"logLevel":     logLevel,
				"defaultLevel": ca.Config.LogLevel,
			}).Warnln("Unknown log level detected, setting to default")
		}
	}
}

func handleFlagOutput(outputFile string, oneRunOnlyMode bool) *os.File {
	if len(outputFile) == 0 {
		return nil
	}

	var output *os.File

	// forward output to stdout
	if outputFile == "-" {
		log.SetOutput(ioutil.Discard)
		output = os.Stdout
		return output
	}

	// if the output file does not exist, try to create it
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		dir := filepath.Dir(outputFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err = os.MkdirAll(dir, 0644)
			if err != nil {
				log.WithFields(log.Fields{
					"dir": dir,
				}).WithError(err).Fatalln("Failed to create the output file directory")
			}
		}
	}

	mode := os.O_WRONLY | os.O_CREATE

	if oneRunOnlyMode {
		mode = mode | os.O_TRUNC
	} else {
		mode = mode | os.O_APPEND
	}

	// Ensure that we can open the output file
	output, err := os.OpenFile(outputFile, mode, 0644)
	if err != nil {
		log.WithFields(log.Fields{
			"file": outputFile,
		}).WithError(err).Fatalln("Failed to open the output file")
	}

	return output
}

func handleFlagOneRunOnlyMode(ca *cagent.Cagent, oneRunOnlyMode bool, output *os.File) {
	if oneRunOnlyMode {
		err := ca.RunOnce(output)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}
}

func handleFlagDaemonizeMode(daemonizeMode bool) {
	if daemonizeMode && os.Getenv("cagent_FORK") != "1" {
		err := rerunDetached()
		if err != nil {
			log.WithError(err).Fatalln("Failed to fork process")
		}
		os.Exit(0)
	}
}

func handleFlagServiceUninstall(ca *cagent.Cagent, serviceUninstallPtr bool) {
	if !serviceUninstallPtr {
		return
	}

	log.Info("Uninstalling cagent service...")

	systemService, err := getServiceFromFlags(ca, "", "")
	if err != nil {
		log.WithError(err).Fatalln("Failed to get system service")
	}
	status, err := systemService.Status()
	if err != nil {
		log.WithError(err).Warnln("Failed to get service status")
	}
	if status == service.StatusRunning {
		err = systemService.Stop()
		if err != nil {
			// don't exit here, just write a warning and try to uninstall
			log.WithError(err).Warnln("Failed to stop the running service")
		}
	}
	err = systemService.Uninstall()
	if err != nil {
		log.WithError(err).Fatalln("Failed to uninstall the service")
	}

	log.Info("Uninstall successful")
	os.Exit(0)
}

func handleFlagServiceUpgrade(
	ca *cagent.Cagent,
	cfgPath string,
	serviceUpgradeFlag *bool,
	serviceInstallUserPtr *string,
) {
	if serviceUpgradeFlag == nil || !*serviceUpgradeFlag {
		return
	}

	installUser := ""
	if serviceInstallUserPtr != nil {
		installUser = *serviceInstallUserPtr
	}

	systemService, err := getServiceFromFlags(ca, cfgPath, installUser)
	if err != nil {
		log.WithError(err).Fatalln("Failed to get system service")
	}

	updateServiceConfig(ca, installUser)
	tryUpgradeServiceUnit(systemService)

	os.Exit(0)
}

func handleFlagServiceInstall(
	ca *cagent.Cagent,
	serviceInstallUserPtr *string,
	serviceInstallPtr *bool,
	cfgPath string,
	assumeYesPtr *bool,
) {
	// serviceInstallPtr is currently used on windows
	// serviceInstallUserPtr is used on other systems
	// if both of them are empty - just return
	if (serviceInstallUserPtr == nil || len(*serviceInstallUserPtr) == 0) &&
		(serviceInstallPtr == nil || !*serviceInstallPtr) {
		return
	}

	username := ""
	if serviceInstallUserPtr != nil {
		username = *serviceInstallUserPtr
	}

	s, err := getServiceFromFlags(ca, cfgPath, username)
	if err != nil {
		log.WithError(err).Fatalln("can't find service")
	}

	updateServiceConfig(ca, username)
	tryInstallService(s, assumeYesPtr)
	tryStartService(s)

	log.Infof("Log file located at: %s", ca.Config.LogFile)
	log.Infof("Config file located at: %s", cfgPath)
	os.Exit(0)
}

func runUnderOsServiceManager(ca *cagent.Cagent) {
	systemService, err := getServiceFromFlags(ca, "", "")
	if err != nil {
		log.WithError(err).Fatalln("Failed to get system service")
	}

	// we are running under OS service manager
	err = systemService.Run()
	if err != nil {
		log.WithError(err).Fatalln("Failed to run system service")
	}

	os.Exit(0)
}

func writePidFileIfNeeded(ca *cagent.Cagent, oneRunOnlyModePtr *bool) {
	if len(ca.Config.PidFile) > 0 && !*oneRunOnlyModePtr && runtime.GOOS != "windows" {
		err := ioutil.WriteFile(ca.Config.PidFile, []byte(strconv.Itoa(os.Getpid())), 0664)
		if err != nil {
			log.WithFields(log.Fields{
				"pidFile": ca.Config.PidFile,
			}).WithError(err).Errorf("Failed to write pid file")
		}
	}
}

func removePidFileIfNeeded(ca *cagent.Cagent, oneRunOnlyModePtr *bool) {
	if len(ca.Config.PidFile) > 0 && !*oneRunOnlyModePtr && runtime.GOOS != "windows" {
		if err := os.Remove(ca.Config.PidFile); err != nil {
			log.WithFields(log.Fields{
				"pidFile": ca.Config.PidFile,
			}).WithError(err).Errorf("Failed to remove pid file")
		}
	}
}

func handleFlagTest(testConfig bool, ca *cagent.Cagent) {
	if !testConfig {
		return
	}

	if ca.Config.IOMode == cagent.IOModeFile {
		localFields := log.Fields{
			"outFile": ca.Config.OutFile,
			"ioMode":  ca.Config.IOMode,
		}
		file, err := os.OpenFile(ca.Config.OutFile, os.O_WRONLY, 0666)
		if err != nil {
			log.WithFields(localFields).WithError(err).Fatalln("Failed to validate config")
		}
		if err := file.Close(); err != nil {
			log.WithFields(localFields).WithError(err).Fatalln("Could not close the config file")
		}
		log.WithFields(localFields).Infoln("Config verified")
		os.Exit(0)
	}

	ctx := context.Background()
	err := ca.CheckHubCredentials(ctx, "hub_url", "hub_user", "hub_password")
	if err != nil {
		if runtime.GOOS == "windows" {
			// ignore toast error to make the main error clear for user
			// toast error probably means toast not supported on the system
			_ = sendErrorNotification("Hub connection check failed", err.Error())
		}
		log.WithError(err).Errorln("Hub connection check failed")
		systemService, err := getServiceFromFlags(ca, "", "")
		if err != nil {
			log.WithError(err).Fatalln("Failed to get system service")
		}

		status, err := systemService.Status()
		if err != nil {
			// service seems not installed
			// no need to show the tip on how to restart it
			os.Exit(1)
		}

		systemManager := service.ChosenSystem()
		if status == service.StatusRunning || status == service.StatusStopped {
			restartCmdSpec := getSystemManagerCommand(systemManager.String(), svcConfig.Name, "restart")
			log.WithFields(log.Fields{
				"restartCmd": restartCmdSpec,
			}).Infoln("Fix the config and then restart the service")
		}

		os.Exit(1)
	}

	if runtime.GOOS == "windows" {
		_ = sendSuccessNotification("Hub connection check is done", "")
	}
	log.Infoln("Hub connection check is done and credentials are correct!")
	os.Exit(0)
}

func rerunDetached() error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	cmd := exec.Command(os.Args[0], os.Args[1:]...)
	cmd.Dir = cwd
	cmd.Env = append(os.Environ(), "CAGENT_FORK=1")

	err = cmd.Start()
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"PID": cmd.Process.Pid,
	}).Infoln("Cagent will continue in background...")

	cmd.Process.Release()
	return nil
}

type serviceWrapper struct {
	Cagent        *cagent.Cagent
	InterruptChan chan struct{}
	DoneChan      chan struct{}
}

func (sw *serviceWrapper) Start(s service.Service) error {
	sw.InterruptChan = make(chan struct{})
	sw.DoneChan = make(chan struct{})
	go func() {
		sw.Cagent.Run(nil, sw.InterruptChan, sw.Cagent.Config)
		sw.DoneChan <- struct{}{}
	}()

	return nil
}

func (sw *serviceWrapper) Stop(s service.Service) error {
	sw.InterruptChan <- struct{}{}
	log.Println("Finishing the batch and stop the service...")
	<-sw.DoneChan
	return nil
}

func getServiceFromFlags(ca *cagent.Cagent, configPath, userName string) (service.Service, error) {
	prg := &serviceWrapper{Cagent: ca}

	if configPath != "" {
		if !filepath.IsAbs(configPath) {
			var err error
			configPath, err = filepath.Abs(configPath)
			if err != nil {
				err = errors.Wrapf(err, "Failed to get absolute path to config at %s", configPath)
				return nil, err
			}
		}
		svcConfig.Arguments = []string{"-c", configPath}
	}

	if userName != "" {
		svcConfig.UserName = userName
	}

	return service.New(prg, svcConfig)
}

func getSystemManagerCommand(manager string, service string, command string) string {
	switch manager {
	case "unix-systemv":
		return "sudo service " + service + " " + command
	case "linux-upstart":
		return "sudo initctl " + command + " " + service
	case "linux-systemd":
		return "sudo systemctl " + command + " " + service + ".service"
	case "darwin-launchd":
		switch command {
		case "stop":
			command = "unload"
		case "start":
			command = "load"
		case "restart":
			return "sudo launchctl unload " + service + " && sudo launchctl load " + service
		}
		return "sudo launchctl " + command + " " + service
	case "windows-service":
		return "sc " + command + " " + service
	default:
		return ""
	}
}
