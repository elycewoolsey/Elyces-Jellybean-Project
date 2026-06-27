package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCopyCommand(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (src, dst string)
		flags       map[string]string
		input       string
		wantContent string
		wantErr     error
	}{
		{
			name: "copy file",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				src := filepath.Join(dir, "src.txt")
				dst := filepath.Join(dir, "dst.txt")
				if err := os.WriteFile(src, []byte("hello"), 0644); err != nil {
					t.Fatal(err)
				}
				return src, dst
			},
			flags:       map[string]string{},
			wantContent: "hello",
		},
		{
			name: "copy with force overwrite",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				src := filepath.Join(dir, "src.txt")
				dst := filepath.Join(dir, "dst.txt")
				if err := os.WriteFile(src, []byte("new content"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(dst, []byte("old content"), 0644); err != nil {
					t.Fatal(err)
				}
				return src, dst
			},
			flags:       map[string]string{"force": "true"},
			wantContent: "new content",
		},
		{
			name: "invalid chars in destination",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				src := filepath.Join(dir, "src.txt")
				if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
					t.Fatal(err)
				}
				return src, filepath.Join(dir, "dst<file>.txt")
			},
			flags:   map[string]string{},
			wantErr: ErrInvalidChar,
		},
		{
			name: "invalid extension in destination",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				src := filepath.Join(dir, "src.txt")
				if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
					t.Fatal(err)
				}
				return src, filepath.Join(dir, "dst.exe")
			},
			flags:   map[string]string{},
			wantErr: ErrInvalidExtension,
		},
		{
			name: "source not found",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				return filepath.Join(dir, "missing.txt"), filepath.Join(dir, "dst.txt")
			},
			flags:   map[string]string{},
			wantErr: ErrFileNotFound,
		},
		{
			name: "destination path too long",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				src := filepath.Join(dir, "src.txt")
				if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
					t.Fatal(err)
				}
				return src, filepath.Join(dir, strings.Repeat("a", 5000)+".txt")
			},
			flags:   map[string]string{},
			wantErr: ErrNameTooLong,
		},
		{
			name: "source path too long",
			setup: func(t *testing.T) (string, string) {
				dir := t.TempDir()
				longSrc := filepath.Join(dir, strings.Repeat("b", 5000)+".txt")
				return longSrc, filepath.Join(dir, "dst.txt")
			},
			flags:   map[string]string{},
			wantErr: ErrNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setup(t)

			c := newCopyCmd()
			c.SetArgs([]string{src, dst})
			setFlags(t, c, tt.flags)
			runWithStdin(c, tt.input)

			err := c.RunE(c, []string{src, dst})
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

			content, err := os.ReadFile(dst)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			if string(content) != tt.wantContent {
				t.Errorf("content = %q, want %q", string(content), tt.wantContent)
			}
		})
	}
}

func TestCopyCommandPrompt(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantErr     error
		wantContent string
	}{
		{name: "yes overwrites", input: "Y\n", wantContent: "new"},
		{name: "YES overwrites", input: "YES\n", wantContent: "new"},
		{name: "no cancels", input: "N\n", wantErr: ErrCopyCancelled, wantContent: "old"},
		{name: "empty input cancels", input: "", wantErr: ErrCopyCancelled, wantContent: "old"},
		{name: "eof cancels", input: "", wantErr: ErrCopyCancelled, wantContent: "old"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			src := filepath.Join(dir, "src.txt")
			dst := filepath.Join(dir, "dst.txt")
			if err := os.WriteFile(src, []byte("new"), 0644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(dst, []byte("old"), 0644); err != nil {
				t.Fatal(err)
			}

			c := newCopyCmd()
			c.SetArgs([]string{src, dst})
			runWithStdin(c, tt.input)

			err := c.RunE(c, []string{src, dst})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			content, _ := os.ReadFile(dst)
			if string(content) != tt.wantContent {
				t.Errorf("content = %q, want %q", string(content), tt.wantContent)
			}
		})
	}
}

func TestCopyCommandPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	t.Run("read permission denied", func(t *testing.T) {
		dir := t.TempDir()
		restricted := filepath.Join(dir, "restricted")
		if err := os.Mkdir(restricted, 0755); err != nil {
			t.Fatal(err)
		}
		// Create the file while the directory is still writable.
		src := filepath.Join(restricted, "src.txt")
		if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
		// Now lock down the directory so the file cannot be read.
		if err := os.Chmod(restricted, 0000); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(restricted, 0755)

		dst := filepath.Join(dir, "dst.txt")
		c := newCopyCmd()
		c.SetArgs([]string{src, dst})

		err := c.RunE(c, []string{src, dst})
		if !errors.Is(err, ErrPermissionDenied) {
			t.Fatalf("expected ErrPermissionDenied, got %v", err)
		}
	})

	t.Run("write permission denied", func(t *testing.T) {
		dir := t.TempDir()
		src := filepath.Join(dir, "src.txt")
		if err := os.WriteFile(src, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}

		restricted := filepath.Join(dir, "restricted")
		if err := os.Mkdir(restricted, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.Chmod(restricted, 0555); err != nil {
			t.Fatal(err)
		}
		defer os.Chmod(restricted, 0755)

		dst := filepath.Join(restricted, "dst.txt")
		c := newCopyCmd()
		c.SetArgs([]string{src, dst})

		err := c.RunE(c, []string{src, dst})
		if !errors.Is(err, ErrPermissionDenied) {
			t.Fatalf("expected ErrPermissionDenied, got %v", err)
		}
	})
}

func TestCopyPreservesMode(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping mode test on Windows")
	}

	dir := t.TempDir()
	src := filepath.Join(dir, "src.sh")
	dst := filepath.Join(dir, "dst.sh")
	if err := os.WriteFile(src, []byte("#!/bin/sh\n"), 0755); err != nil {
		t.Fatal(err)
	}

	c := newCopyCmd()
	c.SetArgs([]string{src, dst})

	if err := c.RunE(c, []string{src, dst}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	info, err := os.Stat(dst)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0755 {
		t.Errorf("dst mode = %o, want 0755", info.Mode().Perm())
	}
}

func TestCopyDirectoryReturnsError(t *testing.T) {
	dir := t.TempDir()
	srcDir := filepath.Join(dir, "subdir")
	if err := os.Mkdir(srcDir, 0755); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "dst.txt")

	c := newCopyCmd()
	c.SetArgs([]string{srcDir, dst})

	err := c.RunE(c, []string{srcDir, dst})
	if !errors.Is(err, ErrIsDirectory) {
		t.Fatalf("expected ErrIsDirectory, got %v", err)
	}
}
