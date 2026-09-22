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

	view.Pane = flex

	return view
}

func (view *CommandPrompt) History() *tview.TextView {
	return view.history
}
