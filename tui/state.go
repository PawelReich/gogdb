package tui

import (
	"github.com/PawelReich/gogdb/client"
	"github.com/rivo/tview"
)

type GoGdb struct {
	Ui            *tview.Application
	CommandPrompt *CommandPrompt
	Debugger      *client.GdbClient
}
