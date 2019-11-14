package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

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
			// add some new here
		}).Info("Start config:")

		// more about signals: https://en.wikipedia.org/wiki/Signal_(IPC)
		signalChan := make(chan os.Signal, 2)
		signal.Notify(signalChan, syscall.SIGHUP,
			syscall.SIGINT,
			syscall.SIGTERM,
			syscall.SIGQUIT)

		// place for prepare graceful shutdown for all entities
		// shutdownCommandsChannels := []chan struct{}{}
		// doneChannels := []chan struct{}{}

		// place for create entities
		// place for facade
		// place for start entities
		<-signalChan // shutdown signal
		// gracefull shutdown all entities
		// for _, shutdownCommandsChannel := range shutdownCommandsChannels {
		// 	shutdownCommandsChannel <- struct{}{}
		// 	close(shutdownCommandsChannel)
		// }

		// for _, stopDoneChannel := range doneChannels {
		// 	<-stopDoneChannel
		// 	close(stopDoneChannel)
		// }

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
