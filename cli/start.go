/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"log"
	"log/slog"
	"smDaemon/daemon"
	utils "utils"

	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "starts the daemon service",
	Long: `Command "start" launch the daemon service. It starts the background screen
	monitoring and recording program of LiScreMon`,
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

		// memoryProfileFile, err := os.Create("memoryProfile.prof")
		// if err != nil {
		// 	panic(err)
		// }
		// defer memoryProfileFile.Close()
		// if err := pprof.StartCPUProfile(memoryProfileFile); err != nil {
		// 	panic(err)
		// }
		// defer pprof.StopCPUProfile()

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
	rootCmd.AddCommand(startCmd)
}
