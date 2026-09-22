package tui

import (
	"fmt"
	"strings"

	"github.com/PawelReich/gogdb/client"

	"github.com/alecthomas/chroma/v2"
	"github.com/alecthomas/chroma/v2/lexers"
	"github.com/mitchellh/mapstructure"
	"github.com/rivo/tview"
)

type CodeView struct {
	Pane *tview.TextView

	gdbClient *client.GdbClient
	app       *tview.Application
}

func NewCodeView(app *tview.Application, client *client.GdbClient) *CodeView {

	view := &CodeView{gdbClient: client}

	textArea := tview.NewTextView()
	textArea.SetDynamicColors(true)
	textArea.SetScrollable(true)
	textArea.SetBorder(true)
	textArea.SetTitle(tview.Escape("Disassembly [none]"))

	view.Pane = textArea
	view.app = app

	return view
}

func (view *CodeView) Update(frame *client.StoppedFrame) {
	view.app.QueueUpdateDraw(func() {
		title := tview.Escape(fmt.Sprintf("Disassembly [%s]", frame.Architecture))

		view.Pane.SetTitle(title)

		disas := <-view.DisassembleAroundPC()
		if disas.Error != nil {
			panic(disas.Error)
		}
		view.Pane.SetText(view.PrettyPrintDisassembly(&disas.Result, frame.Address))
	})
}

type Instruction struct {
	Address string `mapstructure:"address"`
	Inst    string `mapstructure:"inst"`
	Func    string `mapstructure:"func"`
}

type GdbAsmDisassemblyPayload struct {
	AsmInsns []Instruction `mapstructure:"asm_insns"`
}

type AsyncDisassemblyResult struct {
	Result GdbAsmDisassemblyPayload
	Error  error
}

func (view *CodeView) PrettyPrintDisassembly(disas *GdbAsmDisassemblyPayload, pc string) string {
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

func (view *CodeView) DisassembleAroundPC() <-chan AsyncDisassemblyResult {

	ch := make(chan AsyncDisassemblyResult, 1)

	fut := view.gdbClient.SendAsync("data-disassemble", "-s", "$pc-128", "-e", "$pc+128", "--", "0")
	go func() {
		res := <-fut
		cmdres := AsyncDisassemblyResult{}

		if res.Error != nil {
			panic(res.Error)
			cmdres.Error = res.Error
		} else {
			err := mapstructure.Decode(res.Result["payload"], &cmdres.Result)
			if err != nil {
				panic(err)
				cmdres.Error = err
			}
		}

		ch <- cmdres
		close(ch)
	}()

	return ch
}
