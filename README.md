# fileops

File operations CLI tool built with Go and Cobra.

## Commands

| Command | Description |
|---------|-------------|
| `create [file]` | Create a file with optional content (`-c`) |
| `copy [src] [dst]` | Copy a file, prompts on overwrite unless `-f` |
| `combine [files...]` | Combine multiple files into one (`-o` output, `-d` delimiter) |
| `delete [files...]` | Delete one or more files |

## Flags

### create
- `-c, --content string` - Content to write (default: empty)
- `-f, --force` - Overwrite existing file

### copy
- `-f, --force` - Overwrite without prompt

### combine
- `-o, --output string` - Output file (default: combined.txt)
- `-d, --delimiter string` - Delimiter between files (default: none)

### delete
- `-f, --force` - Force deletion (currently unused)

## Error Messages

- `Not a valid character` - Invalid filename characters
- `invalid file extension` - Extension not in allowed list
- `filename too long` - Path exceeds 260 chars (Windows) / 4096 (Linux)
- `file already exists` - Create target exists without `-f`
- `File not found` - Source/file doesn't exist
- `Permission denied` - Read/write permission error
- `is a directory` - Attempted to delete a directory
- `copy cancelled` - User answered 'N' to overwrite prompt

## Tests

Run all tests:
```bash
go test ./cmd/... -v
```

Test coverage per command:
- **create**: 11 tests (valid, invalid chars, ext, exists, force, long name, reserved names)
- **delete**: 4 tests (single, multiple, not found, directory)
- **combine**: 6 tests (2-3 files, delimiters, custom output, not found)
- **copy**: 6 tests (basic, force, invalid chars/ext, not found, long path)
- **copy prompt**: 2 tests (Y=overwrite, N=cancel)
- **permissions**: Skipped on Windows

## Build

```bash
go build -o fileops main.go
```

## Install Go

Required: Go 1.21+ from https://go.dev/dl/