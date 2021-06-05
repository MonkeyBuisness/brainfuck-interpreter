package bf

import (
	"bytes"
	"context"
	"io"
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
	inStream     io.Reader
	outStream    io.Writer
	it           BFInstructionIterator
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

// Brainfuck instruction.
type (
	// BFInstructionNextCell represents handler for the '>' Brainfuck command.
	BFInstructionNextCell struct{}
	// BFInstructionPrevCell represents handler for the '<' Brainfuck command.
	BFInstructionPrevCell struct{}
	// BFInstructionIncValue represents handler for the '+' Brainfuck command.
	BFInstructionIncValue struct{}
	// BFInstructionDecValue represents handler for the '-' Brainfuck command.
	BFInstructionDecValue struct{}
	// BFInstructionStartLoop represents handler for the '[' Brainfuck command.
	BFInstructionStartLoop struct {
		EndLoopIndex int
	}
	// BFInstructionEndLoop represents handler for the ']' Brainfuck command.
	BFInstructionEndLoop struct {
		StartLoopIndex int
	}
	// BFInstructionPrint represents handler for the '.' Brainfuck command.
	BFInstructionPrint struct{}
	// BFInstructionRead represents handler for the ',' Brainfuck command.
	BFInstructionRead struct{}
)

// BFInstructionIterator represents interface to iterate over the Brainfuck instructions.
type BFInstructionIterator interface {
	HasNext() bool
	Next() (BFInstruction, int)
}

type defaultBFIterator struct {
	runtime *BFRuntime
}

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

// Print writes current cell's value to the output writer stream.
func (r *BFRuntime) Print() error {
	_, err := r.outStream.Write([]byte{r.Value()})
	return err
}

// Read reads one byte (symbol) to the current cell's value from the input reader stream.
func (r *BFRuntime) Read() error {
	b := make([]byte, 1)
	_, err := r.inStream.Read(b)
	if err != nil {
		return err
	}

	r.cells[r.index] = b[0]

	return nil
}

// Iterator returns provided runtime iterator.
func (r *BFRuntime) Iterator() BFInstructionIterator {
	return r.it
}

// IterateBy sets custom runtime iterator.
func (r *BFRuntime) IterateBy(it BFInstructionIterator) {
	r.it = it
}

// Instructions returns runtime instructions slice.
func (r *BFRuntime) Instructions() []BFInstruction {
	return r.instructions
}

// Execute starts runtime process.
//
// waitChan (<-chan struct{}) param can be used to debug or pause execution process.
// After each instruction executed this channel will be checked for nil (or closed).
// If it's not nil, then runtime will wait for chanel's value and continue execution.
// If it's closed (or nil), then runtime skip channel reading and continue execution.
// Most of the time you can pass nil as a waitChan value.
func (r *BFRuntime) Execute(ctx context.Context, waitChan <-chan struct{}) error {
	errChan := make(chan error, 1)
	go func(errChan chan error) {
		defer close(errChan)

		for it := r.Iterator(); it.HasNext(); {
			if waitChan != nil {
				<-waitChan
			}

			instruction, index := it.Next()

			if err := instruction.Execute(index, r); err != nil {
				errChan <- err
				return
			}
		}
	}(errChan)

	for {
		select {
		case <-ctx.Done():
			return context.DeadlineExceeded
		case err := <-errChan:
			return err
		}
	}
}

// HasNext returns true if current instruction is not last in the execution list.
func (it *defaultBFIterator) HasNext() bool {
	return it.runtime.instIndex < len(it.runtime.instructions)
}

// Next returns next instruction from the execution list.
func (it *defaultBFIterator) Next() (BFInstruction, int) {
	defer it.runtime.Jump(it.runtime.instIndex + 1)

	return it.runtime.Instruction()
}

// Execute executes command.
func (i *BFInstructionNextCell) Execute(index int, runtime *BFRuntime) error {
	runtime.Next()

	return nil
}

// Cmd returns name (single character) of the command.
func (i *BFInstructionNextCell) Cmd() rune {
	return '>'
}

// Execute executes command.
func (i *BFInstructionPrevCell) Execute(index int, runtime *BFRuntime) error {
	runtime.Prev()

	return nil
}

// Cmd returns name (single character) of the command.
func (i *BFInstructionPrevCell) Cmd() rune {
	return '<'
}

// Execute executes command.
func (i *BFInstructionIncValue) Execute(index int, runtime *BFRuntime) error {
	runtime.Inc()

	return nil
}

// Cmd returns name (single character) of the command.
func (i *BFInstructionIncValue) Cmd() rune {
	return '+'
}

// Execute executes command.
func (i *BFInstructionDecValue) Execute(index int, runtime *BFRuntime) error {
	runtime.Dec()

	return nil
}

// Cmd returns name (single character) of the command.
func (i *BFInstructionDecValue) Cmd() rune {
	return '-'
}

// Execute executes command.
func (i *BFInstructionStartLoop) Execute(index int, runtime *BFRuntime) error {
	if runtime.Value() == 0 {
		runtime.Jump(i.EndLoopIndex)
	}

	return nil
}

// Cmd returns name (single character) of the command.
func (i *BFInstructionStartLoop) Cmd() rune {
	return '['
}

// Execute executes command.
func (i *BFInstructionEndLoop) Execute(index int, runtime *BFRuntime) error {
	if runtime.Value() > 0 {
		runtime.Jump(i.StartLoopIndex)
	}

	return nil
}

// Cmd returns name (single character) of the command.
func (i *BFInstructionEndLoop) Cmd() rune {
	return ']'
}

// Execute executes command.
func (i *BFInstructionPrint) Execute(index int, runtime *BFRuntime) error {
	if err := runtime.Print(); err != nil {
		return NewError(WriteSymbolError, err)
	}

	return nil
}

// Cmd returns name (single character) of the command.
func (i *BFInstructionPrint) Cmd() rune {
	return '.'
}

// Execute executes command.
func (i *BFInstructionRead) Execute(index int, runtime *BFRuntime) error {
	if err := runtime.Read(); err != nil {
		return NewError(ReadSymbolError, err)
	}

	return nil
}

// Cmd returns name (single character) of the command.
func (i *BFInstructionRead) Cmd() rune {
	return ','
}

// NewRuntime creates new Brainfuck runtime instance.
func NewRuntime(instructions []BFInstruction, in io.Reader, out io.Writer) BFRuntime {
	runtime := BFRuntime{
		cells:        make([]byte, 1),
		index:        0,
		instructions: instructions,
		instIndex:    0,
		inStream:     in,
		outStream:    out,
	}
	runtime.it = &defaultBFIterator{&runtime}

	return runtime
}

// Compile compiles Brainfuck code and returns slice of instructions to execute.
func Compile(sourceInput io.Reader) ([]BFInstruction, error) {
	var p bytes.Buffer
	_, err := p.ReadFrom(sourceInput)
	if err != nil {
		return nil, NewError(CompilationError, err)
	}

	instructions := make([]BFInstruction, 0, p.Len())
	loopOffsets := make([]int, 0)

	for i := range p.Bytes() {
		var instruction BFInstruction

		switch p.Bytes()[i] {
		case '>':
			instruction = &BFInstructionNextCell{}
		case '<':
			instruction = &BFInstructionPrevCell{}
		case '+':
			instruction = &BFInstructionIncValue{}
		case '-':
			instruction = &BFInstructionDecValue{}
		case '.':
			instruction = &BFInstructionPrint{}
		case ',':
			instruction = &BFInstructionRead{}
		case '[':
			instruction = &BFInstructionStartLoop{}
			loopOffsets = append(loopOffsets, len(instructions))
		case ']':
			endLoopInstruction := &BFInstructionEndLoop{}
			endLoopInstruction.StartLoopIndex = loopOffsets[len(loopOffsets)-1]
			loopOffsets = loopOffsets[:len(loopOffsets)-1]

			if startLoopInstruction, ok :=
				instructions[endLoopInstruction.StartLoopIndex].(*BFInstructionStartLoop); ok {
				startLoopInstruction.EndLoopIndex = len(instructions)
			}

			instruction = endLoopInstruction
		}

		if instruction != nil {
			instructions = append(instructions, instruction)
		}
	}

	return instructions, nil
}
