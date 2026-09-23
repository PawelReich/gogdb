package client

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

func (gdb *GdbClient) Interrupt() {
	gdb.gdb.Interrupt()
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

func (gdb *GdbClient) DisassembleAroundPC(byteRange int) <-chan AsyncDisassemblyResult {
	ch := make(chan AsyncDisassemblyResult, 1)

	fut := gdb.SendAsync("data-disassemble", "-s", fmt.Sprintf("$pc-%d", byteRange/2), "-e", fmt.Sprintf("$pc+%d", byteRange/2), "--", "0")
	go func() {
		res := <-fut
		cmdres := AsyncDisassemblyResult{}

		if res.Error != nil {
			cmdres.Error = res.Error
		} else {
			cmdres.Error = mapstructure.Decode(res.Result["payload"], &cmdres.Result)
		}

		ch <- cmdres
		close(ch)
	}()

	return ch
}

type GdbStackFrame struct {
	Address  string `mapstructure:"address"`
	Function string `mapstructure:"func"`
	FilePath string `mapstructure:"fullname"`
	FileLine string `mapstructure:"line"`
}

type GdbStackFramePayload struct {
	Frame GdbStackFrame `mapstructure:"frame"`
}

type AsyncStackFrameResult struct {
	Result GdbStackFramePayload
	Error  error
}

func (gdb *GdbClient) GetCurrentStackFrame() <-chan AsyncStackFrameResult {
	ch := make(chan AsyncStackFrameResult, 1)

	fut := gdb.SendAsync("stack-info-frame")
	go func() {
		res := <-fut
		cmdres := AsyncStackFrameResult{}

		if res.Error != nil {
			cmdres.Error = res.Error
		} else {
			cmdres.Error = mapstructure.Decode(res.Result["payload"], &cmdres.Result)
		}

		ch <- cmdres
		close(ch)
	}()

	return ch
}
