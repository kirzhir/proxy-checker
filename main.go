package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"proxy-checker/cmd"
)

type Runner interface {
	Init([]string) error
	Run(ctx context.Context) error
	Name() string
}

func main() {

	cmds := []Runner{
		cmd.NewCliCommand(),
		cmd.NewServerCommand(),
		cmd.NewBotCommand(),
	}

	if err := root(os.Args[1:], cmds); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}

		fmt.Printf("failed, %v\n", err.Error())
		os.Exit(1)
	}
}

func root(args []string, cmds []Runner) error {
	if len(args) < 1 {
		return errors.New("you must pass a sub-command")
	}

	subcommand := args[0]

	for _, c := range cmds {
		if c.Name() != subcommand {
			continue
		}

		if err := c.Init(os.Args[2:]); err != nil {
			return err
		}

		return c.Run(context.Background())
	}

	return fmt.Errorf("unknown subcommand: %s", subcommand)
}
