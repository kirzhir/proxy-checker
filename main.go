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
	if err := root(os.Args[1:]); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}

		fmt.Printf("failed, %v\n", err.Error())
		os.Exit(1)
	}
}

func root(args []string) error {
	if len(args) < 1 {
		return errors.New("you must pass a sub-command")
	}

	cmds := []Runner{
		cmd.NewCliCommand(),
		cmd.NewServerCommand(),
		cmd.NewBotCommand(),
	}

	subcommand := os.Args[1]

	for _, c := range cmds {
		if c.Name() != subcommand {
			continue
		}

		if err := c.Init(os.Args[2:]); err != nil {
			return err
		}

		return c.Run(context.Background())
	}

	var join string
	for _, c := range cmds {
		join = fmt.Sprintf("%s '%s'", join, c.Name())
	}

	return fmt.Errorf("expected%s subcommands", join)
}
