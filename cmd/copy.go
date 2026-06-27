package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

var copyCmd = newCopyCmd()

func newCopyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "copy [src] [dst]",
		Short: "Copy a file, prompting on overwrite unless -f",
		Long: `Copy a file from src to dst.

The destination name is validated and both paths are checked against the
platform path-length limit. If dst already exists, a Y/N prompt is shown
unless --force is given. The destination keeps the source's permission bits.
Symlinks are followed; the link target is copied, not the link itself.`,
		Example: `  jellybeans copy a.txt b.txt
  jellybeans copy a.txt b.txt -f`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			src, dst := args[0], args[1]
			force, _ := cmd.Flags().GetBool("force")

			if err := validateFilename(dst); err != nil {
				return err
			}
			if err := validateExtension(dst); err != nil {
				return err
			}
			if err := validatePathLength(src); err != nil {
				return err
			}
			if err := validatePathLength(dst); err != nil {
				return err
			}

			info, err := os.Stat(src)
			if err != nil {
				return mapOSError(err)
			}
			if info.IsDir() {
				return ErrIsDirectory
			}

			if !force {
				if _, err := os.Stat(dst); err == nil {
					if !confirm(cmd, "Do you want to overwrite the file [Y/N]? ") {
						return ErrCopyCancelled
					}
				}
			}

			return copyFile(src, dst, info.Mode())
		},
	}
	cmd.Flags().BoolP("force", "f", false, "Overwrite destination without prompt")
	return cmd
}

// copyFile copies src to dst, giving dst the same permission bits as src.
func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return mapOSError(err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return mapOSError(err)
	}

	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return mapOSError(err)
	}
	if err := out.Sync(); err != nil {
		out.Close()
		return mapOSError(err)
	}
	if err := out.Close(); err != nil {
		return mapOSError(err)
	}

	// Ensure mode matches even when dst already existed (O_TRUNC keeps the
	// old file's mode).
	return mapOSError(os.Chmod(dst, mode))
}
