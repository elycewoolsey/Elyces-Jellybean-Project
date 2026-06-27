package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCombineCommand(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(t *testing.T) (files []string, output string)
		flags       map[string]string
		wantContent string
		wantErr     error
		skipCheck   bool
	}{
		{
			name: "combine two files",
			setup: func(t *testing.T) ([]string, string) {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "a.txt")
				f2 := filepath.Join(dir, "b.txt")
				if err := os.WriteFile(f1, []byte("first"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f2, []byte("second"), 0644); err != nil {
					t.Fatal(err)
				}
				return []string{f1, f2}, filepath.Join(dir, "out.txt")
			},
			flags:       map[string]string{},
			wantContent: "firstsecond",
		},
		{
			name: "combine with delimiter",
			setup: func(t *testing.T) ([]string, string) {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "a.txt")
				f2 := filepath.Join(dir, "b.txt")
				if err := os.WriteFile(f1, []byte("first"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f2, []byte("second"), 0644); err != nil {
					t.Fatal(err)
				}
				return []string{f1, f2}, filepath.Join(dir, "out.txt")
			},
			flags:       map[string]string{"delimiter": ","},
			wantContent: "first,second",
		},
		{
			name: "combine with newline delimiter",
			setup: func(t *testing.T) ([]string, string) {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "a.txt")
				f2 := filepath.Join(dir, "b.txt")
				if err := os.WriteFile(f1, []byte("first"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f2, []byte("second"), 0644); err != nil {
					t.Fatal(err)
				}
				return []string{f1, f2}, filepath.Join(dir, "out.txt")
			},
			flags:       map[string]string{"delimiter": "\n"},
			wantContent: "first\nsecond",
		},
		{
			name: "combine three files",
			setup: func(t *testing.T) ([]string, string) {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "a.txt")
				f2 := filepath.Join(dir, "b.txt")
				f3 := filepath.Join(dir, "c.txt")
				if err := os.WriteFile(f1, []byte("1"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f2, []byte("2"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f3, []byte("3"), 0644); err != nil {
					t.Fatal(err)
				}
				return []string{f1, f2, f3}, filepath.Join(dir, "out.txt")
			},
			flags:       map[string]string{"delimiter": "-"},
			wantContent: "1-2-3",
		},
		{
			name: "output to custom path",
			setup: func(t *testing.T) ([]string, string) {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "a.txt")
				f2 := filepath.Join(dir, "b.txt")
				if err := os.WriteFile(f1, []byte("x"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f2, []byte("y"), 0644); err != nil {
					t.Fatal(err)
				}
				customDir := filepath.Join(dir, "custom")
				if err := os.MkdirAll(customDir, 0755); err != nil {
					t.Fatal(err)
				}
				return []string{f1, f2}, filepath.Join(customDir, "out.txt")
			},
			flags:       map[string]string{},
			wantContent: "xy",
		},
		{
			name: "file not found",
			setup: func(t *testing.T) ([]string, string) {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "a.txt")
				if err := os.WriteFile(f1, []byte("x"), 0644); err != nil {
					t.Fatal(err)
				}
				return []string{f1, filepath.Join(dir, "missing.txt")}, filepath.Join(dir, "out.txt")
			},
			flags:     map[string]string{},
			wantErr:   ErrFileNotFound,
			skipCheck: true,
		},
		{
			name: "directory input returns error",
			setup: func(t *testing.T) ([]string, string) {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "a.txt")
				if err := os.WriteFile(f1, []byte("x"), 0644); err != nil {
					t.Fatal(err)
				}
				d := filepath.Join(dir, "subdir")
				if err := os.Mkdir(d, 0755); err != nil {
					t.Fatal(err)
				}
				return []string{f1, d}, filepath.Join(dir, "out.txt")
			},
			flags:     map[string]string{},
			wantErr:   ErrIsDirectory,
			skipCheck: true,
		},
		{
			name: "invalid output extension",
			setup: func(t *testing.T) ([]string, string) {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "a.txt")
				f2 := filepath.Join(dir, "b.txt")
				if err := os.WriteFile(f1, []byte("x"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f2, []byte("y"), 0644); err != nil {
					t.Fatal(err)
				}
				return []string{f1, f2}, filepath.Join(dir, "out.exe")
			},
			flags:     map[string]string{},
			wantErr:   ErrInvalidExtension,
			skipCheck: true,
		},
		{
			name: "output path too long",
			setup: func(t *testing.T) ([]string, string) {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "a.txt")
				f2 := filepath.Join(dir, "b.txt")
				if err := os.WriteFile(f1, []byte("x"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f2, []byte("y"), 0644); err != nil {
					t.Fatal(err)
				}
				return []string{f1, f2}, filepath.Join(dir, strings.Repeat("a", 5000)+".txt")
			},
			flags:     map[string]string{},
			wantErr:   ErrNameTooLong,
			skipCheck: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, output := tt.setup(t)

			flags := map[string]string{"output": output}
			for k, v := range tt.flags {
				flags[k] = v
			}

			c := newCombineCmd()
			c.SetArgs(files)
			setFlags(t, c, flags)

			err := c.RunE(c, files)
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

			if tt.skipCheck {
				return
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

func TestCombineCommandAtomic(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "out.txt")
	// Pre-existing output that must be preserved on failure.
	if err := os.WriteFile(out, []byte("original"), 0644); err != nil {
		t.Fatal(err)
	}

	f1 := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(f1, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	missing := filepath.Join(dir, "missing.txt")

	c := newCombineCmd()
	c.SetArgs([]string{f1, missing})
	setFlags(t, c, map[string]string{"output": out})

	err := c.RunE(c, []string{f1, missing})
	if !errors.Is(err, ErrFileNotFound) {
		t.Fatalf("expected ErrFileNotFound, got %v", err)
	}

	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if string(content) != "original" {
		t.Errorf("output changed on failure = %q, want %q (atomic)", string(content), "original")
	}
}

func TestCombineCommandPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.txt")
	f2 := filepath.Join(dir, "b.txt")
	if err := os.WriteFile(f1, []byte("x"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("y"), 0644); err != nil {
		t.Fatal(err)
	}

	restricted := filepath.Join(dir, "restricted")
	if err := os.Mkdir(restricted, 0555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(restricted, 0755)

	out := filepath.Join(restricted, "out.txt")
	c := newCombineCmd()
	c.SetArgs([]string{f1, f2})
	setFlags(t, c, map[string]string{"output": out})

	err := c.RunE(c, []string{f1, f2})
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestCombineCommandOverwrite(t *testing.T) {
	tests := []struct {
		name        string
		flags       map[string]string
		input       string
		wantErr     error
		wantContent string
	}{
		{name: "prompt yes overwrites", input: "Y\n", wantContent: "ab"},
		{name: "prompt no cancels", input: "N\n", wantErr: ErrCombineCancelled, wantContent: "original"},
		{name: "empty input cancels", input: "", wantErr: ErrCombineCancelled, wantContent: "original"},
		{name: "force overwrites", flags: map[string]string{"force": "true"}, wantContent: "ab"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			f1 := filepath.Join(dir, "a.txt")
			f2 := filepath.Join(dir, "b.txt")
			out := filepath.Join(dir, "out.txt")
			for p, c := range map[string]string{f1: "a", f2: "b", out: "original"} {
				if err := os.WriteFile(p, []byte(c), 0644); err != nil {
					t.Fatal(err)
				}
			}

			c := newCombineCmd()
			flags := map[string]string{"output": out}
			for k, v := range tt.flags {
				flags[k] = v
			}
			c.SetArgs([]string{f1, f2})
			setFlags(t, c, flags)
			runWithStdin(c, tt.input)

			err := c.RunE(c, []string{f1, f2})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			content, _ := os.ReadFile(out)
			if string(content) != tt.wantContent {
				t.Errorf("content = %q, want %q", string(content), tt.wantContent)
			}
		})
	}
}
