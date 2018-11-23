package main

var opts struct {
	// Slice of bool will append 'true' each time the option
	// is encountered (can be set multiple times, like -vvv)
	OutputFile         string `short:"o" long:"output" description:"file to write the results"`
	ConfigPath         string `short:"c" long:"config" description:"config file path"`
	Verbose            string `short:"v" long:"verbose" default:"error" description:"log level – overrides the level in config file (values \"error\",\"info\",\"debug\")"`
	Daemonize          bool   `short:"d" long:"daemonize" description:"daemonize – run the process in background"`
	RunOnly            bool   `short:"r" long:"run_only" description:"one run only – perform checks once and exit. Overwrites output file"`
	DumpConfig         bool   `short:"p" long:"print_config" description:"print the active config"`
	TestConfig         bool   `short:"t" description:"test the HUB config"`
	Version            bool   `long:"version" description:"show the cagent version"`
	ServiceInstall     *bool
	ServiceInstallUser *string
	ServiceUninstall   bool
}
