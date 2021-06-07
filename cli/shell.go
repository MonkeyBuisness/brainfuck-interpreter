package cli

import (
	"context"
	"fmt"
	"os"

	tm "github.com/buger/goterm"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

const defaultHelpButtonWidth = 16

var cmdColors = map[rune]int{
	'+': tm.YELLOW,
	'-': tm.YELLOW,
	'>': tm.BLUE,
	'<': tm.BLUE,
	'[': tm.RED,
	']': tm.RED,
	'.': tm.CYAN,
	',': tm.CYAN,
}

var hotKeyButtons = map[string]hotkey{
	"<C-o>": {
		name:    "Open  <Ctrl+O>",
		handler: openFileHandler,
	},
	"<C-s>": {
		name:    "Save  <Ctrl+S>",
		handler: saveFileHandler,
	},
	"<C-d>": {
		name:    "Debug <Ctrl+D>",
		handler: debugHandler,
	},
	"<C-r>": {
		name:    "Run   <Ctrl+R>",
		handler: runHandler,
	},
	"<C-c>": {
		name:    "Exit  <Ctrl+C>",
		handler: exitHandler,
	},
}

type hotkey struct {
	name    string
	handler func(ctx context.Context, codeBuffer []byte) error
}

// RunShell runs interactive shell.
func RunShell(ctx context.Context) error {
	if err := ui.Init(); err != nil {
		return fmt.Errorf("could not init shell ui: %v", err)
	}
	defer ui.Close()

	renderHelpMenu()

	tm.MoveCursor(0, 9)

	for e := range ui.PollEvents() {
		if e.Type != ui.KeyboardEvent {
			continue
		}

		// check hotkeys.
		key, ok := hotKeyButtons[e.ID]
		if ok {
			if err := key.handler(ctx, []byte{}); err != nil {
				tm.Println(tm.Color(err.Error(), tm.RED))
			}

			continue
		}

		switch e.ID {
		default:
			color, ok := cmdColors[[]rune(e.ID)[0]]
			if ok {
				tm.Print(tm.Color(e.ID, color))
			} else {
				tm.Print(e.ID)
			}

			tm.MoveCursorDown(1)

			tm.Flush()
			//p2.TextStyle.Fg = ui.ColorBlue
			//ui.Render(p2)
		}
	}

	return nil
}

func renderHelpMenu() {
	var i int
	for _, key := range hotKeyButtons {
		btn := widgets.NewParagraph()
		btn.Text = key.name
		btn.Border = true
		btn.BorderStyle.Fg = ui.ColorCyan
		btn.SetRect(i*defaultHelpButtonWidth, 0, (i+1)*defaultHelpButtonWidth, 3)
		btn.TextStyle.Fg = ui.ColorWhite
		ui.Render(btn)
		i++
	}
}

func openFileHandler(ctx context.Context, codeBuffer []byte) error {
	// TODO: add
	return nil
}

func saveFileHandler(ctx context.Context, codeBuffer []byte) error {
	// TODO: add
	return nil
}

func debugHandler(ctx context.Context, codeBuffer []byte) error {
	// TODO: add
	return nil
}

func runHandler(ctx context.Context, codeBuffer []byte) error {
	// TODO: add
	return nil
}

func exitHandler(ctx context.Context, codeBuffer []byte) error {
	os.Exit(0)

	return nil
}
