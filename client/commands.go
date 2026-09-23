package client

import (
	"fmt"
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

func (gdb *GdbClient) DisassembleAroundPC(byteRange int) <-chan AsyncDecodedResult[GdbAsmDisassemblyPayload] {
	return SendDecodeAsync[GdbAsmDisassemblyPayload](gdb, "data-disassemble", "-s", fmt.Sprintf("$pc-%d", byteRange/2), "-e", fmt.Sprintf("$pc+%d", byteRange/2), "--", "0")
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

func (gdb *GdbClient) GetCurrentStackFrame() <-chan AsyncDecodedResult[GdbStackFramePayload] {
	return SendDecodeAsync[GdbStackFramePayload](gdb, "stack-info-frame")
}
