package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
)

func (app *GoGdb) LogColor(message string, color string) {
	fmt.Fprintf(app.CommandPrompt.History(), "[%s]%s[-]\n", color, message)
}

func (app *GoGdb) LogInfo(message string) {
	app.LogColor(message, "grey")
}

func (app *GoGdb) LogError(message string) {
	app.LogColor(message, "red")
}

func (app *GoGdb) LogMap(value map[string]any) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(value)
	if err != nil {
		app.LogError(err.Error())
	}
	app.LogInfo(buf.String())
}
