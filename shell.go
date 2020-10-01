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
	cmdLine     []string
	stdin       io.ReadCloser
	stdout      io.WriteCloser
	interceptor shellInterceptor
	command     *exec.Cmd
	ptmx        *os.File
	cleanups    []func(*shell)
}

func (s *shell) Open(options shellOptions) (returnedErr error) {
	options.Sanitize()
	s.cmdLine = options.CmdLine
	s.stdin = options.Stdin
	s.stdout = options.Stdout

	defer func() {
		if returnedErr != nil {
			s.stdin.Close()
			s.stdout.Close()
		}
	}()

	if err := s.createShellInterceptor(options.ShellInterceptorFactory); err != nil {
		return err
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
	s.pipeStdin()
	s.pipeStdout()
	return nil
}

func (s *shell) Close() {
	for _, cleanup := range s.cleanups {
		cleanup(s)
	}

	s.cleanups = nil
}

func (s *shell) Wait() (_ int, returnedErr error) {
	defer func() {
		if returnedErr != nil {
			s.stdin.Close()
		}
	}()

	if err := s.command.Wait(); err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return exitError.ExitCode(), nil
		}

		return 0, fmt.Errorf("wait shell failed: %v", err)
	}

	return 0, nil
}

func (s *shell) createShellInterceptor(shellInterceptorFactory shellInterceptorFactory) error {
	inputDataSender := func(data []byte) bool {
		if _, err := s.ptmx.Write(data); err != nil {
			log.Printf("write ptmx failed: %v", err)
			return false
		}

		return true
	}

	outputDataSender := func(data []byte) bool {
		if _, err := s.stdout.Write(data); err != nil {
			log.Printf("write stdout failed: %v", err)
			return false
		}

		return true
	}

	shellInterceptor, err := shellInterceptorFactory(inputDataSender, outputDataSender)

	if err != nil {
		log.Printf("create shell interceptor failed: %v", err)
		return err
	}

	s.interceptor = shellInterceptor
	return nil
}

func (s *shell) startPTY() error {
	s.command = exec.Command(s.cmdLine[0], s.cmdLine[1:]...)
	var err error
	s.ptmx, err = pty.Start(s.command)

	if err != nil {
		return fmt.Errorf("start pty failed: %v", err)
	}

	s.cleanups = append(s.cleanups, func(s *shell) { s.ptmx.Close() })
	return nil
}

func (s *shell) makeTerminalRaw() error {
	stdin, ok := s.stdin.(*os.File)

	if !ok {
		return nil
	}

	oldState, err := terminal.MakeRaw(int(stdin.Fd()))

	if err != nil {
		return fmt.Errorf("make terminal raw failed: %v", err)
	}

	s.cleanups = append(s.cleanups, func(*shell) { terminal.Restore(int(stdin.Fd()), oldState) })
	return nil
}

func (s *shell) resizePTY() {
	stdin, ok := s.stdin.(*os.File)

	if !ok {
		return
	}

	signals := make(chan os.Signal, 1)
	s.cleanups = append(s.cleanups, func(*shell) { close(signals) })
	signal.Notify(signals, syscall.SIGWINCH)

	go func() {
		for range signals {
			if err := pty.InheritSize(stdin, s.ptmx); err != nil {
				log.Printf("resize pty failed: %s", err)
			}
		}
	}()

	signals <- syscall.SIGWINCH
}

func (s *shell) pipeStdin() {
	go func() {
		buffer := make([]byte, bufferSize)

		for {
			n, err := s.stdin.Read(buffer)

			if err != nil {
				if err == io.EOF {
					return
				}

				log.Printf("read stdin failed: %v", err)
				return
			}

			data := buffer[:n]

			if !s.interceptor.HandleInputData(data) {
				return
			}
		}
	}()
}

func (s *shell) pipeStdout() {
	go func() {
		defer s.stdout.Close()
		buffer := make([]byte, bufferSize)

		for {
			n, err := s.ptmx.Read(buffer)

			if err != nil {
				if err == io.EOF {
					return
				}

				log.Printf("read ptmx failed: %v", err)
				return
			}

			data := buffer[:n]

			if !s.interceptor.HandleOutputData(data) {
				return
			}
		}
	}()
}

type shellOptions struct {
	CmdLine                 []string
	Stdin                   io.ReadCloser
	Stdout                  io.WriteCloser
	ShellInterceptorFactory shellInterceptorFactory
}

func (so *shellOptions) Sanitize() {
	if len(so.CmdLine) == 0 {
		so.CmdLine = []string{getShellName()}
	}

	if so.Stdin == nil {
		so.Stdin = os.Stdin
	}

	if so.Stdout == nil {
		so.Stdout = os.Stdout
	}

	if so.ShellInterceptorFactory == nil {
		so.ShellInterceptorFactory = dummyShellInterceptorFactory
	}
}

type (
	shellInterceptorFactory func(inputDataSender, outputDataSender dataSender) (shellInterceptor shellInterceptor, ok error)
	dataSender              func(data []byte) (ok bool)

	shellInterceptor interface {
		HandleInputData([]byte) (ok bool)
		HandleOutputData([]byte) (ok bool)
	}
)

type dummyShellInterceptor struct {
	InputDataSender  dataSender
	OutputDataSender dataSender
}

func (dsi *dummyShellInterceptor) HandleInputData(data []byte) bool {
	return dsi.InputDataSender(data)
}

func (dsi *dummyShellInterceptor) HandleOutputData(data []byte) bool {
	return dsi.OutputDataSender(data)
}

func dummyShellInterceptorFactory(inputDataSender, outputDataSender dataSender) (shellInterceptor, error) {
	return &dummyShellInterceptor{
		InputDataSender:  inputDataSender,
		OutputDataSender: outputDataSender,
	}, nil
}

func getShellName() string {
	shellName, ok := os.LookupEnv("SHELL")

	if !ok {
		shellName = "sh"
	}

	return shellName
}
