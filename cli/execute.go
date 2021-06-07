package cli

import (
	"context"
	"io"
	"os"

	"github.com/MonkeyBuisness/brainfuck-interpreter/bf"
)

// Execute represents cli command for executing Brainfuck code.
func Execute(ctx context.Context, in io.Reader, out io.Writer) error {
	instructions, err := bf.Compile(in)
	if err != nil {
		return err
	}

	r := bf.NewRuntime(instructions, os.Stdin, out)
	return r.Execute(ctx, nil)
}
