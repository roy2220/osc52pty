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
	InputDataHandler  func([]byte) bool
	OutputDataHandler func([]byte) bool

	parser parser
}

func (oe *oscExecutor) Init() *oscExecutor {
	oe.parser = parser{
		CaptureBegin:        escapeSequenceBegin,
		CaptureEnd:          escapeSequenceEnd,
		CapturedDataHandler: oe.handleCapturedData,
		IgnoredDataHandler:  oe.OutputDataHandler,
	}

	oe.parser.Init()
	return oe
}

func (oe *oscExecutor) FeedData(data []byte) bool {
	return oe.parser.FeedData(data)
}

func (oe *oscExecutor) handleCapturedData(data []byte) bool {
	if err := oe.setClipboard(data); err != nil {
		log.Printf("failed to set clipboard: %v", err)
	}

	return true
}

func (oe *oscExecutor) setClipboard(rawData []byte) error {
	data := make([]byte, base64.StdEncoding.DecodedLen(len(rawData)))

	if _, err := base64.StdEncoding.Decode(data, rawData); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "pbcopy")
	command.Stdin = bytes.NewReader(data)

	if err := command.Run(); err != nil {
		return err
	}

	return nil
}

var (
	escapeSequenceBegin = []byte("\x1b]52;c;")
	escapeSequenceEnd   = []byte("\x07")
)
