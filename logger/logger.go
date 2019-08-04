package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"

	log "github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
)

// NewLogger - new logrus logger
func NewLogger(rawLogOutput, rawLogLevel, rawLogFormat, syslogTag string) (*log.Logger, error) {
	var err error
	logrusLog := log.New()

	err = applyLoggerOut(logrusLog, rawLogOutput, syslogTag)
	if err != nil {
		return nil, err
	}

	logLevel, err := log.ParseLevel(rawLogLevel)
	if err != nil {
		return nil, err
	}
	logrusLog.SetLevel(logLevel)

	err = applyLogFormatter(logrusLog, rawLogFormat)
	if err != nil {
		return nil, err
	}
	return logrusLog, nil
}

func applyLoggerOut(logrusLog *log.Logger, logOutput, syslogTag string) error {
	var out io.Writer

	switch logOutput {
	case "stdout":
		out = os.Stdout
	case "syslog":
		hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, syslogTag)
		if err != nil {
			return fmt.Errorf("can't create hook for syslog: %v", err)
		}
		logrusLog.Hooks.Add(hook)
		out = ioutil.Discard
	default:
		return fmt.Errorf("uknown log output type: %s", logOutput)
	}
	logrusLog.SetOutput(out)
	return nil
}

func applyLogFormatter(logrusLog *log.Logger, rawLogFormat string) error {
	switch rawLogFormat {
	case "json":
		logrusLog.SetFormatter(&log.JSONFormatter{})
		return nil
	case "default":
		logrusLog.SetFormatter(&log.TextFormatter{
			TimestampFormat:  "2006-01-02 15:04:05",
			FullTimestamp:    true,
			QuoteEmptyFields: true,
		})
		return nil
	default:
		return fmt.Errorf("uknown log format: %v", rawLogFormat)
	}
}
