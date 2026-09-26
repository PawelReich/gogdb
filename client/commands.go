package client

import (
	"fmt"
	"strconv"
	"strings"
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

type GdbRegisterNamesPayload struct {
	RegisterNames []string `mapstructure:"register-names"`
}

type RegisterValue struct {
	Index string `mapstructure:"number"`
	Value string `mapstructure:"value"`
}

type GdbRegisterValuesPayload struct {
	RegisterValues []RegisterValue `mapstructure:"register-values"`
}

type Register struct {
	Name   string
	Number string
	Value  string
}

func (gdb *GdbClient) GetRegisters() <-chan AsyncDecodedResult[[]Register] {
	ch := make(chan AsyncDecodedResult[[]Register], 1)

	go func() {

		namesFut := SendDecodeAsync[GdbRegisterNamesPayload](gdb, "data-list-register-names")
		valuesFut := SendDecodeAsync[GdbRegisterValuesPayload](gdb, "data-list-register-values", "x")

		names := <-namesFut
		values := <-valuesFut

		if names.Error != nil {
			ch <- AsyncDecodedResult[[]Register]{Error: names.Error}
			return
		}
		if values.Error != nil {
			ch <- AsyncDecodedResult[[]Register]{Error: values.Error}
			return
		}

		length := len(values.Result.RegisterValues)
		registers := make([]Register, 0, length)

		for _, register := range values.Result.RegisterValues {
			registerIndex, err := strconv.Atoi(register.Index)
			if err != nil {
				panic(err)
			}

			registers = append(registers, Register{
				Name:  names.Result.RegisterNames[registerIndex],
				Value: register.Value,
			})
		}

		ch <- AsyncDecodedResult[[]Register]{Result: registers}
	}()

	return ch
}

func (gdb *GdbClient) GetCachedSymbol(sym string) <-chan AsyncDecodedResult[string] {
	ch := make(chan AsyncDecodedResult[string], 1)

	if resolvedSymbol, ok := gdb.symbolLookup[sym]; ok {
		ch <- AsyncDecodedResult[string]{Result: resolvedSymbol}
		close(ch)
		return ch
	}

	go func() {
		defer close(ch)

		res := <-gdb.GetSymbol(sym)
		if res.Error == nil {
			gdb.symbolLookup[sym] = res.Result
		}
		ch <- res
	}()

	return ch
}

func (gdb *GdbClient) GetSymbol(sym string) <-chan AsyncDecodedResult[string] {
	ch := make(chan AsyncDecodedResult[string], 1)

	go func() {
		defer close(ch)
		res := <-gdb.SendConsoleCommandAsync("info symbol " + sym)
		if res.Error != nil {
			ch <- res
			return
		}

		if strings.Contains(res.Result, "No symbol matches") {
			res.Result = ""
		}
		res.Result = strings.Split(res.Result, " in ")[0]
		ch <- res
	}()

	return ch
}
