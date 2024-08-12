package main

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

type MockRunner struct {
	name string
	init func([]string) error
	run  func(ctx context.Context) error
}

func (m *MockRunner) Init(args []string) error {
	if m.init != nil {
		return m.init(args)
	}
	return nil
}

func (m *MockRunner) Run(ctx context.Context) error {
	if m.run != nil {
		return m.run(ctx)
	}
	return nil
}

func (m *MockRunner) Name() string {
	return m.name
}

func TestRoot(t *testing.T) {
	tests := []struct {
		name        string
		args        []string
		cmds        []Runner
		expectedErr error
	}{
		{
			name: "no subcommand provided",
			args: []string{},
			cmds: []Runner{
				&MockRunner{name: "cli"},
				&MockRunner{name: "server"},
			},
			expectedErr: errors.New("you must pass a sub-command"),
		},
		{
			name: "unknown subcommand",
			args: []string{"unknown"},
			cmds: []Runner{
				&MockRunner{name: "cli"},
				&MockRunner{name: "server"},
			},
			expectedErr: fmt.Errorf("unknown subcommand: %s", "unknown"),
		},
		{
			name: "successful cli command",
			args: []string{"cli"},
			cmds: []Runner{
				&MockRunner{name: "cli", run: func(ctx context.Context) error { return nil }},
				&MockRunner{name: "server"},
			},
			expectedErr: nil,
		},
		{
			name: "cli command init fails",
			args: []string{"cli"},
			cmds: []Runner{
				&MockRunner{name: "cli", init: func(args []string) error { return errors.New("init failed") }},
				&MockRunner{name: "server"},
			},
			expectedErr: errors.New("init failed"),
		},
		{
			name: "cli command run fails",
			args: []string{"cli"},
			cmds: []Runner{
				&MockRunner{name: "cli", run: func(ctx context.Context) error { return errors.New("run failed") }},
				&MockRunner{name: "server"},
			},
			expectedErr: errors.New("run failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := root(tt.args, tt.cmds)
			if err != nil && tt.expectedErr == nil {
				t.Fatalf("expected no error, but got: %v", err)
			}
			if err == nil && tt.expectedErr != nil {
				t.Fatalf("expected error: %v, but got none", tt.expectedErr)
			}
			if err != nil && tt.expectedErr != nil && err.Error() != tt.expectedErr.Error() {
				t.Fatalf("expected error: %v, but got: %v", tt.expectedErr, err)
			}
		})
	}
}
