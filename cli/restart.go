/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"log"
	"log/slog"
	"smDaemon/daemon"
	"time"
	utils "utils"

	"github.com/spf13/cobra"
)

// backgroundCmd represents the background command
var backgroundCmd = &cobra.Command{
	Use:   "restart",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		// cpuProfileFile, err := os.Create("cpuProfile.prof")
		// if err != nil {
		// 	panic(err)
		// }
		// defer cpuProfileFile.Close()
		// if err := pprof.StartCPUProfile(cpuProfileFile); err != nil {
		// 	panic(err)
		// }
		// defer pprof.StopCPUProfile()

		if err := stopScreenMonitor(); err != nil {
			log.Println("could not sucessfully shutdown running LiScreMon", err)
			return
		}

		log.Println("smDaemon would be restarting now...")
		time.Sleep(2 * time.Second) // allow for all resources to be released

		// Logging
		mode, err := cmd.Flags().GetBool("mode")
		if err != nil {
			log.Fatalln("err getting build mode in flag command:", err) // exit
		}

		logger, logFile, err := utils.Logger("daemon.log", mode)
		if err != nil {
			log.Fatalln(err) // exit
		}
		defer logFile.Close()

		slog.SetDefault(logger)

		if err := daemon.DaemonServiceLinux(logger); err != nil {
			logger.Error(err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(backgroundCmd)
}
