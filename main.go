package main

import (
	"os"

	"github.com/MonkeyBuisness/brainfuck-interpreter/bf"
	"github.com/MonkeyBuisness/brainfuck-interpreter/cli"
)

// TODO: remove
func main() {
	f, err := os.OpenFile("./examples/factorial.bf1", os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	inst, err := bf.Compile(f)
	if err != nil {
		panic(err)
	}

	r := bf.NewRuntime(inst, os.Stdin, os.Stdout)

	cli.Debug(&r)

	//r.Execute(context.Background(), nil)
}
