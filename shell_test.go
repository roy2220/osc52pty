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
	_, err = stdinW.Write([]byte("hello world\n"))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	exitCode, err := s.Wait()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, 0, exitCode) {
		t.FailNow()
	}
	output, err := ioutil.ReadAll(stdoutR)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	if !assert.Equal(t, "hello world\r\nhello world\r\n", string(output)) {
		t.FailNow()
	}
}
