package client

func (gdb *GdbClient) Interrupt() {
	gdb.gdb.Interrupt()
}
