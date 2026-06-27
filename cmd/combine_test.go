package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCombineCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() ([]string, string)
		flags       map[string]string
		wantContent string
		wantErr     error
	}{
		{
			name: "combine two files",
			setup: func() ([]string, string) {
				f1 := filepath.Join(tmpDir, "a.txt")
				f2 := filepath.Join(tmpDir, "b.txt")
				os.WriteFile(f1, []byte("first"), 0644)
				os.WriteFile(f2, []byte("second"), 0644)
				out := filepath.Join(tmpDir, "out.txt")
				return []string{f1, f2}, out
			},
			flags:       map[string]string{},
			wantContent: "firstsecond",
		},
		{
			name: "combine with delimiter",
			setup: func() ([]string, string) {
				f1 := filepath.Join(tmpDir, "a.txt")
				f2 := filepath.Join(tmpDir, "b.txt")
				os.WriteFile(f1, []byte("first"), 0644)
				os.WriteFile(f2, []byte("second"), 0644)
				out := filepath.Join(tmpDir, "out.txt")
				return []string{f1, f2}, out
			},
			flags:       map[string]string{"delimiter": ","},
			wantContent: "first,second",
		},
		{
			name: "combine with newline delimiter",
			setup: func() ([]string, string) {
				f1 := filepath.Join(tmpDir, "a.txt")
				f2 := filepath.Join(tmpDir, "b.txt")
				os.WriteFile(f1, []byte("first"), 0644)
				os.WriteFile(f2, []byte("second"), 0644)
				out := filepath.Join(tmpDir, "out.txt")
				return []string{f1, f2}, out
			},
			flags:       map[string]string{"delimiter": "\n"},
			wantContent: "first\nsecond",
		},
		{
			name: "combine three files",
			setup: func() ([]string, string) {
				f1 := filepath.Join(tmpDir, "a.txt")
				f2 := filepath.Join(tmpDir, "b.txt")
				f3 := filepath.Join(tmpDir, "c.txt")
				os.WriteFile(f1, []byte("1"), 0644)
				os.WriteFile(f2, []byte("2"), 0644)
				os.WriteFile(f3, []byte("3"), 0644)
				out := filepath.Join(tmpDir, "out.txt")
				return []string{f1, f2, f3}, out
			},
			flags:       map[string]string{"delimiter": "-"},
			wantContent: "1-2-3",
		},
		{
			name: "output to custom path",
			setup: func() ([]string, string) {
				f1 := filepath.Join(tmpDir, "a.txt")
				f2 := filepath.Join(tmpDir, "b.txt")
				os.WriteFile(f1, []byte("x"), 0644)
				os.WriteFile(f2, []byte("y"), 0644)
				customDir := filepath.Join(tmpDir, "custom")
				os.MkdirAll(customDir, 0755)
				out := filepath.Join(customDir, "out.txt")
				return []string{f1, f2}, out
			},
			flags:       map[string]string{},
			wantContent: "xy",
		},
		{
			name: "file not found",
			setup: func() ([]string, string) {
				f1 := filepath.Join(tmpDir, "a.txt")
				os.WriteFile(f1, []byte("x"), 0644)
				out := filepath.Join(tmpDir, "out.txt")
				return []string{f1, filepath.Join(tmpDir, "missing.txt")}, out
			},
			flags:   map[string]string{},
			wantErr: ErrFileNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, output := tt.setup()
			flags := make(map[string]string)
			for k, v := range tt.flags {
				flags[k] = v
			}
			if flags["output"] == "" {
				flags["output"] = output
			}

			cmd := combineCmd
			cmd.SetArgs(files)
			cmd.Flags().Set("delimiter", "")  // reset
			for k, v := range flags {
				cmd.Flags().Set(k, v)
			}

			err := cmd.RunE(cmd, files)
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

			content, err := os.ReadFile(output)
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			if string(content) != tt.wantContent {
				t.Errorf("content = %q, want %q", string(content), tt.wantContent)
			}
		})
	}
}