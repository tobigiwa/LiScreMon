/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"syscall"

	"strings"

	"github.com/spf13/cobra"
)

// stopCmd represents the stop command
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the daemon service",
	Long: `The 'stop' command is used to halt the execution of the CLI application. 
When this command is invoked, it prints "stop called" to the standard output and then terminates the application.`,
	Run: func(cmd *cobra.Command, args []string) {

		if err := stopScreenMonitor(); err != nil {
			log.Println(err)
		}
		log.Println("smDaemon stopped successfully")

	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
}

func stopScreenMonitor() (Error error) {

	result, err := exec.Command("pidof", "smDaemon").CombinedOutput()
	if err != nil {
		return err
	}

	arrayOfpidByte := bytes.Split(result, []byte(" "))

	if len(arrayOfpidByte) == 1 { // the running program
		log.Println("smDaemon was not running")
		return nil
	}

	for i := 1; i < len(arrayOfpidByte); i++ { // the first one is the running program itself
		otherProcess, err := strconv.Atoi(strings.TrimSpace(string(arrayOfpidByte[i])))
		if err != nil {
			Error = fmt.Errorf("%w:%w", Error, err)
			continue
		}

		if err := syscall.Kill(otherProcess, syscall.SIGTERM); err != nil {
			Error = fmt.Errorf("%w:%w", Error, err)
		}
	}
	return
}
