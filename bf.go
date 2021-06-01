package main

import (
	"bytes"
	"fmt"
	"io"
	"os"

	tm "github.com/buger/goterm"
)

type BFRuntime struct {
	cells        []byte
	index        int
	instructions []BFInstruction
	instIndex    int
}

type BFInstruction interface {
	Execute(i int, runtime *BFRuntime) error
	String() string
}

func (r *BFRuntime) Value() byte {
	return r.cells[r.index]
}

func (r *BFRuntime) Index() int {
	return r.index
}

func (r *BFRuntime) NextCell() {
	r.index++

	if r.index == len(r.cells) {
		r.cells = append(r.cells, 0)
	}
}

func (r *BFRuntime) PrevCell() {
	r.index--
}

func (r *BFRuntime) Inc() {
	r.cells[r.index]++
}

func (r *BFRuntime) Dec() {
	r.cells[r.index]--
}

func (r *BFRuntime) MoveInstructionIndex(i int) {
	r.instIndex = i
}

func (r *BFRuntime) Execute(waitChan <-chan struct{}) {
	for r.instIndex = 0; r.instIndex < len(r.instructions); r.instIndex++ {
		if waitChan != nil {
			<-waitChan
		}
		if err := r.instructions[r.instIndex].Execute(r.instIndex, r); err != nil {
			panic(err)
		}
	}
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

////////////////////////////////////////////////////////

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

func bfNewRuntime(in io.RuneReader, out io.Writer, cmds []byte) bfRuntime {
	r := bfRuntime{
		cells:        make([]rune, 1),
		index:        0,
		cmdIndex:     0,
		loopOffsets:  make([]int, 0),
		inputStream:  in,
		outputStream: out,
		cmdList:      cmds,
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

func (r *bfRuntime) execute() error {
	for r.cmdIndex = 0; r.cmdIndex < len(r.cmdList); r.cmdIndex++ {
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
	if r.cells[r.index] > 0 {
		r.loopOffsets = append(r.loopOffsets, r.cmdIndex)
		return nil
	}

	///
	/*for ; r.cmdList[r.cmdIndex] != ']'; r.cmdIndex++ {
	}*/
	///

	return nil
}

func (r *bfRuntime) endLoop() error {
	if len(r.loopOffsets) == 0 {
		return nil
	}

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
	rt := bfNewRuntime(input, output, p.Bytes())

	return rt.execute()
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
