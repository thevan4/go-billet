package logger

import (
	"fmt"
	"io/ioutil"
	"log/syslog"
	"os"

	"github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	graylogHttpHook "github.com/thevan4/logrus-graylog-http-hook"
)

// Logger ...
type Logger struct {
	Output    []string
	Level     string
	Formatter string
	SyslogTag string
	graylog   Graylog
}

// Graylog ...
type Graylog struct {
	Address string
	Retries int
	Extra   map[string]interface{}
}

// NewLogrusLogger create new logrus logger
func NewLogrusLogger(logger *Logger) (*logrus.Logger, error) {
	var err error
	logrusLog := logrus.New()

	err = logger.ApplyLoggerOut(logrusLog)
	if err != nil {
		return nil, err
	}

	logLevel, err := logrus.ParseLevel(logger.Level)
	if err != nil {
		return nil, err
	}
	logrusLog.SetLevel(logLevel)

	err = logger.ApplyLogFormatter(logrusLog)
	if err != nil {
		return nil, err
	}
	return logrusLog, nil
}

// ApplyLoggerOut set log output
func (logger *Logger) ApplyLoggerOut(logrusLog *logrus.Logger) error {
	for _, logOutput := range logger.Output {
		switch logOutput {
		case "stdout":
			logrusLog.SetOutput(os.Stdout)
			continue
		case "syslog":
			syslog, err := logrus_syslog.NewSyslogHook("", "", syslog.LOG_INFO, logger.SyslogTag)
			if err != nil {
				return fmt.Errorf("can't create syslog hook: %v", err)
			}
			logrusLog.Hooks.Add(syslog)
			logrusLog.SetOutput(ioutil.Discard)
			continue
		case "graylog":
			graylogHook := graylogHttpHook.NewGraylogHook(logger.graylog.Address, logger.graylog.Retries, logger.graylog.Extra)
			logrusLog.AddHook(graylogHook)
			logrusLog.SetOutput(ioutil.Discard)
			continue
		default:
			return fmt.Errorf("uknown log output type: %s", logOutput)
		}
	}
	return nil
}

// ApplyLogFormatter set log format
func (logger *Logger) ApplyLogFormatter(logrusLog *logrus.Logger) error {
	switch logger.Formatter {
	case "json":
		logrusLog.SetFormatter(&logrus.JSONFormatter{})
		return nil
	case "default":
		logrusLog.SetFormatter(&logrus.TextFormatter{
			TimestampFormat:  "2006-01-02 15:04:05",
			FullTimestamp:    true,
			QuoteEmptyFields: true,
		})
		return nil
	default:
		return fmt.Errorf("uknown log format: %v", logger.Formatter)
	}
}
