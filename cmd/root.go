// Package cmd implements the jellybeans subcommands (create, copy, combine,
// delete) and their shared validation and error handling.
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "jellybeans",
	Short: "File operations CLI",
	Long:  `Create, delete, combine, and copy files.`,
}

// Execute runs the root command and exits with a non-zero status on error.
// Cobra prints the error message itself.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(combineCmd)
	rootCmd.AddCommand(copyCmd)
}
