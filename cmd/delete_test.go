package cmd

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestDeleteCommand(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) []string
		flags      map[string]string
		input      string
		wantErr    error
		checkFiles func(t *testing.T, files []string)
	}{
		{
			name: "delete single file with force",
			setup: func(t *testing.T) []string {
				f := filepath.Join(t.TempDir(), "test.txt")
				if err := os.WriteFile(f, []byte("content"), 0644); err != nil {
					t.Fatal(err)
				}
				return []string{f}
			},
			flags: map[string]string{"force": "true"},
			checkFiles: func(t *testing.T, files []string) {
				if _, err := os.Stat(files[0]); !os.IsNotExist(err) {
					t.Error("file should be deleted")
				}
			},
		},
		{
			name: "delete multiple files with force",
			setup: func(t *testing.T) []string {
				dir := t.TempDir()
				f1 := filepath.Join(dir, "test1.txt")
				f2 := filepath.Join(dir, "test2.txt")
				if err := os.WriteFile(f1, []byte("1"), 0644); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(f2, []byte("2"), 0644); err != nil {
					t.Fatal(err)
				}
				return []string{f1, f2}
			},
			flags: map[string]string{"force": "true"},
			checkFiles: func(t *testing.T, files []string) {
				for _, f := range files {
					if _, err := os.Stat(f); !os.IsNotExist(err) {
						t.Errorf("file %q should be deleted", f)
					}
				}
			},
		},
		{
			name: "file not found",
			setup: func(t *testing.T) []string {
				return []string{filepath.Join(t.TempDir(), "nonexistent.txt")}
			},
			flags:   map[string]string{"force": "true"},
			wantErr: ErrFileNotFound,
		},
		{
			name: "directory returns error",
			setup: func(t *testing.T) []string {
				d := filepath.Join(t.TempDir(), "dir")
				if err := os.Mkdir(d, 0755); err != nil {
					t.Fatal(err)
				}
				return []string{d}
			},
			flags:   map[string]string{"force": "true"},
			wantErr: ErrIsDirectory,
			checkFiles: func(t *testing.T, files []string) {
				if _, err := os.Stat(files[0]); err != nil {
					t.Error("directory should still exist")
				}
			},
		},
		{
			name: "fail-fast stops at first missing file",
			setup: func(t *testing.T) []string {
				dir := t.TempDir()
				exists := filepath.Join(dir, "exists.txt")
				if err := os.WriteFile(exists, []byte("x"), 0644); err != nil {
					t.Fatal(err)
				}
				missing := filepath.Join(dir, "missing.txt")
				return []string{missing, exists}
			},
			flags:   map[string]string{"force": "true"},
			wantErr: ErrFileNotFound,
			checkFiles: func(t *testing.T, files []string) {
				// The second file must still be present (fail-fast, not batch).
				if _, err := os.Stat(files[1]); err != nil {
					t.Error("second file should still exist after fail-fast on first")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files := tt.setup(t)

			c := newDeleteCmd()
			c.SetArgs(files)
			setFlags(t, c, tt.flags)
			runWithStdin(c, tt.input)

			err := c.RunE(c, files)
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if tt.checkFiles != nil {
				tt.checkFiles(t, files)
			}
		})
	}
}

func TestDeleteCommandPrompt(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
		deleted bool
	}{
		{name: "yes deletes", input: "Y\n", deleted: true},
		{name: "YES deletes", input: "YES\n", deleted: true},
		{name: "no cancels", input: "N\n", wantErr: ErrDeleteCancelled, deleted: false},
		{name: "empty cancels", input: "", wantErr: ErrDeleteCancelled, deleted: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := filepath.Join(t.TempDir(), "victim.txt")
			if err := os.WriteFile(f, []byte("x"), 0644); err != nil {
				t.Fatal(err)
			}

			c := newDeleteCmd()
			c.SetArgs([]string{f})
			runWithStdin(c, tt.input)

			err := c.RunE(c, []string{f})
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected %v, got %v", tt.wantErr, err)
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			_, statErr := os.Stat(f)
			if tt.deleted {
				if !os.IsNotExist(statErr) {
					t.Error("file should be deleted")
				}
			} else {
				if statErr != nil {
					t.Error("file should still exist")
				}
			}
		})
	}
}

func TestDeleteCommandPermissionDenied(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping permission test on Windows")
	}

	dir := t.TempDir()
	restricted := filepath.Join(dir, "restricted")
	if err := os.Mkdir(restricted, 0755); err != nil {
		t.Fatal(err)
	}
	// Create the file while the directory is writable.
	f := filepath.Join(restricted, "test.txt")
	if err := os.WriteFile(f, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}
	// Lock down the directory so removal is denied.
	if err := os.Chmod(restricted, 0555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(restricted, 0755)

	c := newDeleteCmd()
	c.SetArgs([]string{f})
	setFlags(t, c, map[string]string{"force": "true"})

	err := c.RunE(c, []string{f})
	if !errors.Is(err, ErrPermissionDenied) {
		t.Fatalf("expected ErrPermissionDenied, got %v", err)
	}
}

func TestDeleteSymlinkRemovesLink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping symlink test on Windows")
	}

	dir := t.TempDir()
	target := filepath.Join(dir, "target")
	if err := os.Mkdir(target, 0755); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(dir, "link")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	// Deleting a symlink to a directory must remove the link, not refuse it
	// as "is a directory", and must leave the target intact.
	c := newDeleteCmd()
	c.SetArgs([]string{link})
	setFlags(t, c, map[string]string{"force": "true"})

	if err := c.RunE(c, []string{link}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Lstat(link); !os.IsNotExist(err) {
		t.Error("symlink should be removed")
	}
	if _, err := os.Stat(target); err != nil {
		t.Error("target directory should still exist")
	}
}
