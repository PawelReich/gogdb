package tui

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"
)

var logMu sync.Mutex

func (app *GoGdb) LogColor(message string, color string) {
	logMu.Lock()
	fmt.Fprintf(app.CommandPrompt.History(), "[%s]%s[-]\n", color, message)
	logMu.Unlock()
}

func (app *GoGdb) LogInfo(message string) {
	app.LogColor(message, "grey")
}

func (app *GoGdb) LogError(message string) {
	app.LogColor(message, "red")
}

func (app *GoGdb) LogMap(value map[string]any) {
	var buf bytes.Buffer

	encoder := json.NewEncoder(&buf)
	encoder.SetIndent("", "\t")

	err := encoder.Encode(value)

	if err != nil {
		app.LogError(err.Error())
	}
	app.LogInfo(buf.String())
}
