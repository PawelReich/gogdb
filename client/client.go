package client

import (
	"fmt"
	"github.com/cyrus-and/gdb"
)

func handleGdbNotification(notification map[string]any) {
	fmt.Println(notification)
}

type GdbClient struct {
	gdb           *gdb.Gdb
	notifications chan string
}

func New() (*GdbClient, error) {
	gdbClient := &GdbClient{
		notifications: make(chan string),
	}

	gdb, err := gdb.New(handleGdbNotification)
	if err != nil {
		return nil, err
	}

	gdbClient.gdb = gdb

	_, err = gdbClient.gdb.Send("gdb-set", "mi-async", "on")

	if err != nil {
		gdbClient.Close()
		return nil, err
	}

	return gdbClient, nil
}

func (gdb *GdbClient) Send(operation string, args ...string) (map[string]any, error) {
	ret, err := gdb.gdb.Send(operation, args...)
	return ret, err
}

func (gdb *GdbClient) Close() {
	if gdb.notifications != nil {
		close(gdb.notifications)
	}
	if gdb.gdb != nil {
		gdb.gdb.Exit()
	}
}
