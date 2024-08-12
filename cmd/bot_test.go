package cmd

import (
	"os"
	"testing"
)

func TestBotCommand_Init(t *testing.T) {
	botCmd := NewBotCommand()

	os.Setenv("TELEGRAM_API_TOKEN", "tg-token")

	err := botCmd.Init([]string{"-v", "-c", "10"})
	if err != nil {
		t.Fatalf("unexpected error during init: %v", err)
	}

	if !botCmd.verbose {
		t.Errorf("expected verbose to be true, got false")
	}

	if botCmd.concurrency != 10 {
		t.Errorf("expected concurrency to be 10, got %d", botCmd.concurrency)
	}
}

func TestBotCommand_Name(t *testing.T) {
	botCmd := NewBotCommand()
	expectedName := "bot"
	if botCmd.Name() != expectedName {
		t.Errorf("expected command name to be %s, got %s", expectedName, botCmd.Name())
	}
}
