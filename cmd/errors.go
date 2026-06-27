package cmd

import "errors"

var (
	ErrInvalidChars     = errors.New("Not a valid character")
	ErrInvalidExtension = errors.New("invalid file extension")
	ErrFileExists       = errors.New("file already exists")
	ErrNameTooLong      = errors.New("filename too long")
	ErrPermissionDenied = errors.New("Permission denied")
	ErrFileNotFound     = errors.New("File not found")
	ErrIsDirectory      = errors.New("is a directory")
)