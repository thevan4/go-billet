package cmd

import (
	"fmt"
	"os"
	"strings"

	"../portadapter" // FIXME: fix import path when use that billet
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/thevan4/go-billet/logger"
)

const rootEntity = "root entity"

// Default values
const (
	defaultConfigFilePath      = "./program.properties"
	defaultLogOutput           = "stdout"
	defaultLogLevel            = "trace"
	defaultLogFormat           = "default"
	defaultSystemLogTag        = ""
	defaultPathToProgramID     = "./program-id"
	defaultRawTransportServers = "1.1.1.1:9092,2.2.2.2:9092"
	defaultSendTopic           = "send-topic"
	defaultKafkaGroupID        = "some-group-id"
	defaultPrometheusIP        = "0.0.0.0"
	defaultPrometheusPort      = "9000"
	defaultRawListenTopics     = "listen-topic"
	defaultMaxConsumerJob      = 1000
	// add some new here
)

// Configs
const (
	configFilePathName   = "config-file-path"
	logOutputName        = "log-output"
	logLevelName         = "log-level"
	logFormatName        = "log-format"
	syslogTagName        = "syslog-tag"
	programIDName        = "program-id"
	transportServersName = "transport-servers"
	listenTopicsNames    = "listen-topic"
	sendTopicName        = "send-topic"
	kafkaGroupIDName     = "kafka-group-id"
	prometheusIPName     = "prometheus-ip"
	prometheusPortName   = "prometheus-port"
	maxConsumerJobName   = "max-consumer-jobs"
	// add some new here
)

// For builds with ldflags
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
	transportServers   []string
	listenTopics       []string
)

// return default values
func defaultConfig() map[string]interface{} {
	return map[string]interface{}{
		configFilePathName:   defaultConfigFilePath,
		logOutputName:        defaultLogOutput,
		logLevelName:         defaultLogLevel,
		logFormatName:        defaultLogFormat,
		syslogTagName:        defaultSystemLogTag,
		programIDName:        defaultPathToProgramID,
		transportServersName: defaultRawTransportServers,
		listenTopicsNames:    defaultRawListenTopics,
		sendTopicName:        defaultSendTopic,
		kafkaGroupIDName:     defaultKafkaGroupID,
		prometheusIPName:     defaultPrometheusIP,
		prometheusPortName:   defaultPrometheusPort,
		maxConsumerJobName:   defaultMaxConsumerJob,
		// add some new here
	}
}

func applyDefaultToViper(viperConfig *viper.Viper) {
	for k, v := range defaultConfig() {
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
	pflag.String(transportServersName, defaultRawTransportServers, "Kafka transport servers. Example: '1.1.1.1:9092,2.2.2.2:9092'")
	pflag.String(listenTopicsNames, defaultRawListenTopics, "Kafka listen topics names. Example: 'listen-topic1,listem-topic2'")
	pflag.String(sendTopicName, defaultSendTopic, "Send kafka topic name.")
	pflag.String(kafkaGroupIDName, defaultKafkaGroupID, "Send kafka group id.")
	pflag.String(prometheusIPName, defaultPrometheusIP, "Prometheus ip. Example value: '0.0.0.0'")
	pflag.String(prometheusPortName, defaultPrometheusPort, "Prometheus port. Example value: '9000'")
	pflag.Int(maxConsumerJobName, defaultMaxConsumerJob, "Max parallel jobs from consumer")

	// apply config flag here, beacose need to know config file path
	pflag.Parse()
	err = viperConfig.BindPFlag(configFilePathName, pflag.Lookup(configFilePathName))
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Fatalf("can't bind config flag: %v", err)
	}

	// work with config file
	viperConfig.SetConfigFile(defaultConfigFilePath)
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

	// apply all flags
	err = viperConfig.BindPFlags(pflag.CommandLine)
	if err != nil {
		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Fatalf("can't bind some flag: %v", err)
	}

	// modify logging
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
			}).Warnf("Config file changed: %v", e.Name)
		})
	}()
	// custom modify values. viper can't write to *.properties hashmap and slice. Avoid it! Also viperConfig.SetDefault and viper.Set dosen't work
	transportServers = strings.Split(viperConfig.GetString(transportServersName), ",")
	listenTopics = strings.Split(viperConfig.GetString(listenTopicsNames), ",")
}
