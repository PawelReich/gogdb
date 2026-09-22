package tui

import (
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

	gdbClient   *client.GdbClient
	lastCommand string
}

func NewCommandPrompt(app *tview.Application, client *client.GdbClient) *CommandPrompt {

	view := &CommandPrompt{gdbClient: client}

	cmdHistory := tview.NewTextView()
	cmdHistory.SetDynamicColors(true)
	cmdHistory.SetScrollable(true)

	view.history = cmdHistory

	cmdPrompt := tview.NewInputField()
	cmdPrompt.SetLabel(Prompt)
	cmdPrompt.SetBackgroundColor(tcell.ColorDarkGrey)
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

		fut := client.SendAsync("interpreter-exec", "console", command)
		go func() {
			res := <-fut

			app.QueueUpdateDraw(func() {

				fmt.Fprintf(cmdHistory, "[grey]%s\n", res.Result)

				cmdPrompt.SetText("")
				cmdHistory.ScrollToEnd()
			})
		}()
	})

	view.input = cmdPrompt

	flex := tview.NewFlex().SetDirection(tview.FlexRow)
	flex.SetBorder(true)
	flex.SetTitle("Command Prompt")
	flex.AddItem(cmdHistory, 0, 1, false)
	flex.AddItem(cmdPrompt, 1, 0, true)

	flex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyPgUp:
			cmdHistory.ScrollTo(-20, 0)
		case tcell.KeyPgDn:
			cmdHistory.ScrollTo(20, 0)
		case tcell.KeyUp:
			cmdHistory.ScrollTo(-1, 0)
		case tcell.KeyDown:
			cmdHistory.ScrollTo(1, 0)
		}
		return event
	})

	view.Pane = flex

	return view
}

func (view *CommandPrompt) History() *tview.TextView {
	return view.history
}
