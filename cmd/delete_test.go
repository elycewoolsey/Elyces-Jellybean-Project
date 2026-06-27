package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name       string
		setup      func() string
		args       []string
		flags      map[string]string
		wantErr    error
		checkFiles func(t *testing.T)
	}{
		{
			name: "delete single file",
			setup: func() string {
				f := filepath.Join(tmpDir, "test.txt")
				os.WriteFile(f, []byte("content"), 0644)
				return f
			},
			args:    []string{},
			flags:   map[string]string{},
			wantErr: nil,
			checkFiles: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(tmpDir, "test.txt")); !os.IsNotExist(err) {
					t.Error("file should be deleted")
				}
			},
		},
		{
			name: "delete multiple files",
			setup: func() string {
				f1 := filepath.Join(tmpDir, "test1.txt")
				f2 := filepath.Join(tmpDir, "test2.txt")
				os.WriteFile(f1, []byte("content1"), 0644)
				os.WriteFile(f2, []byte("content2"), 0644)
				return f1 + " " + f2
			},
			args:    []string{},
			flags:   map[string]string{},
			wantErr: nil,
			checkFiles: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(tmpDir, "test1.txt")); !os.IsNotExist(err) {
					t.Error("test1.txt should be deleted")
				}
				if _, err := os.Stat(filepath.Join(tmpDir, "test2.txt")); !os.IsNotExist(err) {
					t.Error("test2.txt should be deleted")
				}
			},
		},
		{
			name: "file not found",
			setup: func() string {
				return filepath.Join(tmpDir, "nonexistent.txt")
			},
			args:    []string{},
			flags:   map[string]string{},
			wantErr: ErrFileNotFound,
		},
		{
			name: "force delete directory returns error",
			setup: func() string {
				d := filepath.Join(tmpDir, "dir")
				os.Mkdir(d, 0755)
				return d
			},
			args:    []string{},
			flags:   map[string]string{"force": "true"},
			wantErr: errors.New("is a directory"),
			checkFiles: func(t *testing.T) {
				if _, err := os.Stat(filepath.Join(tmpDir, "dir")); err != nil {
					t.Error("directory should still exist")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var target string
			if tt.setup != nil {
				target = tt.setup()
			}
			if len(tt.args) == 0 && target != "" {
				tt.args = strings.Fields(target)
			}

			cmd := deleteCmd
			cmd.SetArgs(tt.args)
			for k, v := range tt.flags {
				cmd.Flags().Set(k, v)
			}

			err := cmd.RunE(cmd, tt.args)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr.Error()) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErr.Error(), err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFiles != nil {
				tt.checkFiles(t)
			}
		})
	}
}

func TestDeleteCommandPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()
	restrictedDir := filepath.Join(tmpDir, "restricted")
	os.Mkdir(restrictedDir, 0000)
	defer os.Chmod(restrictedDir, 0700)

	t.Run("permission denied", func(t *testing.T) {
		f := filepath.Join(restrictedDir, "test.txt")
		os.WriteFile(f, []byte("content"), 0644)

		cmd := deleteCmd
		cmd.SetArgs([]string{f})

		err := cmd.RunE(cmd, cmd.Flags().Args())
		if err == nil {
			t.Fatal("expected permission denied error")
		}
		if !strings.Contains(err.Error(), "Permission denied") {
			t.Fatalf("expected permission denied, got %q", err.Error())
		}
	})
}