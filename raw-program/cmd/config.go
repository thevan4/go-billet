package cmd

import (
	"fmt"
	"os"

	"../portadapter" // FIXME: fix import path when use that billet
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thevan4/go-billet/logger"
)

// Default values
const (
	defaultConfigFilePath  = "./program.properties"
	defaultLogOutput       = "stdout"
	defaultLogLevel        = "trace"
	defaultLogFormat       = "default"
	defaultSystemLogTag    = ""
	defaultPathToProgramID = "./netcon-id"

// add some new here
)

// Configs
const (
	configFilePathName = "config-file-path"
	logOutputName      = "log-output"
	logLevelName       = "log-level"
	logFormatName      = "log-format"
	syslogTagName      = "syslog-tag"

	programIDName = "program-id"
	// add some new here
)

// Need builds with ldflags
var (
	version = "TBD @ ldflags"
	commit  = "TBD @ ldflags"
	branch  = "TBD @ ldflags"
)

// Links for viper and logrus logger
var (
	viperConfig        *viper.Viper
	logging            *logrus.Logger
	uuidGenerator      *portadapter.UUIDGenerator
	uuidForRootProcess string
)

// Default определяет значиния переменных конфига по умолчанию
func Default() map[string]interface{} {
	return map[string]interface{}{
		logOutputName: defaultLogOutput,
		logLevelName:  defaultLogLevel,
		logFormatName: defaultLogFormat,
		syslogTagName: defaultSystemLogTag,
		programIDName: defaultPathToProgramID,
		// add some new here
	}
}

func applyDefaultToViper(viperConfig *viper.Viper) {
	for k, v := range Default() {
		viperConfig.SetDefault(k, v)
	}
}

func init() {
	var err error

	// make uuid generator adn uuid for root process
	uuidGenerator = portadapter.NewUUIDGenerator()
	uuidForRootProcess = uuidGenerator.NewUUID().UUID.String()

	// init default viper config
	viperConfig = viper.New()
	applyDefaultToViper(viperConfig)

	// init default logs
	logging, err = logger.NewLogger(viperConfig.GetString(logOutputName),
		viperConfig.GetString(logLevelName),
		viperConfig.GetString(logFormatName),
		viperConfig.GetString(syslogTagName))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// work with flags
	pflag.StringP(configFilePathName, "c", defaultConfigFilePath, "Path to config file. Example value: './program.properties'")

	pflag.String(logOutputName, defaultLogOutput, "Log output. Example values: 'stdout', 'syslog'")
	pflag.String(logLevelName, defaultLogLevel, "Log level. Example values: 'info', 'debug', 'trace'")
	pflag.String(logFormatName, defaultLogFormat, "Log format. Example values: 'default', 'json'")
	pflag.String(syslogTagName, defaultSystemLogTag, "Syslog tag. Example value: 'some_sys_tag'")

	pflag.String(programIDName, defaultPathToProgramID, "Path to program ID. Example value: './program-id'")

	// work with config file
	viperConfig.SetConfigFile(defaultConfigFilePath) // FIXME:
	err = viperConfig.ReadInConfig()
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Warnf("can't find default config file: %v", err)
		err = viperConfig.WriteConfig()
		if err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Fatalf("can't create default config file: %v", err)
		}
	}

	// apply flags
	pflag.Parse()
	viperConfig.BindPFlags(pflag.CommandLine)

	// change logging if needed
	err = logger.ApplyLoggerOut(logging, viperConfig.GetString(logOutputName), viperConfig.GetString(syslogTagName))
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Fatalf("can't apply new config for logger: %v", err)
	}

	err = logger.ApplyLoggerLogLevel(logging, viperConfig.GetString(logLevelName))
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Fatalf("can't apply new config for logger: %v", err)
	}

	err = logger.ApplyLogFormatter(logging, viperConfig.GetString(logFormatName))
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Fatalf("can't apply new config for logger: %v", err)
	}

	// Watch to config file
	go func() {
		viperConfig.WatchConfig()
		viperConfig.OnConfigChange(func(e fsnotify.Event) {
			logging.WithFields(logrus.Fields{
				"entity":     rootEntity,
				"event uuid": uuidForRootProcess,
			}).Warnf("Config file changed: %v. Current config: %v", e.Name, viperConfig.AllSettings())
		})
	}()
}
