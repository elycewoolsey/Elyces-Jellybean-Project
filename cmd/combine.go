package cmd

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var combineCmd = newCombineCmd()

func newCombineCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "combine [files...]",
		Short: "Combine multiple files into one",
		Long: `Combine multiple files into a single output file.

The write is atomic: content goes to a temporary file that is renamed into
place only on success, so the output is never left half-written. Files are
joined in order, with the optional delimiter inserted between them. If the
output already exists, a Y/N prompt is shown unless --force is given.`,
		Example: `  fileops combine a.txt b.txt
  fileops combine a.txt b.txt -o out.txt -d "\n"`,
		Args: cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			output, _ := cmd.Flags().GetString("output")
			delimiter, _ := cmd.Flags().GetString("delimiter")
			force, _ := cmd.Flags().GetBool("force")

			for _, f := range args {
				info, err := os.Stat(f)
				if err != nil {
					return mapOSError(err)
				}
				if info.IsDir() {
					return ErrIsDirectory
				}
			}

			if err := validateFilename(output); err != nil {
				return err
			}
			if err := validateExtension(output); err != nil {
				return err
			}
			if err := validatePathLength(output); err != nil {
				return err
			}

			if !force {
				if _, err := os.Stat(output); err == nil {
					if !confirm(cmd, "Do you want to overwrite the output file [Y/N]? ") {
						return ErrCombineCancelled
					}
				}
			}

			return combineInto(output, delimiter, args)
		},
	}
	cmd.Flags().StringP("output", "o", "combined.txt", "Output file path")
	cmd.Flags().StringP("delimiter", "d", "", "Delimiter to insert between files")
	cmd.Flags().BoolP("force", "f", false, "Overwrite output without prompt")
	return cmd
}

func combineInto(output, delimiter string, files []string) error {
	outDir := filepath.Dir(output)
	tmp, err := os.CreateTemp(outDir, ".combine-*")
	if err != nil {
		return mapOSError(err)
	}
	tmpPath := tmp.Name()

	// On any failure, discard the temp file and leave the existing output
	// untouched.
	fail := func(err error) error {
		tmp.Close()
		os.Remove(tmpPath)
		return mapOSError(err)
	}

	for i, f := range files {
		in, err := os.Open(f)
		if err != nil {
			return fail(err)
		}
		_, copyErr := io.Copy(tmp, in)
		in.Close()
		if copyErr != nil {
			return fail(copyErr)
		}

		if i < len(files)-1 && delimiter != "" {
			if _, err := tmp.WriteString(delimiter); err != nil {
				return fail(err)
			}
		}
	}

	if err := tmp.Sync(); err != nil {
		return fail(err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpPath)
		return mapOSError(err)
	}
	if err := os.Rename(tmpPath, output); err != nil {
		os.Remove(tmpPath)
		return mapOSError(err)
	}
	return nil
}
