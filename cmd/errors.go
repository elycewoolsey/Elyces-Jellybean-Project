package cmd

import (
	"errors"
	"os"
)

// Sentinel errors returned by the commands. They are compared with errors.Is
// and carry the user-facing message printed by the CLI.
var (
	// ErrInvalidChar means the filename or a path component contains a
	// forbidden character (one of < > : " | ? *).
	ErrInvalidChar = errors.New("invalid character in filename")
	// ErrInvalidExtension means the file extension is not in the allowed list.
	ErrInvalidExtension = errors.New("invalid file extension")
	// ErrReservedName means the name is a reserved device name on Windows
	// (e.g. CON, PRN, NUL).
	ErrReservedName = errors.New("reserved filename")
	// ErrFileExists means a create/combine target already exists and --force
	// was not given.
	ErrFileExists = errors.New("file already exists")
	// ErrNameTooLong means the absolute path exceeds the platform limit.
	ErrNameTooLong = errors.New("filename too long")
	// ErrPermissionDenied means a read or write was rejected by the OS.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrFileNotFound means a source or input file does not exist.
	ErrFileNotFound = errors.New("file not found")
	// ErrIsDirectory means a directory was given where a file was expected.
	ErrIsDirectory = errors.New("is a directory")
	// ErrInvalidMode means the --mode flag was not a valid octal value.
	ErrInvalidMode = errors.New("invalid file mode")
	// ErrCopyCancelled means the user declined the copy overwrite prompt.
	ErrCopyCancelled = errors.New("copy cancelled")
	// ErrDeleteCancelled means the user declined the delete confirmation.
	ErrDeleteCancelled = errors.New("delete cancelled")
	// ErrCombineCancelled means the user declined the combine overwrite prompt.
	ErrCombineCancelled = errors.New("combine cancelled")
)

// mapOSError converts common os errors into our sentinel errors so callers
// get consistent messages. Other errors are returned unchanged.
func mapOSError(err error) error {
	switch {
	case err == nil:
		return nil
	case os.IsNotExist(err):
		return ErrFileNotFound
	case os.IsPermission(err):
		return ErrPermissionDenied
	default:
		return err
	}
}
