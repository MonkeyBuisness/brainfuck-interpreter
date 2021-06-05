package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/MonkeyBuisness/brainfuck-interpreter/bf"
	tm "github.com/buger/goterm"
)

func Debug(r *bf.Runtime) {
	waitChan := make(chan struct{}, 1)
	go r.Execute(context.Background(), waitChan)

	tm.Clear()
	for {
		///
		tm.Clear()
		tm.MoveCursor(1, 1)
		instructions := r.Instructions()
		_, instIndex := r.Instruction()

		for i := range instructions {
			str := fmt.Sprintf("%c", instructions[i].Cmd())
			if i == instIndex {
				tm.Print(" |")
				tm.Print(tm.Color(str, tm.RED))
				tm.Print("| ")
				continue
			}
			// tm.Printf("\n\n[###%d####]\n\n", instIndex)

			tm.Print(str)
		}

		tm.Print("\n\nCELLS:\n\n")
		cells := r.Snapshot()
		for i := range cells {
			cellStr := fmt.Sprintf("[%d]: %d\n", i+1, cells[i])
			if i == r.Pointer() {
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
