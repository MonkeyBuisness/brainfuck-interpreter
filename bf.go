package main

import (
	"bufio"
	"bytes"
	"io"
	"os"
)

type bfRuntime struct {
	cells        []rune
	index        int
	inputStream  io.RuneReader
	outputStream io.Writer
}

type bfExpression struct {
	cmdList  []rune
	runtime  *bfRuntime
	innerExp *bfExpression
	isDone   bool
}

func bfNewRuntime(in io.RuneReader, out io.Writer) bfRuntime {
	return bfRuntime{
		cells:        make([]rune, 1),
		index:        0,
		inputStream:  in,
		outputStream: out,
	}
}

func bfNewExpression(runtime *bfRuntime) *bfExpression {
	return &bfExpression{
		cmdList: make([]rune, 0),
		runtime: runtime,
	}
}

func (e *bfExpression) applyCommand(cmd rune) error {
	if e.innerExp != nil && e.innerExp.isDone {
		e.innerExp = nil
	}

	if e.innerExp != nil {
		return e.innerExp.applyCommand(cmd)
	}

	var err error
	switch cmd {
	case '>':
		err = e.next()
	case '<':
		err = e.prev()
	case '+':
		err = e.inc()
	case '-':
		err = e.dec()
	case '.':
		err = e.print()
	case ',':
		err = e.read()
	case '[':
		e.innerExp = bfNewExpression(e.runtime)
		return nil
	case ']':
		for e.runtime.cells[e.runtime.index] > 0 {
			err = e.run()
			if err != nil {
				return err
			}
		}
		e.isDone = true
		return nil
	default:
		// ignore unknown symbol
		return nil
	}

	if err == nil {
		e.cmdList = append(e.cmdList, cmd)
	}

	return err
}

func (e *bfExpression) run() error {
	cmdList := make([]rune, len(e.cmdList))
	copy(cmdList, e.cmdList)
	e.cmdList = []rune{}

	for i := range cmdList {
		if err := e.applyCommand(cmdList[i]); err != nil {
			return err
		}
	}

	return nil
}

func (e *bfExpression) next() error {
	e.runtime.index++
	if e.runtime.index >= len(e.runtime.cells) {
		e.runtime.cells = append(e.runtime.cells, 0)
	}

	return nil
}

func (e *bfExpression) prev() error {
	e.runtime.index--

	return nil
}

func (e *bfExpression) inc() error {
	e.runtime.cells[e.runtime.index]++

	return nil
}

func (e *bfExpression) dec() error {
	e.runtime.cells[e.runtime.index]--

	return nil
}

func (e *bfExpression) print() error {
	_, err := e.runtime.outputStream.Write(
		[]byte{byte(e.runtime.cells[e.runtime.index])},
	)

	return err
}

func (e *bfExpression) read() error {
	r, _, err := e.runtime.inputStream.ReadRune()
	if err != nil {
		return err
	}

	e.runtime.cells[e.runtime.index] = r

	return nil
}

func Execute(sourceInput io.Reader, input io.RuneReader, output io.Writer) error {
	var p bytes.Buffer

	_, err := p.ReadFrom(sourceInput)
	if err != nil {
		return err
	}

	// create runtime.
	rt := bfNewRuntime(input, output)

	// create default expression.
	bfExp := bfNewExpression(&rt)

	for i := range p.Bytes() {
		///
		//fmt.Printf("%c", p.Bytes()[i])
		///
		err = bfExp.applyCommand(rune(p.Bytes()[i]))
		if err != nil {
			return err
		}
	}

	return nil
}

// TODO: remove

func main() {
	f, err := os.OpenFile("./examples/simple.bf", os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	err = Execute(f, bufio.NewReader(os.Stdin), os.Stdout)
	if err != nil {
		panic(err)
	}
}
