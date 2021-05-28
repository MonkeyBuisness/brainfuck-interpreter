package main

import (
	"os"
	"testing"
)

func Test_Execute(t *testing.T) {
	f, err := os.OpenFile("./examples/hello_world.bf", os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	err = Execute(f, os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}
}
