package main

import (
	"time"

	"github.com/PawelReich/gogdb/client"
	"github.com/PawelReich/gogdb/tui"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	gdb, err := client.New()

	if err != nil {
		panic(err)
	}

	ui := tview.NewApplication()

	app := &tui.GoGdb{Ui: ui, Debugger: gdb}

	commandPrompt := tui.NewCommandPrompt(app)
	app.CommandPrompt = commandPrompt
	diassemblyView := tui.NewDisassemblyView(app)
	codeView := tui.NewCodeView(app)

	go func() {
		for notification := range gdb.Notifications() {

			if notification["class"] == "stopped" {

				frame, err := gdb.ParseFrame(notification)
				if err != nil {
					panic(err)
				}

				go func() {
					diassemblyView.Update(frame)
				}()
				go func() {
					codeView.Update(frame)
				}()
			}

			// app.Ui.QueueUpdateDraw(func() {
			// 	var color string
			// 	var str string
			//
			// 	switch notification["type"] {
			// 	case "console":
			// 		fallthrough
			// 	case "log":
			// 		color = "blue"
			// 		str = notification["payload"].(string)
			// 	case "error":
			// 		color = "red"
			// 		str = notification["payload"].(string)
			// 	default:
			// 		return
			// 	}
			// 	fmt.Fprintf(commandPrompt.History(), "[%s] %s", color, str)
			app.LogMap(notification)
			// })
		}
	}()

	app.Ui.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyCtrlQ:
			app.Ui.Stop()
		case tcell.KeyCtrlC:
			app.Debugger.Interrupt()
			return nil
		case tcell.KeyTab:
			if diassemblyView.Pane.HasFocus() {
				app.Ui.SetFocus(codeView.Pane)
			} else if codeView.Pane.HasFocus() {
				app.Ui.SetFocus(commandPrompt.Pane)
			} else {
				app.Ui.SetFocus(diassemblyView.Pane)
			}
		}

		return event
	})

	grid := tview.NewGrid()
	grid.SetRows(0, 20)
	grid.SetColumns(-1, -1)

	grid.AddItem(diassemblyView.Pane, 0, 0, 1, 1, 0, 0, false)
	grid.AddItem(codeView.Pane, 0, 1, 1, 1, 0, 0, false)
	grid.AddItem(commandPrompt.Pane, 1, 0, 1, 2, 0, 0, false)

	go func() {
		time.Sleep(100 * time.Millisecond)
		ret, err := gdb.Send("target-select", "remote", ":3333")

		if err != nil {
			panic(err)
		}

		app.LogMap(ret)
	}()

	err = app.Ui.SetRoot(grid, true).SetFocus(commandPrompt.Pane).Run()
	if err != nil {
		panic(err)
	}

}
