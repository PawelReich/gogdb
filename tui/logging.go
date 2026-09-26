package tui

import (
	"bytes"
	"encoding/json"
	"sync"
)

var logMu sync.Mutex

func (app *GoGdb) LogColor(color string, message string, args ...any) {
	logMu.Lock()
	app.CommandPrompt.LogColorf(color, message, args...)
	logMu.Unlock()
}

func (app *GoGdb) LogInfo(message string, args ...any) {
	app.LogColor("white", message, args...)
}

func (app *GoGdb) LogDebug(message string, args ...any) {
	app.LogColor("grey", message, args...)
}

func (app *GoGdb) LogError(message string, args ...any) {
	app.LogColor("red", message, args...)
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
