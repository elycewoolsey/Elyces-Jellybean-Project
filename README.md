# jellybeans

File operations CLI tool built with Go and Cobra.

## Commands

| Command | Description |
|---------|-------------|
| `create [file]` | Create a file with optional content (`-c`) and mode (`-m`). Overwrite requires `-f`. |
| `copy [src] [dst]` | Copy a file, preserving its mode. Prompts on overwrite unless `-f`. |
| `combine [files...]` | Combine multiple files into one (`-o` output, `-d` delimiter). Atomic via temp + rename. Prompts on overwrite unless `-f`. |
| `delete [files...]` | Delete one or more files. Prompts unless `-f`. Fail-fast on first error. |

Run `jellybeans <command> --help` for full details and examples.

## Flags

### create
- `-c, --content string` - Content to write (default: empty)
- `-f, --force` - Overwrite existing file
- `-m, --mode string` - File permission bits, octal (default: 0644)

### copy
- `-f, --force` - Overwrite destination without prompt

### combine
- `-o, --output string` - Output file (default: combined.txt)
- `-d, --delimiter string` - Delimiter between files (default: none)
- `-f, --force` - Overwrite output without prompt

### delete
- `-f, --force` - Delete without confirmation prompt

## Validation

All commands that write a destination validate:
- **Filename characters**: `< > : " | ? *` are rejected. Every path component is checked, not just the base name (a Windows drive letter like `C:` is allowed).
- **Extension whitelist**: `.txt .md .go .json .yaml .yml .toml .xml .html .css .js .ts .py .rs .java .c .h .cpp .hpp .sh .bat .ps1 .sql .csv` (and no extension). Only the final extension is checked, so `archive.tar.gz` is rejected. Dotfiles like `.gitignore` count as having no extension and are allowed.
- **Path length**: 260 chars on Windows, 4096 on Linux.
- **Windows reserved names**: `CON PRN AUX NUL COM1-9 LPT1-9` are rejected (Windows only).

`combine` validates both its inputs (existence, not-a-directory) and its output path.

## Behavior notes

- **Atomic create**: `create` opens the file with `O_EXCL` when `--force` is not set, so an existing file is rejected without a check-then-write race.
- **Mode preservation**: `copy` gives the destination the source file's permission bits. `create` uses `--mode` (default 0644).
- **Symlinks**: `copy` and `combine` follow symlinks (the target is read). `delete` removes the symlink itself, never the file it points to.
- **Overwrite guards**: `create` (`ErrFileExists`), `copy`, and `combine` all refuse to clobber an existing destination unless `--force` is given (copy and combine prompt first).

## Error Messages

| Error | Meaning |
|-------|---------|
| `invalid character in filename` | Filename/path contains a forbidden character |
| `invalid file extension` | Extension not in the allowed list |
| `reserved filename` | Windows reserved name (e.g. `CON`, `PRN`) |
| `filename too long` | Absolute path exceeds platform limit (260 Windows / 4096 Linux) |
| `file already exists` | `create` target exists without `-f` |
| `file not found` | Source/input file doesn't exist |
| `permission denied` | Read/write permission error |
| `is a directory` | A directory was given where a file was expected |
| `invalid file mode` | `--mode` was not a valid octal value |
| `copy cancelled` | User answered 'N' to the copy overwrite prompt |
| `delete cancelled` | User answered 'N' to the delete confirmation prompt |
| `combine cancelled` | User answered 'N' to the combine overwrite prompt |

## Tests

Run all tests:
```bash
go test ./cmd/... -v
```

With coverage:
```bash
go test ./cmd/... -cover
```

Test coverage per command (subtest counts):
- **create**: 15 tests (empty, content, `-c` short flag, invalid chars, invalid extension, exists, force overwrite, path too long, reserved names [Windows], nonexistent dir, permission denied, `--mode`, invalid mode, extension parsing, dotfile allowed)
- **copy**: 16 tests (basic, force, invalid chars/ext, source not found, dst/src path too long, prompt Y/YES/N/empty/EOF, read & write permission denied, mode preservation, directory source)
- **combine**: 15 tests (2-3 files, delimiters, custom output, not found, directory input, invalid output extension, output too long, atomicity, permission denied, overwrite prompt Y/N/empty/force)
- **delete**: 11 tests (single, multiple, not found, directory, fail-fast, prompt Y/YES/N/empty, permission denied, symlink)

Permission, mode, and symlink tests are skipped on Windows. The Windows reserved-name test is skipped on non-Windows.

## Build

```bash
go build -o jellybeans .
```

## Install Go

Required: Go 1.22+ from https://go.dev/dl/
