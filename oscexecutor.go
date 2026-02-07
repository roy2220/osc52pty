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
	escapeSequenceBegin = []byte("\x1b]52;")
	escapeSequenceEnd   = []byte("\x07")
)

var copyToClipboardCmdLine = []string{"/usr/bin/pbcopy", "-pboard", "general"}

func setClipboard(copyToClipboardCmdLine []string, rawData []byte) error {
	// at this point, string will still be prepended by command and seperator
	// https://invisible-island.net/xterm/ctlseqs/ctlseqs.html
	//        The first, Pc, may contain zero or more characters from the
  //        set c , p , q , s , 0 , 1 , 2 , 3 , 4 , 5 , 6 , and 7 .  It is
  //        used to construct a list of selection parameters for
  //        clipboard, primary, secondary, select, or cut-buffers 0
  //        through 7 respectively, in the order given.  If the parameter
  //        is empty, xterm uses s 0 , to specify the configurable
  //        primary/clipboard selection and cut-buffer 0.
	// (thank you https://github.com/tmux/tmux/issues/4847#issuecomment-3863645137)

	// macOS *kind of* has multiple clipboards (manpage snippet from pbcopy below)
	// but the only one relevant to terminal use is general. therefore the argument from
	// the OSC can be safely dropped and the general clipboard used.
	//        -pboard {general | ruler | find | font}
  //            specifies which pasteboard to copy to or paste from.  If no pasteboard is given, the general pasteboard will be used by default.

	if idx := bytes.IndexByte(rawData, ';'); idx != -1 {
    rawData = rawData[idx+1:]
	}

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
