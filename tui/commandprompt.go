package tui

import (
	"strings"

	"github.com/PawelReich/gogdb/client"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"fmt"
)

const Prompt = "❯ "

type CommandPrompt struct {
	Pane *tview.Flex

	history *tview.TextView
	input   *tview.InputField

	app         *GoGdb
	lastCommand string
}

func NewCommandPrompt(app *GoGdb) *CommandPrompt {
	cmdHistory := tview.NewTextView()
	cmdHistory.SetDynamicColors(true)
	cmdHistory.SetScrollable(true)

	cmdPrompt := tview.NewInputField()
	cmdPrompt.SetLabel(Prompt)
	cmdPrompt.SetFieldBackgroundColor(tcell.ColorBlack)

	flex := tview.NewFlex().SetDirection(tview.FlexRow)

	view := &CommandPrompt{app: app, history: cmdHistory, input: cmdPrompt, Pane: flex}

	cmdPrompt.SetDoneFunc(func(key tcell.Key) {
		if key != tcell.KeyEnter {
			return
		}

		command := cmdPrompt.GetText()
		if command == "" {
			command = view.lastCommand
		} else {
			view.lastCommand = command
		}

		fmt.Fprintf(cmdHistory, "[white]%s%s\n", Prompt, command)

		var fut <-chan client.AsyncDecodedResult[string]

		if command[0] == '-' {
			// Drop '-' as it is assumed in `SendAsync`
			command = command[1:]
			splitCmd := strings.Split(command, " ")

			// Prepare MI command and its arguments
			command = splitCmd[0]
			splitCmd = splitCmd[1:]

			fut = client.SendDecodeAsync[string](app.Debugger, command, splitCmd...)
		} else {
			fut = app.Debugger.SendConsoleCommandAsync(command)
		}

		go func() {
			res := <-fut
			app.LogInfo(res.Result)

			cmdPrompt.SetText("")
			cmdHistory.ScrollToEnd()
		}()
	})

	flex.SetBorder(true)
	flex.SetTitle("Command Prompt")
	flex.AddItem(cmdHistory, 0, 1, false)
	flex.AddItem(cmdPrompt, 1, 0, true)

	scrollLog := cmdHistory.InputHandler()
	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp, tcell.KeyDown, tcell.KeyLeft, tcell.KeyRight,
			tcell.KeyPgUp, tcell.KeyPgDn:
			scrollLog(event, func(tview.Primitive) {})
		}
		return event
	})

	return view
}

func (view *CommandPrompt) History() *tview.TextView {
	return view.history
}
