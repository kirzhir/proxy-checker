package cmd

import (
	"testing"
)

func TestServerCommand_Name(t *testing.T) {
	serverCmd := NewServerCommand()
	expectedName := "serve"
	if serverCmd.Name() != expectedName {
		t.Errorf("expected command name to be %s, got %s", expectedName, serverCmd.Name())
	}
}
