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

type testReader struct{}

// Read reads test bytes slice.
func (r *testReader) Read(p []byte) (int, error) {
	p[0] = '1'
	return 1, nil
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
		require.True(t, errors.Is(err, ReadSymbolError))
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
					for range instructions {
						waitChan <- struct{}{}
					}
				}()
				err = r.Execute(context.Background(), nil)
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
	// TODO: add
}
