package main

import (
	"bytes"
	"io"
	"os"
)

type bfCell byte

type bfProgramm struct {
	i           int
	cells       []bfCell
	out         io.Writer
	in          io.Reader
	subProgramm *bfProgramm
	cmdStack    []rune
}

func Execute(sourceInput, input io.Reader, output io.Writer) error {
	var p bytes.Buffer

	_, err := p.ReadFrom(sourceInput)
	if err != nil {
		return err
	}

	bfp := newBFProgramm(input, output)

	for i := range p.Bytes() {
		err = bfp.resolveBFSymbol(rune(p.Bytes()[i]))
		if err != nil {
			return err
		}
	}

	return nil
}

func newBFProgramm(in io.Reader, out io.Writer) bfProgramm {
	return bfProgramm{
		cells:    make([]bfCell, 10),
		out:      out,
		in:       in,
		cmdStack: make([]rune, 0),
	}
}

func (bfp *bfProgramm) resolveBFSymbol(symbol rune) error {
	// ignore new-line symbol.
	if symbol == '\n' {
		return nil
	}

	if bfp.subProgramm != nil {
		return bfp.subProgramm.resolveBFSymbol(symbol)
	}

	bfp.cmdStack = append(bfp.cmdStack, symbol)

	var err error
	switch symbol {
	case '>':
		err = bfp.next()
	case '<':
		err = bfp.prev()
	case '+':
		err = bfp.inc()
	case '-':
		err = bfp.dec()
	case '.':
		err = bfp.print()
	case ',':
		err = bfp.read()
	case '[':
		bfSubProgramm := newBFProgramm(bfp.in, bfp.out)
		bfSubProgramm.cells = bfp.cells
		bfSubProgramm.i = bfp.i
		bfp.subProgramm = &bfSubProgramm
	case ']':
		if bfp.cells[bfp.i] <= 0 {
			bfp = nil
			return nil
		}

		cmdStack := make([]rune, len(bfp.cmdStack))
		copy(cmdStack, (bfp.cmdStack))
		bfp.cmdStack = []rune{}
		for i := range cmdStack {
			bfp.resolveBFSymbol(cmdStack[i])
		}
	default:
		// ignore unknown symbol
		return nil
	}

	return err
}

func (bfp *bfProgramm) next() error {
	bfp.i++
	if bfp.i > len(bfp.cells) {
		bfp.cells = append(bfp.cells, 0)
	}

	return nil
}

func (bfp *bfProgramm) prev() error {
	bfp.i--

	return nil
}

func (bfp *bfProgramm) inc() error {
	bfp.cells[bfp.i]++

	return nil
}

func (bfp *bfProgramm) dec() error {
	bfp.cells[bfp.i]--

	return nil
}

func (bfp *bfProgramm) print() error {
	_, err := bfp.out.Write([]byte{
		byte(bfp.cells[bfp.i]),
	})

	return err
}

func (bfp *bfProgramm) read() error {
	c := make([]byte, 1)
	if _, err := bfp.in.Read(c); err != nil {
		return err
	}

	bfp.cells[bfp.i] = bfCell(c[0])

	return nil
}

// TODO: remove

func main() {
	f, err := os.OpenFile("./examples/hello_world.bf", os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	err = Execute(f, os.Stdin, os.Stdout)
	if err != nil {
		panic(err)
	}
}
