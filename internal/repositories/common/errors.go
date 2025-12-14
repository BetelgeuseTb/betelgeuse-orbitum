package common

import "errors"

var (
	ErrNotFound     = errors.New("entity not found")
	ErrConflict     = errors.New("entity conflict")
	ErrInvalidInput = errors.New("invalid input")
)
