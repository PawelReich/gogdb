package client

import (
	"strings"
	"sync"

	"github.com/cyrus-and/gdb"
	"github.com/mitchellh/mapstructure"
)

type AsyncResult struct {
	Result map[string]any
	Error  error
}

type AsyncDecodedResult[T any] struct {
	Result T
	Error  error
}

type GdbClient struct {
	gdb *gdb.Gdb

	consoleMutex        sync.Mutex
	consoleCaptureMutex sync.Mutex
	consoleCaptured     *strings.Builder

	notifications chan map[string]any
}

func New() (*GdbClient, error) {
	gdbClient := &GdbClient{
		notifications: make(chan map[string]any, 512),
	}

	gdb, err := gdb.New(gdbClient.handleNotifications)
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

func SendDecodeAsync[T any](gdb *GdbClient, operation string, args ...string) <-chan AsyncDecodedResult[T] {
	ch := make(chan AsyncDecodedResult[T], 1)
	fut := gdb.SendAsync(operation, args...)
	go func() {
		res := <-fut
		cmdres := AsyncDecodedResult[T]{}

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

func (gdb *GdbClient) SendConsoleCommandAsync(command string) <-chan AsyncDecodedResult[string] {

	ch := make(chan AsyncDecodedResult[string], 1)

	go func() {
		defer close(ch)

		gdb.consoleMutex.Lock()
		defer gdb.consoleMutex.Unlock()

		var captured strings.Builder
		gdb.setConsoleCapture(&captured)
		defer gdb.setConsoleCapture(nil)

		_, err := gdb.Send("interpreter-exec", "console", command)

		gdb.setConsoleCapture(nil)
		result := captured.String()

		ch <- AsyncDecodedResult[string]{result, err}
	}()

	return ch
}

func (gdb *GdbClient) setConsoleCapture(capture *strings.Builder) {
	gdb.consoleCaptureMutex.Lock()
	gdb.consoleCaptured = capture
	gdb.consoleCaptureMutex.Unlock()
}

func (gdb *GdbClient) Close() {
	if gdb.gdb != nil {
		gdb.gdb.Exit()
	}
	if gdb.notifications != nil {
		close(gdb.notifications)
	}
}

func (gdb *GdbClient) handleNotifications(notification map[string]any) {
	if notification["type"] == "console" {
		payload := notification["payload"].(string)
		gdb.consoleCaptureMutex.Lock()
		if gdb.consoleCaptured != nil {
			gdb.consoleCaptured.WriteString(payload)
			gdb.consoleCaptureMutex.Unlock()
			return
		}
		gdb.consoleCaptureMutex.Unlock()
	}

	gdb.notifications <- notification
}

func (gdb *GdbClient) Notifications() <-chan map[string]any {
	return gdb.notifications
}
