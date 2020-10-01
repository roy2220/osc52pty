package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"log"
	"os/exec"
	"time"
)

type oscExecutor struct {
	inputDataSender dataSender
	parser          parser
}

var _ shellInterceptor = (*oscExecutor)(nil)

func (oe *oscExecutor) Init(inputDataSender, outputDataSender dataSender) *oscExecutor {
	oe.inputDataSender = inputDataSender
	oe.parser.Init(escapeSequenceBegin, escapeSequenceEnd, oe.handleDataToCopy, dataHandler(outputDataSender))
	return oe
}

func (oe *oscExecutor) HandleInputData(data []byte) bool {
	return oe.inputDataSender(data)
}

func (oe *oscExecutor) HandleOutputData(data []byte) bool {
	return oe.parser.FeedData(data)
}

func (oe *oscExecutor) handleDataToCopy(data []byte) bool {
	if err := setClipboard(copyToClipboardCmdLine, data); err != nil {
		log.Printf("set clipboard failed: %v", err)
	}

	return true
}

var (
	escapeSequenceBegin = []byte("\x1b]52;c;")
	escapeSequenceEnd   = []byte("\x07")
)

var copyToClipboardCmdLine = []string{"pbcopy"}

func setClipboard(copyToClipboardCmdLine []string, rawData []byte) error {
	buffer := make([]byte, base64.StdEncoding.DecodedLen(len(rawData)))
	n, err := base64.StdEncoding.Decode(buffer, rawData)

	if err != nil {
		return err
	}

	data := buffer[:n]
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, copyToClipboardCmdLine[0], copyToClipboardCmdLine[1:]...)
	command.Stdin = bytes.NewReader(data)

	if err := command.Run(); err != nil {
		return err
	}

	return nil
}
