package cli

import "errors"

var (
	// ErrValidationFailed indicates the .env file failed validation against the schema.
	ErrValidationFailed = errors.New("validation failed")

	// ErrIO indicates a file I/O or parsing error (schema or .env file).
	ErrIO = errors.New("io error")
)
