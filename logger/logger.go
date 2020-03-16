package logger

import (
	"fmt"
	"io"
	"io/ioutil"
	"log/syslog"
	"os"

	log "github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	graylogHttpHook "github.com/thevan4/logrus-graylog-http-hook"
)

// NewLogger create new logrus logger
// func NewLogger(rawLogOutput, rawLogLevel, rawLogFormat, syslogTag string) (*log.Logger, error) {
func NewLogger(rawLogOutput []string, rawLogLevel, rawLogFormat string, extraInfo map[string]interface{}) (*log.Logger, error) {
	var err error
	logrusLog := log.New()

	for _, logOutput := range rawLogOutput {
		// extraHook, err := ApplyLoggerOut(logrusLog, logOutput, extraInfo)
		_, err := ApplyLoggerOut(logrusLog, logOutput, extraInfo)
		if err != nil {
			return nil, err
		}
	}

	err = ApplyLoggerLogLevel(logrusLog, rawLogLevel)
	if err != nil {
		return nil, err
	}

	err = ApplyLogFormatter(logrusLog, rawLogFormat)
	if err != nil {
		return nil, err
	}
	return logrusLog, nil
}

// ApplyLoggerLogLevel set log level
func ApplyLoggerLogLevel(logrusLog *log.Logger, rawLogLevel string) error {
	logLevel, err := log.ParseLevel(rawLogLevel)
	if err != nil {
		return err
	}
	logrusLog.SetLevel(logLevel)
	return nil
}

// ApplyLoggerOut set log output
func ApplyLoggerOut(logrusLog *log.Logger,
	logOutput string,
	extraInfo map[string]interface{}) (interface{}, error) {
	var out io.Writer

	switch logOutput {
	case "stdout":
		out = os.Stdout
		return nil, nil
	case "syslog":
		switch syslogTag := extraInfo["syslogTag"].(type) {
		case string:
			hook, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, syslogTag)
			if err != nil {
				return nil, fmt.Errorf("can't create hook for syslog: %v", err)
			}
			logrusLog.Hooks.Add(hook)
			out = ioutil.Discard
			logrusLog.SetOutput(out)
			return nil, nil
		default:
			return nil, fmt.Errorf("unknown type in extraInfo map for key 'syslogTag', expect string, have: %T", syslogTag)
		}

	case "graylog":
		graylogCreator, err := constructDataForGraylogHook(extraInfo)
		if err != nil {
			return nil, fmt.Errorf("can't construct data for graylog hook, got error: %v", err)
		}
		graylogHook := graylogHttpHook.NewGraylogHook(graylogCreator.graylogAddress, graylogCreator.graylogRetries, graylogCreator.greylogExtra)
		logrusLog.AddHook(graylogHook)
		out = ioutil.Discard
		logrusLog.SetOutput(out)
		return graylogHook, nil
	default:
		return nil, fmt.Errorf("uknown log output type: %s", logOutput)
	}
}

func constructDataForGraylogHook(extraInfo map[string]interface{}) (*graylogCreator, error) {
	graylogInfo := &graylogCreator{}
	for extraInfoKey, extraInfoValue := range extraInfo {
		switch extraInfoKey {
		case "graylogAddress":
			switch valueAddress := extraInfoValue.(type) {
			case string:
				graylogInfo.graylogAddress = valueAddress
			default:
				return nil, fmt.Errorf("unknown type in extraInfo map for key 'graylogAddress', expect string, have: %T", valueAddress)
			}
		case "graylogRetries":
			switch valueRetries := extraInfoValue.(type) {
			case int:
				graylogInfo.graylogRetries = valueRetries
			default:
				return nil, fmt.Errorf("unknown type in extraInfo map for key 'graylogRetries', expect int, have: %T", valueRetries)
			}
		case "greylogExtra":
			switch valueExtra := extraInfoValue.(type) {
			case map[string]interface{}:
				graylogInfo.greylogExtra = valueExtra
			default:
				return nil, fmt.Errorf("unknown type in extraInfo map for key 'graylogRetries', expect map[string]interface{}, have: %T", valueExtra)
			}
		default:
			continue
		}
	}
	return graylogInfo, nil
}

type graylogCreator struct {
	graylogAddress string
	graylogRetries int
	greylogExtra   map[string]interface{}
}

// ApplyLogFormatter set log format
func ApplyLogFormatter(logrusLog *log.Logger, rawLogFormat string) error {
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
