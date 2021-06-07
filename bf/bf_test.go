package bf

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var commentRegxp = regexp.MustCompile(`\[(.*?)\]`)

type testReader struct {
	fn func(p []byte) (int, error)
}
type testWriter struct {
	fn func(p []byte) (n int, err error)
}

// Read reads test bytes slice.
func (r *testReader) Read(p []byte) (int, error) {
	if r.fn != nil {
		return r.fn(p)
	}

	p[0] = '1'
	return 1, nil
}

// Write writes test bytes slice.
func (w *testWriter) Write(p []byte) (n int, err error) {
	return w.fn(p)
}

func TestRuntime_Value(t *testing.T) {
	r := Runtime{
		cells: []byte{1, 2, 3},
		index: 1,
	}
	require.Equal(t, byte(2), r.Value())
}

func TestRuntime_Pointer(t *testing.T) {
	r := Runtime{
		index: 1,
	}
	require.Equal(t, 1, r.Pointer())
}

func TestRuntime_Next(t *testing.T) {
	r := Runtime{
		index: 0,
		cells: []byte{0},
	}
	r.Next()

	require.Equal(t, 1, r.index)
	require.Len(t, r.cells, 2)
}

func TestRuntime_Prev(t *testing.T) {
	r := Runtime{
		index: 3,
	}
	r.Prev()

	require.Equal(t, 2, r.index)
}

func TestRuntime_Inc(t *testing.T) {
	r := Runtime{
		index: 0,
		cells: []byte{10},
	}
	r.Inc()

	require.Equal(t, byte(11), r.cells[r.index])
}

func TestRuntime_Dec(t *testing.T) {
	r := Runtime{
		index: 0,
		cells: []byte{10},
	}
	r.Dec()

	require.Equal(t, byte(9), r.cells[r.index])
}

func TestRuntime_Jump(t *testing.T) {
	r := Runtime{
		instIndex: 2,
	}
	r.Jump(5)

	require.Equal(t, 5, r.instIndex)
}

func TestRuntime_Snapshot(t *testing.T) {
	r := Runtime{
		cells: []byte{1, 2, 3},
	}

	snapshot := r.Snapshot()
	require.ElementsMatch(t, snapshot, r.cells)
}

func TestRuntime_Instruction(t *testing.T) {
	r := Runtime{
		instIndex: 2,
		instructions: []Instruction{
			&InstructionDecValue{},
			&InstructionEndLoop{},
			&InstructionPrint{},
			&InstructionRead{},
		},
	}

	instruction, index := r.Instruction()
	require.Equal(t, 2, index)
	require.Equal(t, r.instructions[2], instruction)
}

func TestRuntime_Print(t *testing.T) {
	writer := bytes.Buffer{}

	r := Runtime{
		cells:     []byte{1, 2, 3},
		index:     1,
		outStream: &writer,
	}

	err := r.Print()
	require.NoError(t, err)
	require.Equal(t, byte(2), writer.Bytes()[0])
}

func TestRuntime_Read(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		reader := bytes.NewReader([]byte{})

		r := Runtime{
			cells:    []byte{1, 2, 3},
			index:    1,
			inStream: reader,
		}

		err := r.Read()
		require.EqualError(t, err, io.EOF.Error())
	})

	t.Run("all ok", func(t *testing.T) {
		reader := bytes.NewReader([]byte{100, 200})

		r := Runtime{
			cells:    []byte{1, 2, 3},
			index:    1,
			inStream: reader,
		}

		err := r.Read()
		require.NoError(t, err)
		require.Equal(t, byte(100), r.cells[1])
	})
}

func TestRuntime_Iterator(t *testing.T) {
	r := Runtime{
		it: defaultBFIterator{},
	}

	it := r.Iterator()
	require.NotNil(t, it)
	require.Equal(t, r.it, it)
}

func TestRuntime_IterateBy(t *testing.T) {
	r := Runtime{}

	it := defaultBFIterator{}
	r.IterateBy(it)

	require.NotNil(t, r.it)
	require.Equal(t, it, r.it)
}

func TestRuntime_Instructions(t *testing.T) {
	r := Runtime{
		instructions: []Instruction{
			&InstructionDecValue{},
			&InstructionIncValue{},
		},
	}

	instructions := r.Instructions()
	require.ElementsMatch(t, r.instructions, instructions)
}

func TestRuntime_Execute(t *testing.T) {
	t.Run("context deadline", func(t *testing.T) {
		r := Runtime{
			cells: make([]byte, 1),
			instructions: []Instruction{
				&InstructionIncValue{},
				&InstructionStartLoop{
					EndLoopIndex: 2,
				},
				&InstructionEndLoop{
					StartLoopIndex: 1,
				},
			},
			it: defaultBFIterator{},
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
		defer cancel()

		err := r.Execute(ctx, nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, context.DeadlineExceeded))
	})

	t.Run("execute instruction error", func(t *testing.T) {
		r := Runtime{
			cells:    make([]byte, 1),
			inStream: os.Stdin,
			instructions: []Instruction{
				&InstructionRead{},
			},
			it: defaultBFIterator{},
		}

		err := r.Execute(context.Background(), nil)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrReadSymbol))
	})

	t.Run("all ok", func(t *testing.T) {
		wd, err := os.Getwd()
		require.NoError(t, err)

		files, err := filepath.Glob(path.Join(wd, "..", "examples", "*.bf"))
		require.NoError(t, err)
		require.NotEmpty(t, files)

		outStream := bytes.Buffer{}
		inStream := testReader{}

		for i := range files {
			t.Run(fmt.Sprintf("test file: %s", files[i]), func(t *testing.T) {
				defer outStream.Reset()

				f, err := os.OpenFile(files[i], os.O_RDONLY, 0666)
				require.NoError(t, err)

				instructions, err := Compile(f)
				require.NoError(t, err)
				require.NotEmpty(t, instructions)

				r := Runtime{
					cells:        make([]byte, 1),
					instructions: instructions,
					inStream:     &inStream,
					outStream:    &outStream,
					it:           defaultBFIterator{},
				}

				waitChan := make(chan struct{}, len(instructions))
				go func() {
					waitChan <- struct{}{}
					close(waitChan)
				}()
				err = r.Execute(context.Background(), waitChan)
				require.NoError(t, err)
				require.NotNil(t, outStream)
				defer func() {
					err = f.Close()
					require.NoError(t, err)
				}()

				_, err = f.Seek(0, 0)
				require.NoError(t, err)
				var fContent bytes.Buffer
				_, err = fContent.ReadFrom(f)
				require.NoError(t, err)
				require.NotEmpty(t, fContent)

				comment := commentRegxp.Find(fContent.Bytes())
				require.NotEmpty(t, comment)
				require.Equal(t, string(comment[1:len(comment)-1]), outStream.String())
			})
		}
	})
}

func Test_defaultBFIterator_HasNext(t *testing.T) {
	r := Runtime{
		instIndex: 2,
		it:        defaultBFIterator{},
	}

	require.False(t, r.it.HasNext(&r))
	r.instructions = []Instruction{
		&InstructionDecValue{},
		&InstructionIncValue{},
		&InstructionEndLoop{},
		&InstructionPrevCell{},
	}
	require.True(t, r.it.HasNext(&r))
}

func Test_defaultBFIterator_Next(t *testing.T) {
	r := Runtime{
		instIndex: 2,
		instructions: []Instruction{
			&InstructionDecValue{},
			&InstructionIncValue{},
			&InstructionEndLoop{},
			&InstructionPrevCell{},
		},
		it: defaultBFIterator{},
	}

	instruction, instIndex := r.it.Next(&r)
	require.Equal(t, 2, instIndex)
	require.NotNil(t, instruction)
	require.Equal(t, r.instructions[2], instruction)
	require.Equal(t, 3, r.instIndex)
}

func TestInstruction_Cmd(t *testing.T) {
	instructionCmdMap := map[rune]Instruction{
		'>': &InstructionNextCell{},
		'<': &InstructionPrevCell{},
		'-': &InstructionDecValue{},
		'+': &InstructionIncValue{},
		']': &InstructionEndLoop{},
		'[': &InstructionStartLoop{},
		'.': &InstructionPrint{},
		',': &InstructionRead{},
	}

	for cmd, instruction := range instructionCmdMap {
		require.Equal(t, cmd, instruction.Cmd())
	}
}

func TestInstructionNextCell_Execute(t *testing.T) {
	r := Runtime{
		index: 1,
	}

	inst := InstructionNextCell{}
	err := inst.Execute(1, &r)
	require.NoError(t, err)
	require.Equal(t, 2, r.index)
}

func TestInstructionPrevCell_Execute(t *testing.T) {
	r := Runtime{
		index: 1,
	}

	inst := InstructionPrevCell{}
	err := inst.Execute(1, &r)
	require.NoError(t, err)
	require.Equal(t, 0, r.index)
}

func TestInstructionIncValue_Execute(t *testing.T) {
	r := Runtime{
		index: 1,
		cells: []byte{1, 2, 5},
	}

	inst := InstructionIncValue{}
	err := inst.Execute(1, &r)
	require.NoError(t, err)
	require.Equal(t, byte(3), r.cells[r.index])
}

func TestInstructionDecValue_Execute(t *testing.T) {
	r := Runtime{
		index: 1,
		cells: []byte{1, 2, 5},
	}

	inst := InstructionDecValue{}
	err := inst.Execute(1, &r)
	require.NoError(t, err)
	require.Equal(t, byte(1), r.cells[r.index])
}

func TestInstructionStartLoop_Execute(t *testing.T) {
	r := Runtime{
		index: 1,
		cells: []byte{1, 0, 5},
	}

	inst := InstructionStartLoop{
		EndLoopIndex: 2,
	}
	err := inst.Execute(1, &r)
	require.NoError(t, err)
	require.Equal(t, 2, r.instIndex)
}

func TestInstructionEndLoop_Execute(t *testing.T) {
	r := Runtime{
		index: 1,
		cells: []byte{1, 3, 5},
	}

	inst := InstructionEndLoop{
		StartLoopIndex: 2,
	}
	err := inst.Execute(1, &r)
	require.NoError(t, err)
	require.Equal(t, 2, r.instIndex)
}

func TestInstructionPrint_Execute(t *testing.T) {
	t.Run("print error", func(t *testing.T) {
		writeErr := errors.New("write error")
		writer := testWriter{
			fn: func(p []byte) (n int, err error) {
				return 0, writeErr
			},
		}
		r := Runtime{
			index:     1,
			cells:     []byte{1, 6, 5},
			outStream: &writer,
		}

		inst := InstructionPrint{}
		err := inst.Execute(1, &r)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrWriteSymbol))
	})

	t.Run("all ok", func(t *testing.T) {
		writer := bytes.NewBuffer([]byte{123})
		r := Runtime{
			index:     1,
			cells:     []byte{1, 6, 5},
			outStream: writer,
		}

		inst := InstructionPrint{}
		err := inst.Execute(1, &r)
		require.NoError(t, err)
		require.Equal(t, 2, writer.Len())
		require.Equal(t, byte(6), writer.Bytes()[1])
	})
}

func TestInstructionRead_Execute(t *testing.T) {
	t.Run("read error", func(t *testing.T) {
		readErr := errors.New("read error")
		reader := testReader{
			fn: func(p []byte) (n int, err error) {
				return 0, readErr
			},
		}
		r := Runtime{
			index:    1,
			cells:    []byte{1, 6, 5},
			inStream: &reader,
		}

		inst := InstructionRead{}
		err := inst.Execute(1, &r)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrReadSymbol))
	})

	t.Run("all ok", func(t *testing.T) {
		reader := bytes.NewBuffer([]byte{123})
		r := Runtime{
			index:    1,
			cells:    []byte{1, 6, 5},
			inStream: reader,
		}

		inst := InstructionRead{}
		err := inst.Execute(1, &r)
		require.NoError(t, err)
		require.Equal(t, byte(123), r.cells[r.index])
	})
}

func Test_NewRuntime(t *testing.T) {
	instructions := []Instruction{
		&InstructionDecValue{},
		&InstructionNextCell{},
		&InstructionPrevCell{},
	}
	in := testReader{}
	out := testWriter{}

	r := NewRuntime(instructions, &in, &out)
	require.NotNil(t, r)
	require.Len(t, r.cells, 1)
	require.Equal(t, 0, r.index)
	require.Equal(t, instructions, r.instructions)
	require.Equal(t, 0, r.instIndex)
	require.Equal(t, &in, r.inStream)
	require.Equal(t, &out, r.outStream)
	require.NotNil(t, r.it)
}

func Test_Compile(t *testing.T) {
	t.Run("compilation error", func(t *testing.T) {
		sourceReader := testReader{
			fn: func(p []byte) (int, error) {
				return 0, errors.New("read error")
			},
		}

		_, err := Compile(&sourceReader)
		require.Error(t, err)
		require.True(t, errors.Is(err, ErrCompilation))
	})

	t.Run("all ok", func(t *testing.T) {
		sourceReader := testReader{
			fn: func(p []byte) (int, error) {
				code := []byte{'+', '+', '+', '>', '+'}
				copy(p, code)

				return len(code), io.EOF
			},
		}

		instructions, err := Compile(&sourceReader)
		require.NoError(t, err)
		require.NotEmpty(t, instructions)
		expInstructions := []Instruction{
			&InstructionIncValue{},
			&InstructionIncValue{},
			&InstructionIncValue{},
			&InstructionNextCell{},
			&InstructionIncValue{},
		}
		require.Equal(t, expInstructions, instructions)
	})
}
