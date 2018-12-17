package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	"github.com/cloudradar-monitoring/cagent"
	"github.com/kardianos/service"
	log "github.com/sirupsen/logrus"
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
			log.Fatal("Failed to read confirmation", err.Error())
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

	// Setup flag pointers
	outputFilePtr := flag.String("o", "", "file to write the results (default ./results.out)")
	cfgPathPtr := flag.String("c", cagent.DefaultCfgPath, "config file path")
	logLevelPtr := flag.String("v", "", "log level – overrides the level in config file (values \"error\",\"info\",\"debug\")")
	daemonizeModePtr := flag.Bool("d", false, "daemonize – run the proccess in background")
	oneRunOnlyModePtr := flag.Bool("r", false, "one run only – perform checks once and exit. Overwrites output file")
	serviceUninstallPtr := flag.Bool("u", false, fmt.Sprintf("stop and uninstall the system service(%s)", systemManager.String()))
	printConfigPtr := flag.Bool("p", false, "print the active config")
	testConfigPtr := flag.Bool("t", false, "test the HUB config")
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
			fmt.Println("Output file(-o) flag can't be used together with service install(-s) flag")
			os.Exit(1)
		}

		if *serviceUninstallPtr {
			fmt.Println("Service uninstall(-u) flag can't be used together with service install(-s) flag")
			os.Exit(1)
		}
	}

	cfg, err := cagent.HandleAllConfigSetup(*cfgPathPtr)
	if err != nil {
		log.Fatalf("Failed to handle cagent configuration: %s", err)
	}

	ca := cagent.New(cfg, version)

	handleFlagPrintConfig(*printConfigPtr, cfg)

	setDefaultLogFormatter()

	// log level set in flag has a precedence. If specified we need to set it ASAP
	handleFlagLogLevel(ca, *logLevelPtr)

	writePidFileIfNeeded(ca, oneRunOnlyModePtr)
	defer removePidFileIfNeeded(ca, oneRunOnlyModePtr)

	handleToastFeedback(ca, *cfgPathPtr)
	handleFlagTest(*testConfigPtr, ca)

	if !service.Interactive() {
		runUnderOsServiceManager(ca)
	}

	handleFlagServiceUninstall(ca, *serviceUninstallPtr)
	handleFlagServiceInstall(ca, systemManager, serviceInstallUserPtr, serviceInstallPtr, *cfgPathPtr)
	handleFlagDaemonizeMode(*daemonizeModePtr)

	output := handleFlagOutput(*outputFilePtr, *oneRunOnlyModePtr)

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
		ca.Run(output, interruptChan)
		doneChan <- struct{}{}
	}()

	//  Handle interrupts
	select {
	case sig := <-sigc:
		log.Infof("Got %s signal. Finishing the batch and exit...", sig.String())
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

func handleFlagPrintConfig(printConfig bool, cfg *cagent.Config) {
	if printConfig {
		fmt.Println(cfg.DumpToml())
		os.Exit(0)
	}
}

func handleFlagLogLevel(ca *cagent.Cagent, logLevel string) {
	// Check loglevel and if needed warn user and set to default
	if logLevel == string(cagent.LogLevelError) || logLevel == string(cagent.LogLevelInfo) || logLevel == string(cagent.LogLevelDebug) {
		ca.SetLogLevel(cagent.LogLevel(logLevel))
	} else if logLevel != "" {
		log.Warnf("Invalid log level: \"%s\". Set to default: \"%s\"", logLevel, ca.Config.LogLevel)
	}
}

func handleFlagOutput(outputFile string, oneRunOnlyMode bool) *os.File {
	if outputFile == "" {
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
				log.WithError(err).Fatalf("Failed to create the output file directory: '%s'", dir)
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
		log.WithError(err).Fatalf("Failed to open the output file: '%s'", outputFile)
	}
	defer output.Close()

	return output
}

func handleFlagOneRunOnlyMode(ca *cagent.Cagent, oneRunOnlyMode bool, output *os.File) {
	if oneRunOnlyMode {
		err := ca.RunOnce(output)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}
}

func handleFlagDaemonizeMode(daemonizeMode bool) {
	if daemonizeMode && os.Getenv("cagent_FORK") != "1" {
		err := rerunDetached()
		if err != nil {
			fmt.Println("Failed to fork process: ", err.Error())
			os.Exit(1)
		}

		os.Exit(0)
	}
}

func handleFlagServiceUninstall(ca *cagent.Cagent, serviceUninstallPtr bool) {
	if !serviceUninstallPtr {
		return
	}

	systemService, err := getServiceFromFlags(ca, "", "")
	if err != nil {
		log.Fatalf("Failed to get system service", err.Error())
	}

	err = systemService.Stop()
	if err != nil {
		// don't return error here, just write a warning and try to uninstall
		fmt.Println("Failed to stop the service: ", err.Error())
	}

	err = systemService.Uninstall()
	if err != nil {
		fmt.Println("Failed to uninstall the service: ", err.Error())
		os.Exit(1)
	}

	os.Exit(0)
}

func handleFlagServiceInstall(ca *cagent.Cagent, systemManager service.System, serviceInstallUserPtr *string, serviceInstallPtr *bool, cfgPath string) {
	// serviceInstallPtr is currently used on windows
	// serviceInstallUserPtr is used on other systems
	// if both of them are empty - just return
	if (serviceInstallUserPtr == nil || *serviceInstallUserPtr == "") &&
		(serviceInstallPtr == nil || !*serviceInstallPtr) {
		return
	}

	username := ""
	if serviceInstallUserPtr != nil {
		username = *serviceInstallUserPtr
	}

	s, err := getServiceFromFlags(ca, cfgPath, username)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	if ca.Config.HubURL == "" {
		fmt.Printf("To install the service you first need to set 'hub_url' config param")
		os.Exit(1)
	}

	if runtime.GOOS != "windows" {
		userName := *serviceInstallUserPtr
		u, err := user.Lookup(userName)
		if err != nil {
			fmt.Printf("Failed to find the user '%s'\n", userName)
			os.Exit(1)
		}

		svcConfig.UserName = userName
		// we need to chown log file with user who will run service
		// because installer can be run under root so the log file will be also created under root
		err = chownFile(ca.Config.LogFile, u)
		if err != nil {
			fmt.Printf("Failed to chown log file for '%s' user\n", userName)
		}
	}
	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		err = s.Install()
		// Check error case where the service already exists
		if err != nil && strings.Contains(err.Error(), "already exists") {
			fmt.Printf("cagent service(%s) already installed: %s\n", systemManager.String(), err.Error())

			if attempt == maxAttempts {
				fmt.Printf("Give up after %d attempts\n", maxAttempts)
				os.Exit(1)
			}

			osSpecificNote := ""
			if runtime.GOOS == "windows" {
				osSpecificNote = " Windows Services Manager app should not be opened!"
			}
			if askForConfirmation("Do you want to overwrite it?" + osSpecificNote) {
				err = s.Stop()
				if err != nil {
					fmt.Println("Failed to stop the service: ", err.Error())
				}

				// lets try to uninstall despite of this error
				err := s.Uninstall()
				if err != nil {
					fmt.Println("Failed to unistall the service: ", err.Error())
					os.Exit(1)
				}
			}

			// Check general error case
		} else if err != nil {
			fmt.Printf("cagent service(%s) installing error: %s\n", systemManager.String(), err.Error())
			os.Exit(1)
			// Service install was success so we can exit the loop
		} else {
			break
		}
	}

	fmt.Printf("cagent service(%s) installed. Starting...\n", systemManager.String())
	err = s.Start()
	if err != nil {
		fmt.Println(err.Error())
	}

	switch systemManager.String() {
	case "unix-systemv":
		fmt.Printf("Run this command to stop it:\nsudo service %s stop\n\n", svcConfig.Name)
	case "linux-upstart":
		fmt.Printf("Run this command to stop it:\nsudo initctl stop %s\n\n", svcConfig.Name)
	case "linux-systemd":
		fmt.Printf("Run this command to stop it:\nsudo systemctl stop %s.service\n\n", svcConfig.Name)
	case "darwin-launchd":
		fmt.Printf("Run this command to stop it:\nsudo launchctl unload %s\n\n", svcConfig.Name)
	case "windows-service":
		fmt.Printf("Use the Windows Service Manager to stop it\n\n")
	}

	fmt.Printf("Log file located at: %s\n", cfgPath)
	os.Exit(0)
}

func runUnderOsServiceManager(ca *cagent.Cagent) {
	systemService, err := getServiceFromFlags(ca, "", "")
	if err != nil {
		log.Fatalf("Failed to get system service: %s", err.Error())
	}

	// we are running under OS service manager
	err = systemService.Run()
	if err != nil {
		log.Fatalf("Failed to run system service: %s", err.Error())
	}

	os.Exit(0)
}

func writePidFileIfNeeded(ca *cagent.Cagent, oneRunOnlyModePtr *bool) {
	if ca.Config.PidFile != "" && !*oneRunOnlyModePtr && runtime.GOOS != "windows" {
		err := ioutil.WriteFile(ca.Config.PidFile, []byte(strconv.Itoa(os.Getpid())), 0664)
		if err != nil {
			log.Errorf("Failed to write pid file at: %s", ca.Config.PidFile)
		}
	}
}

func removePidFileIfNeeded(ca *cagent.Cagent, oneRunOnlyModePtr *bool) {
	if ca.Config.PidFile != "" && !*oneRunOnlyModePtr && runtime.GOOS != "windows" {
		err := os.Remove(ca.Config.PidFile)
		if err != nil {
			log.Errorf("Failed to remove pid file at: %s", ca.Config.PidFile)
		}
	}
}

func handleFlagTest(testConfig bool, ca *cagent.Cagent) {
	if testConfig {

		err := ca.TestHub()
		if err != nil {
			if runtime.GOOS == "windows" {
				sendErrorNotification("Cagent connection test failed", err.Error())
			}
			fmt.Printf("Cagent HUB test failed: %s\n", err.Error())
			os.Exit(1)

		}

		if runtime.GOOS == "windows" {
			sendSuccessNotification("Cagent connection test succeed", "")
		}

		fmt.Printf("HUB connection test succeed and credentials are correct!\n")
		os.Exit(0)
	}
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

	fmt.Printf("Cagent will continue in background...\nPID %d", cmd.Process.Pid)

	cmd.Process.Release()
	return nil
}

func chownFile(filePath string, u *user.User) error {
	uid, err := strconv.Atoi(u.Uid)
	if err != nil {
		return fmt.Errorf("Chown files: error converting UID(%s) to int", u.Uid)
	}

	gid, err := strconv.Atoi(u.Gid)
	if err != nil {
		return fmt.Errorf("Chown files: error converting GID(%s) to int", u.Gid)
	}

	return os.Chown(filePath, uid, gid)
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
		sw.Cagent.Run(nil, sw.InterruptChan)
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
				return nil, fmt.Errorf("Failed to get absolute path to config at '%s': %s", configPath, err)
			}
		}
		svcConfig.Arguments = []string{"-c", configPath}
	}

	if userName != "" {
		svcConfig.UserName = userName
	}

	return service.New(prg, svcConfig)
}

func setDefaultLogFormatter() {
	tfmt := log.TextFormatter{FullTimestamp: true}
	if runtime.GOOS == "windows" {
		tfmt.DisableColors = true
	}

	log.SetFormatter(&tfmt)
}
