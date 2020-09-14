package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/cloudradar-monitoring/cagent"
	"github.com/cloudradar-monitoring/cagent/pkg/csender"
)

type boolFlag struct {
	set   bool
	value bool
}

func (sf *boolFlag) Set(value string) error {
	i, err := strconv.Atoi(value)
	if err != nil {
		return err
	}

	switch i {
	case 0:
		sf.value = false
	case 1:
		sf.value = true
	default:
		return fmt.Errorf("arg can be either 0 or 1")
	}

	sf.set = true
	return nil
}

func (sf *boolFlag) String() string {
	var s string
	if sf.set {
		if sf.value {
			s = "1"
		} else {
			s = "0"
		}
	}

	return s
}

func (sf *boolFlag) BoolPtr() *bool {
	if !sf.set {
		return nil
	}

	return &sf.value
}

func fatal(msg string) {
	_, _ = fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}

func main() {
	var successFlag boolFlag

	checkNamePtr := flag.String("n", "", "check name (*required)")
	tokenPtr := flag.String("t", "", "custom check token (*required)")
	hubURLPtr := flag.String("u", "https://hub.cloudradar.io/cct/", "hub URL to use")
	flag.Var(&successFlag, "s", "set success [0,1]")
	alertMessagePtr := flag.String("a", "", "alert message")
	warningMessagePtr := flag.String("w", "", "warning message")

	versionPtr := flag.Bool("version", false, "show the csender version")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		fmt.Fprintln(flag.CommandLine.Output(), "  key=value\n"+
			"        Arbitrary data to send. Use multiple times.")
		fmt.Fprintln(flag.CommandLine.Output(), "See https://docs.cloudradar.io/configuring-hosts/managing-checks/custom-checks#sending-data-using-csender")

		fmt.Fprintln(flag.CommandLine.Output(), "")
		fmt.Fprintf(flag.CommandLine.Output(), `Example:
%s -t <TOKEN> -n <CHECK_NAME> -s 1 -a "This text triggers an alert. Optional" -w "This text triggers a warning. Optional" any_number=1 any_float=0.1245 any_string="Put your check result here"`+"\n", os.Args[0])
	}
	flag.Parse()

	if *versionPtr {
		printVersion()
		return
	}

	if tokenPtr == nil || *tokenPtr == "" {
		fatal("-t token arg is required")
	}

	if checkNamePtr == nil || *checkNamePtr == "" {
		fatal("-n check name arg is required")
	}

	if hubURLPtr == nil || *hubURLPtr == "" {
		fatal("-u hub url arg can't be empty")
	}

	cs := csender.Csender{
		HubURL:    *hubURLPtr,
		HubToken:  *tokenPtr,
		CheckName: *checkNamePtr,
		HubGzip:   true,
	}

	var kvParams []string
	var skipNext bool
	for _, arg := range os.Args[1:] {
		if skipNext {
			skipNext = false
			continue
		}

		if strings.HasPrefix("-", arg) {
			skipNext = true
			continue
		}

		if !strings.Contains(arg, "=") {
			continue
		}

		kvParams = append(kvParams, arg)
	}

	err := cs.AddMultipleKeyValue(kvParams)
	if err != nil {
		fatal(err.Error())
	}

	if successFlag.set {
		err := cs.SetSuccess(successFlag.value)
		if err != nil {
			fatal(err.Error())
		}
	}

	if alertMessagePtr != nil && *alertMessagePtr != "" {
		err := cs.SetAlert(*alertMessagePtr)
		if err != nil {
			fatal(err.Error())
		}
	}

	if warningMessagePtr != nil && *warningMessagePtr != "" {
		err := cs.SetWarning(*warningMessagePtr)
		if err != nil {
			fatal(err.Error())
		}
	}

	retries := 0
	retryLimit := 5
	retryInterval := 2 * time.Second
	var retryIn time.Duration

	for {
		err = cs.Send()
		if err == nil {
			// graceful exit on success
			return
		}

		if err == cagent.ErrHubTooManyRequests {
			// for error code 429, wait 10 seconds and try again
			retryIn = 10 * time.Second
		} else if err == cagent.ErrHubServerError {
			// for error codes 5xx, wait for 2 seconds and try again
			retryIn = retryInterval
			retries++
			if retries > retryLimit {
				log.Errorf("csender: hub connection error, giving up")
				return
			} else {
				log.Infof("csender: hub connection error %d/%d, retrying in %v", retries, retryLimit, retryInterval)
			}
		} else {
			fatal(err.Error())
		}

		select {
		case <-time.After(retryIn):
			continue
		}
	}
}

func printVersion() {
	fmt.Printf("csender - tool for sending custom check results to Hub.\nPart of cagent package v%s %s\n", cagent.Version, cagent.LicenseInfo)
}
