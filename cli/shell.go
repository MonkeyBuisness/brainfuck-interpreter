package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	tm "github.com/buger/goterm"
	"github.com/c-bata/go-prompt"
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
	tm.Flush()

	completer := func(d prompt.Document) []prompt.Suggest {
		s := []prompt.Suggest{
			{Text: "users", Description: "Store the username and age"},
			{Text: "articles", Description: "Store the article text posted by user"},
			{Text: "comments", Description: "Store the text commented to articles"},
		}

		return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
	}

	pp := prompt.New(prompt.NewBuffer().JoinNextLine, completer)
	pp.Run()

	t := prompt.Input("> ", completer, prompt.OptionBreakLineCallback(func(d *prompt.Document) {

	}))
	fmt.Println(t)

	go func() {
		time.Sleep(time.Second)
		tm.Print(tm.Background(" ", tm.RED))
		tm.Flush()
	}()

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
		btn.BorderStyle.Modifier = ui.ModifierClear
		btn.SetRect(i*defaultHelpButtonWidth, 0, (i+1)*defaultHelpButtonWidth, 3)
		btn.TextStyle.Fg = ui.ColorWhite
		ui.Render(btn)
		i++
	}
	/*sinFloat64 := (func() []float64 {
		n := 400
		data := make([]float64, n)
		for i := range data {
			data[i] = 1 + math.Sin(float64(i)/5)
		}
		return data
	})()

	sl := widgets.NewSparkline()
	sl.Data = sinFloat64[:100]
	sl.LineColor = ui.ColorCyan
	sl.TitleStyle.Fg = ui.ColorWhite

	slg := widgets.NewSparklineGroup(sl)
	slg.Title = "Sparkline"

	lc := widgets.NewPlot()
	lc.Title = "braille-mode Line Chart"
	lc.Data = append(lc.Data, sinFloat64)
	lc.AxesColor = ui.ColorWhite
	lc.LineColors[0] = ui.ColorYellow

	gs := make([]*widgets.Gauge, 3)
	for i := range gs {
		gs[i] = widgets.NewGauge()
		gs[i].Percent = i * 10
		gs[i].BarColor = ui.ColorRed
	}

	ls := widgets.NewList()
	ls.Rows = []string{
		"[1] Downloading File 1",
		"",
		"",
		"",
		"[2] Downloading File 2",
		"",
		"",
		"",
		"[3] Uploading File 3",
	}
	ls.Border = false

	p := widgets.NewParagraph()
	p.Text = "<> This row has 3 columns\n<- Widgets can be stacked up like left side\n<- Stacked widgets are treated as a single widget"
	p.Title = "Demonstration"

	grid := ui.NewGrid()
	termWidth, termHeight := ui.TerminalDimensions()
	grid.SetRect(0, 0, termWidth, termHeight)

	grid.Set(
		ui.NewRow(1.0/2,
			ui.NewCol(1.0/2, slg),
			ui.NewCol(1.0/2, lc),
		),
		ui.NewRow(1.0/2,
			ui.NewCol(1.0/4, ls),
			ui.NewCol(1.0/4,
				ui.NewRow(.9/3, gs[0]),
				ui.NewRow(.9/3, gs[1]),
				ui.NewRow(1.2/3, gs[2]),
			),
			ui.NewCol(1.0/2, p),
		),
	)

	ui.Render(grid)

	tickerCount := 1
	uiEvents := ui.PollEvents()
	ticker := time.NewTicker(time.Second).C
	for {
		select {
		case e := <-uiEvents:
			switch e.ID {
			case "q", "<C-c>":
				return
			case "<Resize>":
				payload := e.Payload.(ui.Resize)
				grid.SetRect(0, 0, payload.Width, payload.Height)
				ui.Clear()
				ui.Render(grid)
			}
		case <-ticker:
			if tickerCount == 100 {
				return
			}
			for _, g := range gs {
				g.Percent = (g.Percent + 3) % 100
			}
			slg.Sparklines[0].Data = sinFloat64[tickerCount : tickerCount+100]
			lc.Data[0] = sinFloat64[2*tickerCount:]
			ui.Render(grid)
			tickerCount++
		}
	}*/
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
