package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"

	tm "github.com/buger/goterm"
)

// BFRuntime represents Brainfuck runtime instance.
//
// This instance is responsible for executing list of the provided
// Brainfuck commands and managing memory cells via execution process.
type BFRuntime struct {
	cells        []byte
	index        int
	instructions []BFInstruction
	instIndex    int
}

// BFInstruction represents execution interface of a single Brainfuck instruction
// on the runtime process.
//
// In other words it represents command handler for a single Brainfuck operator
// such as +, -, <, >, etc.
type BFInstruction interface {
	Execute(i int, runtime *BFRuntime) error
	Cmd() rune
}

type (
	BFInstructionMoveNextCell struct{}
	BFInstructionMovePrevCell struct{}
	BFInstructionIncValue     struct{}
	BFInstructionDecValue     struct{}
	BFInstructionStartLoop    struct {
		endLoopIndex int
	}
	BFInstructionEndLoop struct {
		startLoopIndex int
	}
	BFInstructionPrint struct{}
	BFInstructionRead  struct{}
)

// Value returns value of a current cell.
func (r *BFRuntime) Value() byte {
	return r.cells[r.index]
}

// Pointer returns current's cell index.
func (r *BFRuntime) Pointer() int {
	return r.index
}

// Next moves pointer to the next cell.
func (r *BFRuntime) Next() {
	r.index++

	if r.index == len(r.cells) {
		r.cells = append(r.cells, 0)
	}
}

// Prev moves pointer to the previous cell.
func (r *BFRuntime) Prev() {
	r.index--
}

// Inc increments current's cell value.
func (r *BFRuntime) Inc() {
	r.cells[r.index]++
}

// Dec decrements current's cell value.
func (r *BFRuntime) Dec() {
	r.cells[r.index]--
}

// Jump sets instruction index to execute.
func (r *BFRuntime) Jump(i int) {
	r.instIndex = i
}

// Snapshot returns runtime's cell values as a byte slice.
func (r *BFRuntime) Snapshot() []byte {
	cp := make([]byte, len(r.cells))
	copy(cp, r.cells)

	return cp
}

// Instruction returns current instruction to execute and its index.
func (r *BFRuntime) Instruction() (BFInstruction, int) {
	return r.instructions[r.instIndex], r.instIndex
}

// Execute starts runtime process.
//
// waitChan (<-chan struct{}) param can be used to debug or pause execution process.
// After each instruction executed this channel will be checked for nil (or closed).
// If it's not nil, then runtime will wait for chanel's value and continue execution.
// If it's closed (or nil), then runtime skip channel reading and continue execution.
// Most of the time you can pass nil as a waitChan value.
func (r *BFRuntime) Execute(ctx context.Context, waitChan <-chan struct{}) error {
	for r.instIndex = 0; r.instIndex < len(r.instructions); r.instIndex++ {
		if waitChan != nil {
			<-waitChan
		}
		if err := r.instructions[r.instIndex].Execute(r.instIndex, r); err != nil {
			panic(err)
		}
	}

	return nil
}

func Compile(sourceInput io.Reader) ([]BFInstruction, error) {
	var p bytes.Buffer
	_, err := p.ReadFrom(sourceInput)
	if err != nil {
		return nil, err
	}

	instructions := make([]BFInstruction, 0, p.Len())
	loopOffsets := make([]int, 0)

	for i := range p.Bytes() {
		var instruction BFInstruction
		switch p.Bytes()[i] {
		case '>':
			instruction = &BFInstructionMoveNextCell{}
		case '<':
			instruction = &BFInstructionMovePrevCell{}
		case '+':
			instruction = &BFInstructionIncValue{}
		case '-':
			instruction = &BFInstructionDecValue{}
		case '.':
			instruction = &BFInstructionPrint{}
		case '[':
			instruction = &BFInstructionStartLoop{}
			loopOffsets = append(loopOffsets, len(instructions))
		case ']':
			endLoopInstruction := &BFInstructionEndLoop{}
			endLoopInstruction.startLoopIndex = loopOffsets[len(loopOffsets)-1]
			loopOffsets = loopOffsets[:len(loopOffsets)-1]

			if startLoopInstruction, ok := instructions[endLoopInstruction.startLoopIndex].(*BFInstructionStartLoop); ok {
				startLoopInstruction.endLoopIndex = len(instructions)
			}

			instruction = endLoopInstruction
		}

		if instruction != nil {
			instructions = append(instructions, instruction)
		}
	}

	return instructions, nil
}

func (i *BFInstructionMoveNextCell) Execute(index int, runtime *BFRuntime) error {
	runtime.NextCell()
	return nil
}

func (i *BFInstructionMovePrevCell) Execute(index int, runtime *BFRuntime) error {
	runtime.PrevCell()
	return nil
}

func (i *BFInstructionIncValue) Execute(index int, runtime *BFRuntime) error {
	runtime.Inc()
	return nil
}

func (i *BFInstructionDecValue) Execute(index int, runtime *BFRuntime) error {
	runtime.Dec()
	return nil
}

func (i *BFInstructionStartLoop) Execute(index int, runtime *BFRuntime) error {
	if runtime.Value() == 0 {
		runtime.MoveInstructionIndex(i.endLoopIndex)
	}

	return nil
}

func (i *BFInstructionEndLoop) Execute(index int, runtime *BFRuntime) error {
	if runtime.Value() > 0 {
		runtime.MoveInstructionIndex(i.startLoopIndex)
	}
	return nil
}

func (i *BFInstructionPrint) Execute(index int, runtime *BFRuntime) error {
	fmt.Printf("%c", runtime.Value())
	return nil
}

func (i *BFInstructionMoveNextCell) String() string {
	return ">"
}

func (i *BFInstructionMovePrevCell) String() string {
	return "<"
}

func (i *BFInstructionIncValue) String() string {
	return "+"
}

func (i *BFInstructionDecValue) String() string {
	return "-"
}

func (i *BFInstructionStartLoop) String() string {
	return "["
}

func (i *BFInstructionEndLoop) String() string {
	return "]"
}

func (i *BFInstructionPrint) String() string {
	return "."
}

func (r *BFRuntime) Debug() {
	waitChan := make(chan struct{}, 1)
	go r.Execute(waitChan)

	tm.Clear()
	for {
		///
		tm.Clear()
		tm.MoveCursor(1, 1)
		for i := range r.instructions {
			if i == r.instIndex {
				tm.Print(" |")
				tm.Print(tm.Color(r.instructions[i].String(), tm.RED))
				tm.Print("| ")
				continue
			}

			tm.Print(r.instructions[i])
		}

		tm.Print("\n\nCELLS:\n\n")
		for i := range r.cells {
			cellStr := fmt.Sprintf("[%d]: %d\n", i+1, r.cells[i])
			if i == r.index {
				tm.Println(tm.Background(tm.Color(cellStr, tm.BLACK), tm.BLUE))
				continue
			}
			tm.Printf(cellStr)
		}

		tm.Flush()

		///

		b := make([]byte, 1)
		os.Stdin.Read(b)
		waitChan <- struct{}{}
	}
}

// TODO: remove
func main() {
	f, err := os.OpenFile("./examples/test.bf", os.O_RDONLY, 0666)
	if err != nil {
		panic(err)
	}

	inst, err := Compile(f)
	if err != nil {
		panic(err)
	}

	r := BFRuntime{
		cells:        make([]byte, 1),
		index:        0,
		instructions: inst,
		instIndex:    0,
	}

	//r.Execute()
	//r.Debug()
	r.Execute(nil)
}
