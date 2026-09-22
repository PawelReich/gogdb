package tui

import (
	"fmt"
	"strings"

	"github.com/PawelReich/gogdb/client"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/rivo/tview"
)

type DisassemblyView struct {
	Pane *tview.TextView

	gdbClient *client.GdbClient
	app       *tview.Application
}

func NewDisassemblyView(app *tview.Application, client *client.GdbClient) *DisassemblyView {

	view := &DisassemblyView{gdbClient: client}

	textArea := tview.NewTextView()
	textArea.SetDynamicColors(true)
	textArea.SetScrollable(true)
	textArea.SetBorder(true)
	textArea.SetTitle(tview.Escape("Disassembly [none]"))

	view.Pane = textArea
	view.app = app

	return view
}

func (view *DisassemblyView) Update(frame *client.StoppedFrame) {
	view.app.QueueUpdateDraw(func() {
		title := tview.Escape(fmt.Sprintf("Disassembly [%s]", frame.Architecture))

		view.Pane.SetTitle(title)

		disas := <-view.gdbClient.DisassembleAroundPC()
		if disas.Error != nil {
			panic(disas.Error)
		}
		view.Pane.SetText(view.PrettyPrintDisassembly(&disas.Result, frame.Address))
	})
}

func (view *DisassemblyView) PrettyPrintDisassembly(disas *client.GdbAsmDisassemblyPayload, pc string) string {
	var sb strings.Builder

	lexer := lexers.Get("gas")

	for _, insn := range disas.AsmInsns {

		sb.WriteString("[grey::i]")
		sb.WriteString(insn.Address)

		sb.WriteString("[white::I]")
		if insn.Address == pc {
			sb.WriteString("->[::b]")
		} else {
			sb.WriteString("  ")
		}

		iterator, err := lexer.Tokenise(nil, insn.Inst)
		if err != nil {
			panic(err)
		}
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
				// Hack: Constants (e.g. 0x80e8153c  add r2, r1, #304    @ 0x130)
				// Are improperly tokenized as comments, force them as `Punctuation` for now
				if token.Value[0] == '#' {
					colorTag = "[white]"
				} else {
					colorTag = "[grey]"
				}

			default:
				colorTag = fmt.Sprintf("[-]\n:FAILED %s = %s\n", token.Type.Category(), tview.Escape(token.Value))
			}
			sb.WriteString(colorTag)
			sb.WriteString(tview.Escape(token.Value))
		}

		sb.WriteString("[::B]\n")
	}

	return sb.String()
}
