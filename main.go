package main

import (
	"fmt"
	"os"

	bfCli "github.com/MonkeyBuisness/brainfuck-interpreter/cli"
	"github.com/urfave/cli/v2"
)

func main() {
	err := (&cli.App{
		Name:  "Brainfuck interpreter",
		Usage: "run your Brainfuck code",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"in", "if"},
				Usage:   "input file (stdin by default)",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"out", "of"},
				Usage:   "output file (stdout by default)",
			},
			&cli.BoolFlag{
				Name:    "debug",
				Aliases: []string{"dbg", "d"},
				Usage:   "execute Brainfuck code in debug mode",
			},
		},
		Action: func(c *cli.Context) error {
			in := os.Stdin
			out := os.Stdout
			if inputFile := c.String("input"); inputFile != "" {
				f, err := os.OpenFile(inputFile, os.O_RDONLY, 0666)
				if err != nil {
					return err
				}

				in = f
			}

			if outputFile := c.String("output"); outputFile != "" {
				f, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0666)
				if err != nil {
					return err
				}

				out = f
			}

			if err := bfCli.Execute(c.Context, in, out); err != nil {
				return fmt.Errorf("could not execute code: %v", err)
			}

			if err := in.Close(); err != nil {
				return fmt.Errorf("could not close input reader: %v", err)
			}

			if err := out.Close(); err != nil {
				return fmt.Errorf("could not close output reader: %v", err)
			}

			return nil
		},
	}).Run(os.Args)

	if err != nil {
		panic(err)
	}
}
