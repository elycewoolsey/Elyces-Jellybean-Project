package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "fileops",
	Short: "File operations CLI",
	Long:  `Create, delete, combine, and copy files.`,
}

func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(combineCmd)
	rootCmd.AddCommand(copyCmd)
}