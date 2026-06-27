package cmd

import (
	"io"
	"os"

	"github.com/spf13/cobra"
)

var combineCmd = &cobra.Command{
	Use:   "combine [files...]",
	Short: "Combine multiple files into one",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		output, _ := cmd.Flags().GetString("output")
		delimiter, _ := cmd.Flags().GetString("delimiter")

		for _, f := range args {
			if _, err := os.Stat(f); err != nil {
				if os.IsNotExist(err) {
					return ErrFileNotFound
				}
				return err
			}
		}

		outFile, err := os.Create(output)
		if err != nil {
			return err
		}
		defer outFile.Close()

		for i, f := range args {
			inFile, err := os.Open(f)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, inFile); err != nil {
				inFile.Close()
				return err
			}
			inFile.Close()

			if i < len(args)-1 && delimiter != "" {
				if _, err := outFile.WriteString(delimiter); err != nil {
					return err
				}
			}
		}
		return nil
	},
}

func init() {
	combineCmd.Flags().StringP("output", "o", "combined.txt", "Output file path")
	combineCmd.Flags().StringP("delimiter", "d", "", "Delimiter to insert between files")
}