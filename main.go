package main

import (
	"context"
	"os"

	"github.com/MonkeyBuisness/brainfuck-interpreter/bf"
)

// TODO: remove
func main() {
	f, err := os.OpenFile("./examples/factorial.bf", os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	inst, err := bf.Compile(f)
	if err != nil {
		panic(err)
	}

	r := bf.NewRuntime(inst, os.Stdin, os.Stdout)

	//cli.Debug(&r)

	r.Execute(context.Background(), nil)
}
