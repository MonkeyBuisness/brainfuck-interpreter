package main

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/stretchr/testify/require"
)

var commentRegxp = regexp.MustCompile(`\[(.*?)\]`)

type testRuneReader struct{}

// ReadRune reads test rune.
func (rr *testRuneReader) ReadRune() (r rune, size int, err error) {
	return '1', 1, nil
}

func Test_Execute(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NotEmpty(t, wd)

	files, err := filepath.Glob(path.Join(wd, "examples", "*.bf1"))
	require.NoError(t, err)
	require.NotEmpty(t, files)

	outStream := bytes.Buffer{}
	inStream := testRuneReader{}

	for i := range files {
		t.Run(fmt.Sprintf("test file: %s", files[i]), func(t *testing.T) {
			defer outStream.Reset()

			f, err := os.OpenFile(files[i], os.O_RDONLY, 0666)
			require.NoError(t, err)

			err = Execute(f, &inStream, &outStream)
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
}

func Test_test(t *testing.T) {
	f, err := os.OpenFile("./examples/factorial.bf1", os.O_RDONLY, 0666)
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

	r.Execute(nil)
}
