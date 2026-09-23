package client

import "github.com/mitchellh/mapstructure"

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

func (gdb *GdbClient) DisassembleAroundPC() <-chan AsyncDisassemblyResult {
	ch := make(chan AsyncDisassemblyResult, 1)

	fut := gdb.SendAsync("data-disassemble", "-s", "$pc-128", "-e", "$pc+128", "--", "0")
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