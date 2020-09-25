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
	shell := shell{}

	if err := shell.Open(); err != nil {
		return 0, err
	}

	defer shell.Close()
	exitCode, err := shell.Wait()

	if err != nil {
		return 0, err
	}

	return exitCode, nil
}
