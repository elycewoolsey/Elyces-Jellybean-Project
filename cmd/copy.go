package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy [src] [dst]",
	Short: "Copy a file",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		src, dst := args[0], args[1]
		force, _ := cmd.Flags().GetBool("force")

		if err := validateFilename(src); err != nil {
			return err
		}
		if err := validateFilename(dst); err != nil {
			return err
		}

		if err := validateExtension(dst); err != nil {
			return err
		}

		if err := validatePathLength(dst); err != nil {
			return err
		}

		if _, err := os.Stat(src); err != nil {
			if os.IsNotExist(err) {
				return ErrFileNotFound
			}
			if os.IsPermission(err) {
				return ErrPermissionDenied
			}
			return err
		}

		if !force {
			if _, err := os.Stat(dst); err == nil {
				fmt.Print("Do you want to overwrite the file [Y/N]? ")
				scanner := bufio.NewScanner(cmd.InOrStdin())
				scanner.Scan()
				response := strings.ToUpper(strings.TrimSpace(scanner.Text()))
				if response != "Y" && response != "YES" {
					return errors.New("copy cancelled")
				}
			}
		}

		inFile, err := os.Open(src)
		if err != nil {
			if os.IsPermission(err) {
				return ErrPermissionDenied
			}
			return err
		}
		defer inFile.Close()

		outFile, err := os.Create(dst)
		if err != nil {
			if os.IsPermission(err) {
				return ErrPermissionDenied
			}
			return err
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, inFile)
		if err != nil {
			if os.IsPermission(err) {
				return ErrPermissionDenied
			}
			return err
		}
		return nil
	},
}

func init() {
	copyCmd.Flags().BoolP("force", "f", false, "Overwrite destination if exists")
}