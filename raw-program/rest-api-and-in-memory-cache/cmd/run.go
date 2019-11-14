package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"../application"
	"../portadapter"
	"github.com/patrickmn/go-cache"
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
	Use:   "run",                 // FIXME: rename that
	Short: "starts program_name", // FIXME: rename that
	Long: `I'am // FIXME: rename that
	a
	log
	help fo example`,
	Run: func(cmd *cobra.Command, args []string) {
		logging.WithFields(logrus.Fields{
			"entity":             rootEntity,
			"event uuid":         uuidForRootProcess,
			"Config file path":   viperConfig.GetString(configFilePathName),
			"Path to program id": viperConfig.GetString(programIDName),
			"Log output":         viperConfig.GetString(logOutputName),
			"Log level":          viperConfig.GetString(logLevelName),
			"Log format":         viperConfig.GetString(logFormatName),
			"Syslog tag":         viperConfig.GetString(syslogTagName),
			"Rest API ip":        viperConfig.GetString(restAPIIPName),
			"Rest API port":      viperConfig.GetString(restAPIPortName),
			"Rest API help info, avalible at root handler '/'": viperConfig.GetString(helpInfoName),
			"In-memory cache expire time":                      viperConfig.GetDuration(inMemoryCacheExpireName),
			"In-memory cache refresh time":                     viperConfig.GetDuration(inMemoryCacheRefreshName),
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

		shutdownCommandForRestAPI := make(chan struct{}, 1) // if shutdown when rest api down (breake restart rest api)
		gracefulShutdownCommandForRestAPI := make(chan struct{}, 1)
		restAPIisDone := make(chan struct{}, 1)
		shutdownCommandsChannels = append(shutdownCommandsChannels, shutdownCommandForRestAPI) // breake restart rest api
		shutdownCommandsChannels = append(shutdownCommandsChannels, gracefulShutdownCommandForRestAPI)
		doneChannels = append(doneChannels, restAPIisDone)

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

		// make in-memory cache
		inMemoryCache := cache.New(viperConfig.GetDuration(inMemoryCacheExpireName), viperConfig.GetDuration(inMemoryCacheRefreshName))

		// place for facade
		facade := application.NewProgramFacade(program,
			uuidGenerator,
			inMemoryCache,
			logging)

		uuidForRestAPI := uuidGenerator.NewUUID().UUID.String()
		restAPI := application.NewRestAPIentity(viperConfig.GetString(restAPIIPName),
			viperConfig.GetString(restAPIPortName),
			viperConfig.GetString(helpInfoName),
			uuidForRestAPI,
			facade)

		// place for start entities
		go restAPI.UpRestAPI(signalChan, shutdownCommandForRestAPI)
		go restAPI.GracefulShutdownRestAPI(gracefulShutdownCommandForRestAPI, restAPIisDone)

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
