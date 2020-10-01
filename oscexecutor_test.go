package main

import (
	"encoding/base64"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetClipboard(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "osc52pty")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	// defer os.Remove(tempFile.Name())
	text := "hello clipboard\n"
	encodedText := base64.StdEncoding.EncodeToString([]byte(text))
	err = setClipboard([]string{"tee", tempFile.Name()}, []byte(encodedText))
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	t.Log(tempFile.Name())
	data, err := ioutil.ReadFile(tempFile.Name())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.Equal(t, text, string(data))
}
