package main

import (
	"bytes"
	"io"
)

const (
	bfCommandsCount = 8
)

type bfCommandHandler func() error

type bfRuntime struct {
	cells        []rune
	index        int
	inputStream  io.RuneReader
	outputStream io.Writer
	cmdList      []byte
	cmdIndex     int
	loopOffsets  []int
	cmdHandlers  map[byte]bfCommandHandler
}

func bfNewRuntime(in io.RuneReader, out io.Writer) bfRuntime {
	r := bfRuntime{
		cells:        make([]rune, 1),
		index:        0,
		cmdIndex:     0,
		loopOffsets:  make([]int, 0),
		inputStream:  in,
		outputStream: out,
		cmdHandlers:  make(map[byte]bfCommandHandler, bfCommandsCount),
	}

	r.cmdHandlers['>'] = r.next
	r.cmdHandlers['<'] = r.prev
	r.cmdHandlers['+'] = r.inc
	r.cmdHandlers['-'] = r.dec
	r.cmdHandlers['.'] = r.print
	r.cmdHandlers[','] = r.read
	r.cmdHandlers['['] = r.startLoop
	r.cmdHandlers[']'] = r.endLoop

	return r
}

func (r *bfRuntime) execute(cmds []byte) error {
	r.cmdList = cmds

	for r.cmdIndex = 0; r.cmdIndex < len(cmds); r.cmdIndex++ {
		if err := r.executeCurrentCmd(); err != nil {
			return err
		}
	}

	return nil
}

func (r *bfRuntime) executeCurrentCmd() error {
	cmd := r.cmdList[r.cmdIndex]

	h, ok := r.cmdHandlers[cmd]
	if !ok {
		return nil
	}

	return h()
}

func (r *bfRuntime) next() error {
	r.index++
	if r.index == len(r.cells) {
		r.cells = append(r.cells, 0)
	}

	return nil
}

func (r *bfRuntime) prev() error {
	r.index--

	return nil
}

func (r *bfRuntime) inc() error {
	r.cells[r.index]++

	return nil
}

func (r *bfRuntime) dec() error {
	r.cells[r.index]--

	return nil
}

func (r *bfRuntime) print() error {
	_, err := r.outputStream.Write(
		[]byte{byte(r.cells[r.index])},
	)

	return err
}

func (r *bfRuntime) read() error {
	symbol, _, err := r.inputStream.ReadRune()
	if err != nil {
		return err
	}

	r.cells[r.index] = symbol

	return nil
}

func (r *bfRuntime) startLoop() error {
	r.loopOffsets = append(r.loopOffsets, r.cmdIndex)

	return nil
}

func (r *bfRuntime) endLoop() error {
	defer func() {
		r.loopOffsets = r.loopOffsets[:len(r.loopOffsets)-1]
	}()

	if r.cells[r.index] <= 0 {
		return nil
	}

	r.cmdIndex = r.loopOffsets[len(r.loopOffsets)-1]

	return nil
}

// Execute executes brainfuck code.
//
// The sourceInput must provide bf source code reader.
// The input and output have to provide input stream and output stream.
func Execute(sourceInput io.Reader, input io.RuneReader, output io.Writer) error {
	var p bytes.Buffer

	_, err := p.ReadFrom(sourceInput)
	if err != nil {
		return err
	}

	// create runtime.
	rt := bfNewRuntime(input, output)

	return rt.execute(p.Bytes())
}
