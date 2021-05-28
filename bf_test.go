package main

import (
	"bufio"
	"os"
	"testing"
)

func Test_Execute(t *testing.T) {
	f, err := os.OpenFile("./examples/simple.bf", os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	err = Execute(f, bufio.NewReader(os.Stdin), bufio.NewWriter(os.Stdout))
	if err != nil {
		panic(err)
	}
}
