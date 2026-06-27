package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

const maxPathLength = 260

var invalidChars = []string{
	"<", ">", ":", "\"", "/", "\\", "|", "?", "*",
}

var allowedExtensions = map[string]bool{
	".txt": true, ".md": true, ".go": true, ".json": true,
	".yaml": true, ".yml": true, ".toml": true, ".xml": true,
	".html": true, ".css": true, ".js": true, ".ts": true,
	".py": true, ".rs": true, ".java": true, ".c": true,
	".h": true, ".cpp": true, ".hpp": true, ".sh": true,
	".bat": true, ".ps1": true, ".sql": true, ".csv": true,
	"": true,
}

var createCmd = &cobra.Command{
	Use:   "create [file]",
	Short: "Create a file with content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		content, _ := cmd.Flags().GetString("content")
		force, _ := cmd.Flags().GetBool("force")

		path := args[0]

		if err := validateFilename(path); err != nil {
			return err
		}

		if err := validateExtension(path); err != nil {
			return err
		}

		if err := validatePathLength(path); err != nil {
			return err
		}

		if !force {
			if _, err := os.Stat(path); err == nil {
				return ErrFileExists
			}
		}

		dir := filepath.Dir(path)
		if dir != "." && dir != "" {
			if err := checkDirWritable(dir); err != nil {
				return err
			}
		}

		return os.WriteFile(path, []byte(content), 0644)
	},
}

func validateFilename(name string) error {
	base := filepath.Base(name)
	for _, c := range invalidChars {
		if strings.Contains(base, c) {
			return ErrInvalidChars
		}
	}
	if runtime.GOOS == "windows" {
		reserved := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
		upper := strings.ToUpper(strings.TrimSuffix(base, filepath.Ext(base)))
		for _, r := range reserved {
			if upper == r {
				return ErrInvalidChars
			}
		}
	}
	return nil
}

func validateExtension(path string) error {
	ext := strings.ToLower(filepath.Ext(path))
	if !allowedExtensions[ext] {
		return ErrInvalidExtension
	}
	return nil
}

func validatePathLength(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if len(abs) > maxPathLength {
		return ErrNameTooLong
	}
	return nil
}

func checkDirWritable(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return os.ErrNotExist
		}
		return ErrPermissionDenied
	}
	if !info.IsDir() {
		return ErrPermissionDenied
	}
	testFile := filepath.Join(dir, ".write_test_"+randomString(8))
	f, err := os.Create(testFile)
	if err != nil {
		return ErrPermissionDenied
	}
	f.Close()
	os.Remove(testFile)
	return nil
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[i%len(letters)]
	}
	return string(b)
}

func init() {
	createCmd.Flags().StringP("content", "c", "", "Content to write to file")
	createCmd.Flags().BoolP("force", "f", false, "Overwrite existing file")
}