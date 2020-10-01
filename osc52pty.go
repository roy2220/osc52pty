package main

import (
	"log"
	"os"
)

func main() {
	exitCode, err := runShell()

	if err != nil {
		log.Fatal(err)
	}

	os.Exit(exitCode)
}

func runShell() (int, error) {
	var shell shell

	shellOptions := shellOptions{
		ShellInterceptorFactory: func(inputDataSender, outputDataSender dataSender) (shellInterceptor, error) {
			return new(oscExecutor).Init(inputDataSender, outputDataSender), nil
		},
	}

	if err := shell.Open(shellOptions); err != nil {
		return 0, err
	}

	defer shell.Close()
	exitCode, err := shell.Wait()

	if err != nil {
		return 0, err
	}

	return exitCode, nil
}
