package cmd

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

func maxPathLength() int {
	if runtime.GOOS == "windows" {
		return 260
	}
	return 4096
}

var invalidChars = []string{
	"<", ">", ":", "\"", "|", "?", "*",
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

var reservedWindowsNames = []string{
	"CON", "PRN", "AUX", "NUL",
	"COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9",
	"LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9",
}

var createCmd = newCreateCmd()

func newCreateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [file]",
		Short: "Create a file with optional content",
		Long: `Create a file, optionally writing content provided via -c.

The filename is validated and an existing file is rejected unless --force is
given. Use --mode to set the file permission bits (octal, default 0644).`,
		Example: `  fileops create notes.txt
  fileops create greeting.txt -c "hello world"
  fileops create over.txt -c "x" -f
  fileops create script.sh -m 0755`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			content, _ := cmd.Flags().GetString("content")
			force, _ := cmd.Flags().GetBool("force")
			modeStr, _ := cmd.Flags().GetString("mode")

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

			mode, err := parseMode(modeStr)
			if err != nil {
				return err
			}

			return writeFile(path, []byte(content), mode, force)
		},
	}
	cmd.Flags().StringP("content", "c", "", "Content to write to file")
	cmd.Flags().BoolP("force", "f", false, "Overwrite existing file")
	cmd.Flags().StringP("mode", "m", "0644", "File permission bits (octal)")
	return cmd
}

// writeFile creates path with the given content and mode. When force is false
// it uses O_EXCL so creation fails atomically if the file already exists,
// closing the check-then-write race. When force is true an existing file is
// truncated and overwritten.
func writeFile(path string, content []byte, mode os.FileMode, force bool) error {
	flags := os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	if !force {
		flags = os.O_CREATE | os.O_WRONLY | os.O_EXCL
	}

	f, err := os.OpenFile(path, flags, mode)
	if err != nil {
		if errors.Is(err, fs.ErrExist) {
			return ErrFileExists
		}
		return mapOSError(err)
	}

	if _, err := f.Write(content); err != nil {
		f.Close()
		return mapOSError(err)
	}
	return mapOSError(f.Close())
}

// parseMode parses an octal permission string such as "0644" or "755".
func parseMode(s string) (os.FileMode, error) {
	v, err := strconv.ParseUint(s, 8, 32)
	if err != nil {
		return 0, ErrInvalidMode
	}
	return os.FileMode(v), nil
}

func validateFilename(name string) error {
	cleaned := filepath.Clean(name)
	for _, component := range pathComponents(cleaned) {
		for _, c := range invalidChars {
			if strings.Contains(component, c) {
				return ErrInvalidChar
			}
		}
		if runtime.GOOS == "windows" {
			upper := strings.ToUpper(strings.TrimSuffix(component, filepath.Ext(component)))
			for _, r := range reservedWindowsNames {
				if upper == r {
					return ErrReservedName
				}
			}
		}
	}
	return nil
}

// pathComponents splits a path into its individual names (e.g. "a/b.txt" ->
// ["b.txt", "a"]) so each one can be validated. A Windows drive letter like
// "C:" is skipped because its colon is legitimate.
func pathComponents(path string) []string {
	path = filepath.Clean(path)
	sep := string(filepath.Separator)
	var comps []string
	for {
		base := filepath.Base(path)
		dir := filepath.Dir(path)
		if base == path || base == "." || base == sep || base == "/" {
			break
		}
		if runtime.GOOS == "windows" && len(base) == 2 && base[1] == ':' {
			break
		}
		comps = append(comps, base)
		if dir == path {
			break
		}
		path = dir
	}
	return comps
}

func validateExtension(path string) error {
	if !allowedExtensions[fileExtension(path)] {
		return ErrInvalidExtension
	}
	return nil
}

// fileExtension returns the lowercased extension used for the whitelist check.
// A dotfile with no other dot (".gitignore") and a file with no dot both count
// as having no extension. For compound names like "a.tar.gz" only the final
// part (".gz") is considered, matching filepath.Ext.
func fileExtension(path string) string {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") {
		// Strip the leading dot so ".gitignore" is treated as a name, not an
		// extension; ".tar.gz" still yields ".gz".
		base = base[1:]
	}
	return strings.ToLower(filepath.Ext(base))
}

func validatePathLength(path string) error {
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	if len(abs) > maxPathLength() {
		return ErrNameTooLong
	}
	return nil
}
