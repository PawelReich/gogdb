package client

import (
	"github.com/cyrus-and/gdb"
)

type AsyncResult struct {
	Result map[string]any
	Error  error
}

type GdbClient struct {
	gdb           *gdb.Gdb
	notifications chan map[string]any
}

func New() (*GdbClient, error) {
	gdbClient := &GdbClient{
		notifications: make(chan map[string]any, 512),
	}

	gdb, err := gdb.New(func(notification map[string]any) {
		gdbClient.notifications <- notification
	})
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

func (gdb *GdbClient) SendAsync(operation string, args ...string) <-chan AsyncResult {
	ch := make(chan AsyncResult, 1)
	go func() {
		res, err := gdb.gdb.Send(operation, args...)
		ch <- AsyncResult{Result: res, Error: err}
		close(ch)
	}()
	return ch
}

func (gdb *GdbClient) Close() {
	if gdb.notifications != nil {
		close(gdb.notifications)
	}
	if gdb.gdb != nil {
		gdb.gdb.Exit()
	}
}

func (gdb *GdbClient) Notifications() <-chan map[string]any {
	return gdb.notifications
}
