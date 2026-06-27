package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete [files...]",
	Short: "Delete one or more files",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		for _, f := range args {
			info, err := os.Stat(f)
			if err != nil {
				if os.IsNotExist(err) {
					return ErrFileNotFound
				}
				return err
			}
			if info.IsDir() {
				return ErrIsDirectory
			}
			if err := os.Remove(f); err != nil {
				if os.IsPermission(err) {
					return ErrPermissionDenied
				}
				if os.IsNotExist(err) {
					return ErrFileNotFound
				}
				return err
			}
		}
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolP("force", "f", false, "Force deletion without prompt")
}