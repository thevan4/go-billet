package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"../application"
	"../domain"
	"../portadapter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const programLogName = "program-name" // FIXME: rename that

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows program version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version is %v\ncommit: %v\nbranch: %v", version, commit, branch)
	},
}

var runCmd = &cobra.Command{
	Use:   "run", // FIXME: rename that
	Short: "starts program_name",
	Long: `I'am
	a
	log
	help fo example`,
	Run: func(cmd *cobra.Command, args []string) {
		logging.WithFields(logrus.Fields{
			"entity":                          rootEntity,
			"event uuid":                      uuidForRootProcess,
			"Config file path":                viperConfig.GetString(configFilePathName),
			"Path to program id":              viperConfig.GetString(programIDName),
			"Log output":                      viperConfig.GetString(logOutputName),
			"Log level":                       viperConfig.GetString(logLevelName),
			"Log format":                      viperConfig.GetString(logFormatName),
			"Syslog tag":                      viperConfig.GetString(syslogTagName),
			"Send topic":                      viperConfig.GetString(sendTopicName),
			"Kafka servers":                   transportServers,
			"Listen topics names":             viperConfig.GetString(listenTopicsNames),
			"Kafka group ID":                  viperConfig.GetString(kafkaGroupIDName),
			"Prometheus ip":                   viperConfig.GetString(prometheusIPName),
			"Prometheus port":                 viperConfig.GetString(prometheusPortName),
			"Max parallel jobs from consumer": viperConfig.GetInt(maxConsumerJobName),
			// add some new here
		}).Info("Start config:")

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// place for prepare graceful shutdown for all entities
		shutdownCommandsChannels := []chan struct{}{}
		doneChannels := []chan struct{}{}

		shutdownCommandForSaramaListenAndHandle := make(chan struct{}, 1)
		saramaListenAndHandleIsDone := make(chan struct{}, 1)
		shutdownCommandsChannels = append(shutdownCommandsChannels, shutdownCommandForSaramaListenAndHandle)
		doneChannels = append(doneChannels, saramaListenAndHandleIsDone)

		// place for create entities
		// create program struct
		uuidForCreateProgram := uuidGenerator.NewUUID().UUID.String()
		programLocal := portadapter.NewProgramConfigLocal("some", logging)
		program, err := programLocal.GetProgram()
		if err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     programLogName,
				"event uuid": uuidForCreateProgram,
			}).Warnf("will create a new one program, can't get old: %v", err)
			program = programLocal.NewProgram()
			err = programLocal.SaveProgram(*program)
			if err != nil {
				logging.WithFields(logrus.Fields{
					"entity":     programLogName,
					"event uuid": uuidForCreateProgram,
				}).Fatalf("can't save new program id: %v", err)
			}
		}

		// create monitoring
		uuidForMonitoring := uuidGenerator.NewUUID().UUID.String()
		monitoring := domain.NewPrometheusConfig(viperConfig.GetString(prometheusIPName), viperConfig.GetString(prometheusPortName), uuidForMonitoring, logging)

		// create sarama port adapter entity
		uuidForNewSaramaPortAdapter := uuidGenerator.NewUUID().UUID.String()
		saramaPortAdapter, err := portadapter.NewSaramaPortAdapter(uuidGenerator,
			program,
			transportServers,
			listenTopics,
			viperConfig.GetString(kafkaGroupIDName),
			viperConfig.GetString(sendTopicName),
			uuidForNewSaramaPortAdapter,
			monitoring,
			viperConfig.GetInt(maxConsumerJobName),
			logging)
		if err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     programLogName,
				"event uuid": uuidForNewSaramaPortAdapter,
			}).Fatalf("can't create sarama port adapter: %v", err)
		}

		//facade part
		someSuccessStruct := domain.SomeSuccessStruct{Some: ""}
		someErrorStruct := domain.SomeErrorStruct{Some: ""}

		facade := application.NewProgramFacade(program,
			saramaPortAdapter,
			uuidGenerator,
			someSuccessStruct,
			someErrorStruct,
			monitoring,
			logging)

		//sarama subscribe
		facade.MessageBus.Subscribe("SomeSuccess", &application.SomeSuccessHandlerConfigurator{Facade: facade})
		facade.MessageBus.Subscribe("SomeError", &application.SomeErrorHandlerConfigurator{Facade: facade})

		// place for start entities
		// up monitoring
		go monitoring.UpMonitoring(uuidForMonitoring)
		err = monitoring.CheckMonitoringIsUp()
		if err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     programLogName,
				"event uuid": uuidForMonitoring,
			}).Fatalf("can't start monitoring: %v", err)
		}

		// connect consumer and subsrcibe
		uuidForConnectConsumer := uuidGenerator.NewUUID().UUID.String()
		err = saramaPortAdapter.Consumer.ConnectConsumer(listenTopics, uuidForConnectConsumer)
		if err != nil {
			logging.WithFields(logrus.Fields{
				"entity":     programLogName,
				"event uuid": uuidForConnectConsumer,
			}).Fatalf("can't connect sarama consumer, retry through 10 second. Error received: %v", err)
		}
		go saramaPortAdapter.GracefulShutdownSaramaJobs(shutdownCommandForSaramaListenAndHandle, saramaListenAndHandleIsDone)
		go saramaPortAdapter.ListenAndHandle()
		go saramaPortAdapter.MsgHandler()

		monitoring.ComplexHealthCheck()
		<-signalChan // shutdown signal
		// gracefull shutdown all entities
		for _, shutdownCommandsChannel := range shutdownCommandsChannels {
			shutdownCommandsChannel <- struct{}{}
			close(shutdownCommandsChannel)
		}

		for _, stopDoneChannel := range doneChannels {
			<-stopDoneChannel
			close(stopDoneChannel)
		}

		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
		}).Info("Program stopped")
	},
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "prints out config",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(viperConfig.AllSettings())
	},
}

var rootCmd = &cobra.Command{
	Use:   "program-name",              // FIXME: rename that
	Short: "program-name do stuff ;-)", // FIXME: rename that
}

//Execute without flags, or unknown flags. Return help in stdout
func Execute() error {
	help := &cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Long: `Help provides help for any command in the application.
Simply type ` + rootCmd.Name() + ` help [path to rootCmdommand] for full details.`,
		Run:               rootCmd.HelpFunc(),
		PersistentPreRun:  func(cmd *cobra.Command, args []string) {},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {},
	}
	help.SetOutput(os.Stderr)

	rootCmd.SetHelpCommand(help)
	rootCmd.AddCommand(runCmd, versionCmd, configCmd)
	return rootCmd.Execute()
}
