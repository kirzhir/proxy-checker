package cmd

import (
	"testing"
)

func TestCliCommand_Init(t *testing.T) {
	cliCmd := NewCliCommand()

	err := cliCmd.Init([]string{"-i", "stdin", "-o", "stdout", "-c", "10", "-v"})
	if err != nil {
		t.Fatalf("unexpected error during init: %v", err)
	}

	if cliCmd.input != "stdin" {
		t.Errorf("expected input to be input.txt, got %s", cliCmd.input)
	}

	if cliCmd.output != "stdout" {
		t.Errorf("expected output to be output.txt, got %s", cliCmd.output)
	}

	if cliCmd.concurrency != 10 {
		t.Errorf("expected concurrency to be 10, got %d", cliCmd.concurrency)
	}

	if !cliCmd.verbose {
		t.Errorf("expected verbose to be true, got false")
	}
}

func TestCliCommand_Name(t *testing.T) {
	cliCmd := NewCliCommand()
	expectedName := "cli"
	if cliCmd.Name() != expectedName {
		t.Errorf("expected command name to be %s, got %s", expectedName, cliCmd.Name())
	}
}
