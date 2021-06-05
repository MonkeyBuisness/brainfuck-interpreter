package bf

import (
	"bytes"
	"context"
	"io"
)

// Runtime represents Brainfuck runtime instance.
//
// This instance is responsible for executing list of the provided
// Brainfuck commands and managing memory cells via execution process.
type Runtime struct {
	cells        []byte
	index        int
	instructions []Instruction
	instIndex    int
	inStream     io.Reader
	outStream    io.Writer
	it           InstructionIterator
}

// Instruction represents execution interface of a single Brainfuck instruction
// on the runtime process.
//
// In other words it represents command handler for a single Brainfuck operator
// such as +, -, <, >, etc.
type Instruction interface {
	Execute(i int, runtime *Runtime) error
	Cmd() rune
}

// Brainfuck instruction.
type (
	// InstructionNextCell represents handler for the '>' Brainfuck command.
	InstructionNextCell struct{}
	// InstructionPrevCell represents handler for the '<' Brainfuck command.
	InstructionPrevCell struct{}
	// InstructionIncValue represents handler for the '+' Brainfuck command.
	InstructionIncValue struct{}
	// InstructionDecValue represents handler for the '-' Brainfuck command.
	InstructionDecValue struct{}
	// InstructionStartLoop represents handler for the '[' Brainfuck command.
	InstructionStartLoop struct {
		EndLoopIndex int
	}
	// InstructionEndLoop represents handler for the ']' Brainfuck command.
	InstructionEndLoop struct {
		StartLoopIndex int
	}
	// InstructionPrint represents handler for the '.' Brainfuck command.
	InstructionPrint struct{}
	// InstructionRead represents handler for the ',' Brainfuck command.
	InstructionRead struct{}
)

// InstructionIterator represents interface to iterate over the Brainfuck instructions.
type InstructionIterator interface {
	HasNext(runtime *Runtime) bool
	Next(runtime *Runtime) (Instruction, int)
}

type defaultBFIterator struct{}

// Value returns value of a current cell.
func (r *Runtime) Value() byte {
	return r.cells[r.index]
}

// Pointer returns current's cell index.
func (r *Runtime) Pointer() int {
	return r.index
}

// Next moves pointer to the next cell.
func (r *Runtime) Next() {
	r.index++

	if r.index == len(r.cells) {
		r.cells = append(r.cells, 0)
	}
}

// Prev moves pointer to the previous cell.
func (r *Runtime) Prev() {
	r.index--
}

// Inc increments current's cell value.
func (r *Runtime) Inc() {
	r.cells[r.index]++
}

// Dec decrements current's cell value.
func (r *Runtime) Dec() {
	r.cells[r.index]--
}

// Jump sets instruction index to execute.
func (r *Runtime) Jump(i int) {
	r.instIndex = i
}

// Snapshot returns runtime's cell values as a byte slice.
func (r *Runtime) Snapshot() []byte {
	cp := make([]byte, len(r.cells))
	copy(cp, r.cells)

	return cp
}

// Instruction returns current instruction to execute and its index.
func (r *Runtime) Instruction() (Instruction, int) {
	return r.instructions[r.instIndex], r.instIndex
}

// Print writes current cell's value to the output writer stream.
func (r *Runtime) Print() error {
	_, err := r.outStream.Write([]byte{r.Value()})
	return err
}

// Read reads one byte (symbol) to the current cell's value from the input reader stream.
func (r *Runtime) Read() error {
	b := make([]byte, 1)
	_, err := r.inStream.Read(b)
	if err != nil {
		return err
	}

	r.cells[r.index] = b[0]

	return nil
}

// Iterator returns provided runtime iterator.
func (r *Runtime) Iterator() InstructionIterator {
	return r.it
}

// IterateBy sets custom runtime iterator.
func (r *Runtime) IterateBy(it InstructionIterator) {
	r.it = it
}

// Instructions returns runtime instructions slice.
func (r *Runtime) Instructions() []Instruction {
	return r.instructions
}

// Execute starts runtime process.
//
// waitChan (<-chan struct{}) param can be used to debug or pause execution process.
// After each instruction executed this channel will be checked for nil (or closed).
// If it's not nil, then runtime will wait for chanel's value and continue execution.
// If it's closed (or nil), then runtime skip channel reading and continue execution.
// Most of the time you can pass nil as a waitChan value.
func (r *Runtime) Execute(ctx context.Context, waitChan <-chan struct{}) error {
	errChan := make(chan error, 1)
	go func(errChan chan error) {
		defer close(errChan)

		for it := r.Iterator(); it.HasNext(r); {
			if waitChan != nil {
				<-waitChan
			}

			instruction, index := it.Next(r)

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
func (it defaultBFIterator) HasNext(r *Runtime) bool {
	return r.instIndex < len(r.instructions)
}

// Next returns next instruction from the execution list.
func (it defaultBFIterator) Next(r *Runtime) (Instruction, int) {
	defer r.Jump(r.instIndex + 1)

	return r.Instruction()
}

// Execute executes command.
func (i *InstructionNextCell) Execute(index int, runtime *Runtime) error {
	runtime.Next()

	return nil
}

// Cmd returns name (single character) of the command.
func (i *InstructionNextCell) Cmd() rune {
	return '>'
}

// Execute executes command.
func (i *InstructionPrevCell) Execute(index int, runtime *Runtime) error {
	runtime.Prev()

	return nil
}

// Cmd returns name (single character) of the command.
func (i *InstructionPrevCell) Cmd() rune {
	return '<'
}

// Execute executes command.
func (i *InstructionIncValue) Execute(index int, runtime *Runtime) error {
	runtime.Inc()

	return nil
}

// Cmd returns name (single character) of the command.
func (i *InstructionIncValue) Cmd() rune {
	return '+'
}

// Execute executes command.
func (i *InstructionDecValue) Execute(index int, runtime *Runtime) error {
	runtime.Dec()

	return nil
}

// Cmd returns name (single character) of the command.
func (i *InstructionDecValue) Cmd() rune {
	return '-'
}

// Execute executes command.
func (i *InstructionStartLoop) Execute(index int, runtime *Runtime) error {
	if runtime.Value() == 0 {
		runtime.Jump(i.EndLoopIndex)
	}

	return nil
}

// Cmd returns name (single character) of the command.
func (i *InstructionStartLoop) Cmd() rune {
	return '['
}

// Execute executes command.
func (i *InstructionEndLoop) Execute(index int, runtime *Runtime) error {
	if runtime.Value() > 0 {
		runtime.Jump(i.StartLoopIndex)
	}

	return nil
}

// Cmd returns name (single character) of the command.
func (i *InstructionEndLoop) Cmd() rune {
	return ']'
}

// Execute executes command.
func (i *InstructionPrint) Execute(index int, runtime *Runtime) error {
	if err := runtime.Print(); err != nil {
		return NewError(WriteSymbolError, err)
	}

	return nil
}

// Cmd returns name (single character) of the command.
func (i *InstructionPrint) Cmd() rune {
	return '.'
}

// Execute executes command.
func (i *InstructionRead) Execute(index int, runtime *Runtime) error {
	if err := runtime.Read(); err != nil {
		return NewError(ReadSymbolError, err)
	}

	return nil
}

// Cmd returns name (single character) of the command.
func (i *InstructionRead) Cmd() rune {
	return ','
}

// NewRuntime creates new Brainfuck runtime instance.
func NewRuntime(instructions []Instruction, in io.Reader, out io.Writer) Runtime {
	runtime := Runtime{
		cells:        make([]byte, 1),
		index:        0,
		instructions: instructions,
		instIndex:    0,
		inStream:     in,
		outStream:    out,
		it:           defaultBFIterator{},
	}

	return runtime
}

// Compile compiles Brainfuck code and returns slice of instructions to execute.
func Compile(sourceInput io.Reader) ([]Instruction, error) {
	var p bytes.Buffer
	_, err := p.ReadFrom(sourceInput)
	if err != nil {
		return nil, NewError(CompilationError, err)
	}

	instructions := make([]Instruction, 0, p.Len())
	loopOffsets := make([]int, 0)

	for i := range p.Bytes() {
		var instruction Instruction

		switch p.Bytes()[i] {
		case '>':
			instruction = &InstructionNextCell{}
		case '<':
			instruction = &InstructionPrevCell{}
		case '+':
			instruction = &InstructionIncValue{}
		case '-':
			instruction = &InstructionDecValue{}
		case '.':
			instruction = &InstructionPrint{}
		case ',':
			instruction = &InstructionRead{}
		case '[':
			instruction = &InstructionStartLoop{}
			loopOffsets = append(loopOffsets, len(instructions))
		case ']':
			endLoopInstruction := &InstructionEndLoop{}
			endLoopInstruction.StartLoopIndex = loopOffsets[len(loopOffsets)-1]
			loopOffsets = loopOffsets[:len(loopOffsets)-1]

			if startLoopInstruction, ok :=
				instructions[endLoopInstruction.StartLoopIndex].(*InstructionStartLoop); ok {
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
