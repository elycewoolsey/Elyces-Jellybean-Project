package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCreateCommand(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name     string
		args     []string
		flags    map[string]string
		wantFile string
		wantErr  bool
	}{
		{
			name:     "create empty file",
			args:     []string{filepath.Join(tmpDir, "empty.txt")},
			flags:    map[string]string{},
			wantFile: "",
		},
		{
			name:     "create file with content",
			args:     []string{filepath.Join(tmpDir, "hello.txt")},
			flags:    map[string]string{"content": "hello world"},
			wantFile: "hello world",
		},
		{
			name:     "create with short flag",
			args:     []string{filepath.Join(tmpDir, "short.txt")},
			flags:    map[string]string{"content": "short flag"},
			wantFile: "short flag",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := createCmd
			cmd.SetArgs(tt.args)
			for k, v := range tt.flags {
				cmd.Flags().Set(k, v)
			}

			err := cmd.RunE(cmd, tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("RunE() error = %v, wantErr %v", err, tt.wantErr)
			}

			content, err := os.ReadFile(tt.args[0])
			if err != nil {
				t.Fatalf("ReadFile: %v", err)
			}
			if string(content) != tt.wantFile {
				t.Errorf("content = %q, want %q", string(content), tt.wantFile)
			}
		})
	}
}