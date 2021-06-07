package bf

import (
	"errors"
	"fmt"
)

// Error represents typedef for the Brainfuck error.
type Error error

// Brainfuck error.
var (
	ErrReadSymbol  Error = errors.New("could not read symbol")
	ErrWriteSymbol Error = errors.New("could not write symbol")
	ErrCompilation Error = errors.New("could not compile code")
)

// NewError returns new error instance.
func NewError(inErr Error, err error) error {
	return fmt.Errorf("%w: %v", inErr, err)
}
