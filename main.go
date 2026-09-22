package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/PawelReich/gogdb/client"
	"github.com/PawelReich/gogdb/tui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func flatten(value map[string]any) ([]byte, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(value)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	gdb, err := client.New()

	if err != nil {
		panic(err)
	}

	app := tview.NewApplication()
	commandPrompt := tui.NewCommandPrompt(app, gdb)

	go func() {
		for notification := range gdb.Notifications() {

			app.QueueUpdateDraw(func() {
				var color string
				var str string

				switch notification["type"] {
				case "console":
					fallthrough
				case "log":
					color = "blue"
					str = notification["payload"].(string)
				case "error":
					color = "red"
					str = notification["payload"].(string)
				default:
					color = "grey"
					x, _ := flatten(notification)
					str = string(x)
				}
				fmt.Fprintf(commandPrompt.History(), "[%s] %s", color, str)
			})
		}
	}()

	ret, err := gdb.Send("target-select", "remote", ":3333")

	if err != nil {
		panic(err)
	}

	str, err := flatten(ret)
	if err != nil {
		panic(err)
	}

	fmt.Println(string(str))

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			app.Stop()
		case tcell.KeyCtrlC:
			gdb.Interrupt()
			return nil
		}
		return event
	})

	err = app.SetRoot(commandPrompt.Pane, true).SetFocus(commandPrompt.Pane).Run()
	if err != nil {
		panic(err)
	}

}
