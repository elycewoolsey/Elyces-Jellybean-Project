package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestCopyCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() (string, string)
		flags       map[string]string
		input       string
		wantContent string
		wantErr     error
	}{
		{
			name: "copy file",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "src.txt")
				dst := filepath.Join(tmpDir, "dst.txt")
				os.WriteFile(src, []byte("hello"), 0644)
				return src, dst
			},
			flags:       map[string]string{},
			wantContent: "hello",
		},
		{
			name: "copy with force overwrite",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "src.txt")
				dst := filepath.Join(tmpDir, "dst.txt")
				os.WriteFile(src, []byte("new content"), 0644)
				os.WriteFile(dst, []byte("old content"), 0644)
				return src, dst
			},
			flags:       map[string]string{"force": "true"},
			wantContent: "new content",
		},
		{
			name: "invalid chars in destination",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "src.txt")
				dst := filepath.Join(tmpDir, "dst<file>.txt")
				os.WriteFile(src, []byte("content"), 0644)
				return src, dst
			},
			flags:   map[string]string{},
			wantErr: ErrInvalidChars,
		},
		{
			name: "invalid extension in destination",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "src.txt")
				dst := filepath.Join(tmpDir, "dst.exe")
				os.WriteFile(src, []byte("content"), 0644)
				return src, dst
			},
			flags:   map[string]string{},
			wantErr: ErrInvalidExtension,
		},
		{
			name: "source not found",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "missing.txt")
				dst := filepath.Join(tmpDir, "dst.txt")
				return src, dst
			},
			flags:   map[string]string{},
			wantErr: ErrFileNotFound,
		},
		{
			name: "destination path too long",
			setup: func() (string, string) {
				src := filepath.Join(tmpDir, "src.txt")
				os.WriteFile(src, []byte("content"), 0644)
				longName := strings.Repeat("a", 300) + ".txt"
				dst := filepath.Join(tmpDir, longName)
				return src, dst
			},
			flags:   map[string]string{},
			wantErr: ErrNameTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			src, dst := tt.setup()

			cmd := copyCmd
			cmd.SetArgs([]string{src, dst})
			for k, v := range tt.flags {
				cmd.Flags().Set(k, v)
			}

			if tt.input != "" {
				cmd.SetIn(strings.NewReader(tt.input))
			}

			err := cmd.RunE(cmd, []string{src, dst})
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
	tmpDir := t.TempDir()

	makeCmd := func(stdin string) *cobra.Command {
		cmd := &cobra.Command{
			Use:   "copy [src] [dst]",
			Short: "Copy a file",
			Args:  cobra.ExactArgs(2),
			RunE:  copyCmd.RunE,
		}
		cmd.Flags().BoolP("force", "f", false, "Overwrite destination if exists")
		cmd.SetIn(strings.NewReader(stdin))
		return cmd
	}

	t.Run("overwrite prompt - yes", func(t *testing.T) {
		src := filepath.Join(tmpDir, "src.txt")
		dst := filepath.Join(tmpDir, "dst.txt")
		os.WriteFile(src, []byte("new"), 0644)
		os.WriteFile(dst, []byte("old"), 0644)

		cmd := makeCmd("Y\n")
		cmd.SetArgs([]string{src, dst})

		err := cmd.RunE(cmd, []string{src, dst})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		content, _ := os.ReadFile(dst)
		if string(content) != "new" {
			t.Errorf("expected 'new', got %q", string(content))
		}
	})

	t.Run("overwrite prompt - no", func(t *testing.T) {
		src := filepath.Join(tmpDir, "src2.txt")
		dst := filepath.Join(tmpDir, "dst2.txt")
		os.WriteFile(src, []byte("new"), 0644)
		os.WriteFile(dst, []byte("old"), 0644)

		cmd := makeCmd("N\n")
		cmd.SetArgs([]string{src, dst})

		err := cmd.RunE(cmd, []string{src, dst})
		if err == nil {
			t.Fatal("expected cancellation error")
		}
		if !strings.Contains(err.Error(), "copy cancelled") {
			t.Fatalf("expected copy cancelled, got %q", err.Error())
		}

		content, _ := os.ReadFile(dst)
		if string(content) != "old" {
			t.Errorf("expected 'old' (unchanged), got %q", string(content))
		}
	})
}

func TestCopyCommandPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	tmpDir := t.TempDir()
	restrictedDir := filepath.Join(tmpDir, "restricted")
	os.Mkdir(restrictedDir, 0000)
	defer os.Chmod(restrictedDir, 0700)

	t.Run("read permission denied", func(t *testing.T) {
		src := filepath.Join(restrictedDir, "src.txt")
		dst := filepath.Join(tmpDir, "dst.txt")
		os.WriteFile(src, []byte("content"), 0644)

		cmd := copyCmd
		cmd.SetArgs([]string{src, dst})

		err := cmd.RunE(cmd, []string{src, dst})
		if err == nil {
			t.Fatal("expected permission denied")
		}
		if !strings.Contains(err.Error(), "Permission denied") {
			t.Fatalf("expected permission denied, got %q", err.Error())
		}
	})

	t.Run("write permission denied", func(t *testing.T) {
		src := filepath.Join(tmpDir, "src.txt")
		dst := filepath.Join(restrictedDir, "dst.txt")
		os.WriteFile(src, []byte("content"), 0644)

		cmd := copyCmd
		cmd.SetArgs([]string{src, dst})

		err := cmd.RunE(cmd, []string{src, dst})
		if err == nil {
			t.Fatal("expected permission denied")
		}
		if !strings.Contains(err.Error(), "Permission denied") {
			t.Fatalf("expected permission denied, got %q", err.Error())
		}
	})
}