package cagent

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelError LogLevel = "error"
)

func (lvl LogLevel) IsValid() bool {
	switch lvl {
	case LogLevelDebug:
		fallthrough
	case LogLevelInfo:
		fallthrough
	case LogLevelError:
		return true
	default:
		return false
	}
}

func (lvl LogLevel) LogrusLevel() logrus.Level {
	switch lvl {
	case LogLevelDebug:
		return logrus.DebugLevel
	case LogLevelError:
		return logrus.ErrorLevel
	default:
		return logrus.InfoLevel
	}
}

type logrusFileHook struct {
	file      *os.File
	flag      int
	chmod     os.FileMode
	formatter *logrus.TextFormatter
}

func addLogFileHook(file string, flag int, chmod os.FileMode) error {

	dir := filepath.Dir(file)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		logrus.WithError(err).Errorf("Failed to create the logs dir: '%s'", dir)
	}

	plainFormatter := &logrus.TextFormatter{FullTimestamp: true, DisableColors: true}
	logFile, err := os.OpenFile(file, flag, chmod)
	if err != nil {
		return fmt.Errorf("Unable to write log file: %s", err.Error())
	}

	hook := &logrusFileHook{logFile, flag, chmod, plainFormatter}

	logrus.AddHook(hook)

	return nil
}

// Fire event
func (hook *logrusFileHook) Fire(entry *logrus.Entry) error {
	plainformat, err := hook.formatter.Format(entry)
	line := string(plainformat)
	_, err = hook.file.WriteString(line)
	if err != nil {
		fmt.Fprintf(os.Stderr, "unable to write file on filehook(entry.String)%v", err)
		return err
	}

	return nil
}

func (hook *logrusFileHook) Levels() []logrus.Level {
	return []logrus.Level{
		logrus.PanicLevel,
		logrus.FatalLevel,
		logrus.ErrorLevel,
		logrus.WarnLevel,
		logrus.InfoLevel,
		logrus.DebugLevel,
	}
}

// Sets Log level and corresponding logrus level
func (ca *Cagent) SetLogLevel(lvl LogLevel) {
	ca.Config.LogLevel = lvl
	logrus.SetLevel(lvl.LogrusLevel())
}

func (ca *Cagent) configureLogger() {
	tfmt := logrus.TextFormatter{FullTimestamp: true, DisableColors: true}

	logrus.SetFormatter(&tfmt)

	ca.SetLogLevel(ca.Config.LogLevel)

	if ca.Config.LogFile != "" {
		logrus.Debug("Adding log file hook", ca.Config.LogFile)
		err := addLogFileHook(ca.Config.LogFile, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
		if err != nil {
			logrus.Error("Can't write logs to file: ", err.Error())
		}
	}

	// If a logfile is specified, syslog must be disabled and logs are written to that file and nowhere else.
	if ca.Config.LogSyslog != "" {
		logrus.Debug("Adding syslog hook", ca.Config.LogSyslog)
		err := addSyslogHook(ca.Config.LogSyslog)
		if err != nil {
			logrus.Error("Can't set up syslog: ", err.Error())
		}
	}

	// sets standard logging to /dev/null
	devNull, err := os.OpenFile(os.DevNull, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		logrus.Error("err", err)
	}
	logrus.SetOutput(devNull)
}
