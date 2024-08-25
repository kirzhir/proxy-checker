package cmd

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"os/signal"
	"proxy-checker/internal/config"
	"proxy-checker/internal/proxy"
	"syscall"
	"time"
)

type CliCommand struct {
	fs  *flag.FlagSet
	cfg *config.Config

	output      string
	input       string
	verbose     bool
	concurrency uint
}

func NewCliCommand() *CliCommand {
	gc := &CliCommand{
		fs: flag.NewFlagSet("cli", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.output, "o", "stdout", "output file")
	gc.fs.StringVar(&gc.input, "i", "stdin", "input file")
	gc.fs.UintVar(&gc.concurrency, "c", 0, "concurrency limit")
	gc.fs.BoolVar(&gc.verbose, "v", false, "verbosity mode")

	return gc
}

func (g *CliCommand) Name() string {
	return g.fs.Name()
}

func (g *CliCommand) Init(args []string) error {
	if err := g.fs.Parse(args); err != nil {
		return err
	}

	if err := setConcurrencyEnv(g.concurrency); err != nil {
		return err
	}

	if err := setVerbosityMode(g.verbose); err != nil {
		return err
	}

	g.cfg = config.MustLoad()

	setupLogger(g.cfg)
	checkInternetConnection()

	slog.Info("starting", slog.String("in", g.input), slog.String("out", g.output))
	slog.Debug("debug enabled")

	return nil
}

func (g *CliCommand) Run(ctx context.Context) error {
	exit := make(chan error)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		<-stop
		cancel()

		<-time.After(g.cfg.ShutdownTimeout)
		exit <- nil
	}()

	eg, ctx := errgroup.WithContext(ctx)

	proxiesCh := make(chan string)
	eg.Go(func() error {
		return proxy.NewReader(g.input).Read(ctx, proxiesCh)
	})

	resultCh, errorsCh := proxy.NewChecker(g.cfg.ProxyChecker).Check(ctx, proxiesCh)
	eg.Go(func() error {
		return proxy.NewWriter(g.output).Write(ctx, resultCh)
	})

	eg.Go(func() error {
		return <-errorsCh
	})

	go func() {
		exit <- eg.Wait()
	}()

	return <-exit
}

func setConcurrencyEnv(concurrency uint) error {
	if concurrency <= 0 {
		return nil
	}

	return os.Setenv("CONCURRENCY", fmt.Sprintf("%d", concurrency))
}

func setVerbosityMode(verbose bool) error {
	if !verbose {
		return nil
	}

	return os.Setenv("VERBOSE", fmt.Sprintf("%t", verbose))
}
