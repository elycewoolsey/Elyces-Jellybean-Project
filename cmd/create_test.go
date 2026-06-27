package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCreateCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		args        []string
		flags       map[string]string
		wantFile    string
		wantErr     error
		wantContent string
		setup       func(t *testing.T, path string)
	}{
		{
			name:        "create empty file",
			args:        []string{filepath.Join(tmpDir, "empty.txt")},
			flags:       map[string]string{},
			wantFile:    filepath.Join(tmpDir, "empty.txt"),
			wantContent: "",
		},
		{
			name:        "create file with content via --content",
			args:        []string{filepath.Join(tmpDir, "hello.txt")},
			flags:       map[string]string{"content": "hello world"},
			wantFile:    filepath.Join(tmpDir, "hello.txt"),
			wantContent: "hello world",
		},
		{
			name:    "invalid characters in filename",
			args:    []string{filepath.Join(tmpDir, "bad<name>.txt")},
			flags:   map[string]string{},
			wantErr: ErrInvalidChar,
		},
		{
			name:    "invalid extension",
			args:    []string{filepath.Join(tmpDir, "file.exe")},
			flags:   map[string]string{},
			wantErr: ErrInvalidExtension,
		},
		{
			name:    "file already exists without force",
			args:    []string{filepath.Join(tmpDir, "exists.txt")},
			flags:   map[string]string{},
			wantErr: ErrFileExists,
			setup: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("first"), 0644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name: "force overwrites existing file",
			args: []string{filepath.Join(tmpDir, "force.txt")},
			flags: map[string]string{
				"force":   "true",
				"content": "new",
			},
			wantFile:    filepath.Join(tmpDir, "force.txt"),
			wantContent: "new",
			setup: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("old"), 0644); err != nil {
					t.Fatal(err)
				}
			},
		},
		{
			name:    "path too long",
			args:    []string{filepath.Join(tmpDir, strings.Repeat("a", 5000)+".txt")},
			flags:   map[string]string{},
			wantErr: ErrNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup(t, tt.args[0])
			}

			c := newCreateCmd()
			c.SetArgs(tt.args)
			setFlags(t, c, tt.flags)

			err := c.RunE(c, tt.args)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			content, err := os.ReadFile(tt.wantFile)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			if string(content) != tt.wantContent {
				t.Errorf("content = %q, want %q", string(content), tt.wantContent)
			}
		})
	}
}

func TestCreateCommandShortFlag(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "short.txt")

	c := newCreateCmd()
	// Exercise the -c shorthand via real flag parsing rather than Flags().Set.
	c.SetArgs([]string{"-c", "short flag content", path})

	if err := c.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != "short flag content" {
		t.Errorf("content = %q, want %q", string(content), "short flag content")
	}
}

func TestCreateCommandReservedName(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Skipping reserved-name test on non-Windows")
	}

	tmpDir := t.TempDir()
	c := newCreateCmd()
	c.SetArgs([]string{filepath.Join(tmpDir, "CON.txt")})

	err := c.RunE(c, []string{filepath.Join(tmpDir, "CON.txt")})
	if !errors.Is(err, ErrReservedName) {
		t.Fatalf("expected ErrReservedName, got %v", err)
	}
}

func TestCreateCommandNonexistentDir(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "nodir", "file.txt")

	c := newCreateCmd()
	c.SetArgs([]string{path})

	err := c.RunE(c, []string{path})
	if !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("expected ErrFileNotFound, got %v", err)
	}
}

func TestCreateCommandPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()
	roDir := filepath.Join(tmpDir, "readonly")
	if err := os.Mkdir(roDir, 0755); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(roDir, 0755)
	if err := os.Chmod(roDir, 0555); err != nil {
		t.Fatal(err)
	}

	path := filepath.Join(roDir, "file.txt")
	c := newCreateCmd()
	c.SetArgs([]string{path})

	err := c.RunE(c, []string{path})
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestCreateCommandMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping mode test on Windows")
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "script.sh")

	c := newCreateCmd()
	c.SetArgs([]string{"-m", "0600", path})
	if err := c.Execute(); err != nil {
		t.Fatalf("Execute: %v", err)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0600 {
		t.Errorf("mode = %o, want 0600", info.Mode().Perm())
	}
}

func TestCreateCommandInvalidMode(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")

	c := newCreateCmd()
	c.SetArgs([]string{path})
	setFlags(t, c, map[string]string{"mode": "notoctal"})

	err := c.RunE(c, []string{path})
	if !errors.Is(err, ErrInvalidMode) {
		t.Fatalf("expected ErrInvalidMode, got %v", err)
	}
}

func TestFileExtension(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"file.txt", ".txt"},
		{"FILE.TXT", ".txt"},
		{"noext", ""},
		{".gitignore", ""},
		{".env", ""},
		{"archive.tar.gz", ".gz"},
		{"a/b/c.md", ".md"},
	}
	for _, tt := range tests {
		if got := fileExtension(tt.path); got != tt.want {
			t.Errorf("fileExtension(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestCreateAllowsDotfile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".gitignore")

	c := newCreateCmd()
	c.SetArgs([]string{path})
	if err := c.RunE(c, []string{path}); err != nil {
		t.Fatalf("expected dotfile to be allowed, got %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("dotfile not created: %v", err)
	}
}
