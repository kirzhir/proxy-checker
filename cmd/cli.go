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
	concurrency int
}

func NewCliCommand() *CliCommand {
	gc := &CliCommand{
		fs: flag.NewFlagSet("cli", flag.ContinueOnError),
	}

	gc.fs.StringVar(&gc.output, "o", "stdout", "output file")
	gc.fs.StringVar(&gc.input, "i", "stdin", "input file")
	gc.fs.IntVar(&gc.concurrency, "c", 0, "concurrency limit")
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

	if err := g.setConcurrencyEnv(); err != nil {
		return err
	}

	g.cfg = config.MustLoad()
	g.setupLogger()

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

		<-time.After(1 * time.Second)
		exit <- nil
	}()

	eg, ctx := errgroup.WithContext(ctx)

	proxiesCh := runReading(ctx, g.input, eg)
	resultCh, errorsCh := proxy.NewChecker(g.cfg.ProxyChecker).Check(ctx, proxiesCh)
	runWriting(ctx, g.output, resultCh, eg)

	eg.Go(func() error {
		return <-errorsCh
	})

	go func() {
		exit <- eg.Wait()
	}()

	return <-exit
}

func runWriting(ctx context.Context, out string, proxiesCh <-chan string, eg *errgroup.Group) {
	var writer proxy.Writer
	if out == "stdout" {
		writer = proxy.NewStdoutWriter()
	} else {
		writer = proxy.NewFileWriter(out)
	}

	eg.Go(func() error {
		return writer.Write(ctx, proxiesCh)
	})
}

func runReading(ctx context.Context, in string, eg *errgroup.Group) <-chan string {
	var reader proxy.Reader

	if in == "stdin" {
		reader = proxy.NewStdinReader()
	} else {
		reader = proxy.NewFileReader(in)
	}

	proxiesCh := make(chan string)
	eg.Go(func() error {
		defer close(proxiesCh)

		return reader.Read(ctx, proxiesCh)
	})

	return proxiesCh
}

func (g *CliCommand) setConcurrencyEnv() error {
	if g.concurrency <= 0 {
		return nil
	}

	return os.Setenv("CONCURRENCY", fmt.Sprintf("%d", g.concurrency))
}

func (g *CliCommand) setupLogger() {
	level := slog.LevelInfo

	if g.verbose {
		level = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(level)
}
