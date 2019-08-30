package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

const rootEntity = "root entity"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "shows program version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("version is %s\n", version)
	},
}

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "starts program_name",
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
		// place for create entities
		// place for start entities
		<-signalChan // shutdown signal

		logging.WithFields(logrus.Fields{
			"entity":     rootEntity,
			"event uuid": uuidForRootProcess,
			// add some new here
		}).Info("Program stoped")
		// place for magic graceful shutdown for all entities
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
	Use:   "program_name",
	Short: "program_name do stuff ;-)",
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
	rootCmd.AddCommand(runCmd, configCmd)
	return rootCmd.Execute()
}
