package logger

import (
	"fmt"
	"io/ioutil"
	"log/syslog"
	"os"
	"path"
	"runtime"

	"github.com/sirupsen/logrus"
	logrus_syslog "github.com/sirupsen/logrus/hooks/syslog"
	graylogHttpHook "github.com/thevan4/logrus-graylog-http-hook"
)

// Logger ...
type Logger struct {
	Output           []string
	Level            string
	Formatter        string
	SyslogTag        string
	Graylog          *Graylog
	LogEventLocation bool
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
			graylogHook := graylogHttpHook.NewGraylogHook(logger.Graylog.Address, logger.Graylog.Retries, logger.Graylog.Extra)
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
		jsonFormatter := &logrus.JSONFormatter{}
		if logger.LogEventLocation {
			jsonFormatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename, f.Line)
			}
		}
		logrusLog.SetFormatter(&logrus.JSONFormatter{})
	case "text":
		textFormatter := &logrus.TextFormatter{
			TimestampFormat:  "2006-01-02 15:04:05",
			FullTimestamp:    true,
			QuoteEmptyFields: true,
		}
		if logger.LogEventLocation {
			textFormatter.CallerPrettyfier = func(f *runtime.Frame) (string, string) {
				filename := path.Base(f.File)
				return fmt.Sprintf("function=%s()", f.Function), fmt.Sprintf("file=%s:%d", filename, f.Line)
			}
		}
		logrusLog.SetFormatter(textFormatter)
	default:
		return fmt.Errorf("uknown log format: %v", logger.Formatter)
	}
	return nil
}
