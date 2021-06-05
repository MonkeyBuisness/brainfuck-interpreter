package main

import (
	"errors"
	"fmt"
)

// BFError represents typedef for the Brainfuck error.
type BFError error

// Brainfuck error.
var (
	ReadSymbolError  BFError = errors.New("could not read symbol")
	WriteSymbolError BFError = errors.New("could not write symbol")
	CompilationError BFError = errors.New("could not compile code")
)

// NewError returns new error instance.
func NewError(inErr BFError, err error) error {
	return fmt.Errorf("%w: %v", inErr, err)
}
