package main

import (
	"io"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShell(t *testing.T) {
	var s shell
	stdinR, stdinW := io.Pipe()
	stdoutR, stdoutW := io.Pipe()
	err := s.Open(shellOptions{
		CmdLine: []string{"head", "-1"},
		Stdin:   stdinR,
		Stdout:  stdoutW,
	})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	defer s.Close()
	// Drain output concurrently because io.Pipe blocks the writer until the
	// reader consumes data, and waiting for the child first can lose PTY output.
	outputDone := make(chan struct {
		data []byte
		err  error
	}, 1)
	go func() {
		data, err := ioutil.ReadAll(stdoutR)
		outputDone <- struct {
			data []byte
			err  error
		}{data: data, err: err}
	}()

	_, err = stdinW.Write([]byte("hello world\n"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.NoError(t, stdinW.Close()) {
		t.FailNow()
	}
	exitCode, err := s.Wait()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, 0, exitCode) {
		t.FailNow()
	}
	output := <-outputDone
	if !assert.NoError(t, output.err) {
		t.FailNow()
	}
	if !assert.Equal(t, "hello world\r\nhello world\r\n", string(output.data)) {
		t.FailNow()
	}
}
