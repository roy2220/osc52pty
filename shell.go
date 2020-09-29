package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"golang.org/x/crypto/ssh/terminal"
)

const bufferSize = 8 * 1024

type shell struct {
	Stdin  *os.File
	Stdout *os.File

	command  *exec.Cmd
	pty      *os.File
	cleanups []func(*shell)
}

func (s *shell) Open() (returnedErr error) {
	if s.Stdin == nil {
		s.Stdin = os.Stdin
	}

	if s.Stdout == nil {
		s.Stdout = os.Stdout
	}

	defer func() {
		if returnedErr != nil {
			s.Close()
		}
	}()

	if err := s.startPTY(); err != nil {
		return err
	}

	if err := s.makeTerminalRaw(); err != nil {
		return err
	}

	s.resizePTY()
	go s.pipeStdin()
	go s.pipeStdout()
	return nil
}

func (s *shell) Close() {
	for _, cleanup := range s.cleanups {
		cleanup(s)
	}

	s.cleanups = nil
}

func (s *shell) Wait() (int, error) {
	if err := s.command.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), nil
		}

		return 0, fmt.Errorf("wait shell failed: %v", err)
	}

	return 0, nil
}

func (s *shell) startPTY() error {
	s.command = makeShellCommand()
	var err error
	s.pty, err = pty.Start(s.command)

	if err != nil {
		return fmt.Errorf("start pty failed: %v", err)
	}

	s.cleanups = append(s.cleanups, func(s *shell) { s.pty.Close() })
	return nil
}

func (s *shell) makeTerminalRaw() error {
	oldState, err := terminal.MakeRaw(int(s.Stdin.Fd()))

	if err != nil {
		return fmt.Errorf("make terminal raw failed: %v", err)
	}

	s.cleanups = append(s.cleanups, func(*shell) { terminal.Restore(int(s.Stdin.Fd()), oldState) })
	return nil
}

func (s *shell) resizePTY() {
	signals := make(chan os.Signal, 1)
	s.cleanups = append(s.cleanups, func(*shell) { close(signals) })
	signal.Notify(signals, syscall.SIGWINCH)

	go func() {
		for range signals {
			if err := pty.InheritSize(s.Stdin, s.pty); err != nil {
				log.Printf("resize pty failed: %s", err)
			}
		}
	}()

	signals <- syscall.SIGWINCH
}

func (s *shell) pipeStdin() {
	buffer := make([]byte, bufferSize)

	for {
		n, err := s.Stdin.Read(buffer)

		if err != nil {
			if err == io.EOF {
				return
			}

			log.Printf("read stdin failed: %v", err)
			return
		}

		data := buffer[:n]

		if _, err := s.pty.Write(data); err != nil {
			log.Printf("write pty failed: %v", err)
			return
		}
	}
}

func (s *shell) pipeStdout() {
	buffer := make([]byte, bufferSize)

	oscExecutor := (&oscExecutor{
		InputDataHandler: func(data []byte) bool {
			if _, err := s.pty.Write(data); err != nil {
				log.Printf("write pty failed: %v", err)
				return false
			}

			return true
		},

		OutputDataHandler: func(data []byte) bool {
			if _, err := s.Stdout.Write(data); err != nil {
				log.Printf("write stdout failed: %v", err)
				return false
			}

			return true
		},
	}).Init()

	for {
		n, err := s.pty.Read(buffer)

		if err != nil {
			if err == io.EOF {
				return
			}

			log.Printf("read pty failed: %v", err)
			return
		}

		data := buffer[:n]

		if !oscExecutor.FeedData(data) {
			return
		}
	}
}

func makeShellCommand() *exec.Cmd {
	var shellCommand *exec.Cmd

	if args := os.Args[1:]; len(args) >= 1 {
		shellCommand = exec.Command(args[0], args[1:]...)
	} else {
		shellName := getShellName()
		shellCommand = exec.Command(shellName)
	}

	return shellCommand
}

func getShellName() string {
	shellName, ok := os.LookupEnv("SHELL")

	if !ok {
		shellName = "sh"
	}

	return shellName
}
