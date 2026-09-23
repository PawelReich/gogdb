package tui

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/PawelReich/gogdb/client"
	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/rivo/tview"
)

type CodeView struct {
	Pane *tview.TextView

	app *GoGdb
}

func NewCodeView(app *GoGdb) *CodeView {

	textView := tview.NewTextView()
	textView.SetTitle("Code [none]")
	textView.SetDynamicColors(true)
	textView.SetScrollable(true)
	textView.SetBorder(true)

	return &CodeView{app: app, Pane: textView}
}

func (view *CodeView) Update(frame *client.StoppedFrame) {

	fut := view.app.Debugger.GetCurrentStackFrame()
	view.app.Ui.QueueUpdateDraw(func() {
		stackFrame := <-fut
		if stackFrame.Error != nil {
			panic(stackFrame.Error)
		}

		title := tview.Escape(fmt.Sprintf("Code [%s]", stackFrame.Result.Frame.Function))
		view.Pane.SetTitle(title)

		view.Pane.SetText(view.PrettyPrintCode(&stackFrame.Result.Frame))

		lineInt, err := strconv.Atoi(stackFrame.Result.Frame.FileLine)
		if err != nil {
			view.app.LogError(fmt.Sprintf("Error parsing file line from stack frame: %s", stackFrame.Result.Frame.FileLine))
			lineInt = view.Pane.GetFieldHeight()
		}
		_, _, _, viewHeight := view.Pane.GetRect()
		middlePosition := lineInt - viewHeight/2
		view.app.LogError(fmt.Sprintf("lineInt: %d, fieldheight: %d, middlePosition: %d", lineInt, viewHeight, middlePosition))

		view.Pane.ScrollTo(middlePosition, 0)
	})
}

func (view *CodeView) PrettyPrintCode(frame *client.GdbStackFrame) string {
	var sb strings.Builder

	code, err := os.ReadFile(frame.FilePath)
	if err != nil {
		return fmt.Sprintf("[red::b]Could not read file: %s", frame.FilePath)
	}

	lexer := lexers.Match(frame.FilePath)

	iterator, err := lexer.Tokenise(nil, string(code))

	for _, token := range iterator.Tokens() {
		var colorTag string
		switch token.Type.Category() {
		case chroma.Keyword:
			colorTag = "[yellow]"
		case chroma.Name:
			colorTag = "[cyan]"
		case chroma.Literal:
			colorTag = "[green]"
		case chroma.Text:
			colorTag = "[magenta]"
		case chroma.Punctuation:
			colorTag = "[white]"
		case chroma.Comment:
			// // Hack: Constants (e.g. 0x80e8153c  add r2, r1, #304    @ 0x130)
			// // Are improperly tokenized as comments, force them as `Punctuation` for now
			// if token.Value[0] == '#' {
			// 	colorTag = "[white]"
			// } else {
			colorTag = "[grey]"
			// }

		default:
			colorTag = fmt.Sprintf("[-]\n:FAILED %s = %s\n", token.Type.Category(), tview.Escape(token.Value))
		}
		sb.WriteString(colorTag)
		sb.WriteString(tview.Escape(token.Value))
	}

	sb.WriteString("[::B]\n")

	return sb.String()
}
