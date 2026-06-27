package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var deleteCmd = newDeleteCmd()

func newDeleteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete [files...]",
		Short: "Delete one or more files, prompting unless -f",
		Long: `Delete one or more files.

Unless --force is given, a Y/N prompt is shown before anything is removed.
Deletion is fail-fast: it stops at the first error and leaves the remaining
files untouched. A symlink is removed itself, not the file it points to.`,
		Example: `  fileops delete junk.txt
  fileops delete a.txt b.txt -f`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			force, _ := cmd.Flags().GetBool("force")

			// Check everything up front so we don't delete some files and
			// then fail partway through on a missing file or a directory.
			// Lstat (not Stat) so a symlink is treated as a file: we remove
			// the link, never follow it to a directory.
			for _, f := range args {
				info, err := os.Lstat(f)
				if err != nil {
					return mapOSError(err)
				}
				if info.IsDir() {
					return ErrIsDirectory
				}
			}

			if !force {
				prompt := fmt.Sprintf("Delete %d file(s)? [Y/N] ", len(args))
				if !confirm(cmd, prompt) {
					return ErrDeleteCancelled
				}
			}

			for _, f := range args {
				if err := os.Remove(f); err != nil {
					return mapOSError(err)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolP("force", "f", false, "Delete without confirmation prompt")
	return cmd
}
