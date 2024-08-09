/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cli

import (
	"github.com/spf13/cobra"
)

// webCmd represents the web command
var webCmd = &cobra.Command{
	Use:     "serve",
	Aliases: []string{"webserver", "browser"},
	Short:   "launch the analytics application on a local port",
	Long: `
The web command launches the analytics application on a local port.

Usage:
  web [flags]

Flags:
  -h, --help   help for web
  -d, --detach run the server in the background (detach from terminal)

Examples:
  // Launch the analytics application on port 8080, viewable on the browser
`,
	Run: func(cmd *cobra.Command, args []string) {
	},
}

func init() {
	webCmd.Flags().BoolP("detach", "d", false, "run the server in the background (detach from terminal)")
	rootCmd.AddCommand(webCmd)
}
