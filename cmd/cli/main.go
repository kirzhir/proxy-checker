package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log/slog"
	"os"
	"os/signal"
	"proxy-checker/internal/proxy"
	"sync"
	"syscall"
	"time"
)

type options struct {
	Output string `name:"output" default:"stdout"`
	Input  string `name:"input"  default:"stdin"`
	Debug  bool   `name:"debug" default:"false"`
}

func parseOpts(o *options) {
	flag.StringVar(&o.Output, "output", "stdout", "output file")
	flag.StringVar(&o.Input, "input", "stdin", "input file")
	flag.BoolVar(&o.Debug, "debug", false, "enable debug mode")

	flag.Parse()
}

func main() {
	var opts options

	parseOpts(&opts)
	setupLogger(&opts)

	slog.Info("starting", slog.String("in", opts.Input), slog.String("out", opts.Output))
	slog.Debug("debug enabled")

	ctx := context.Background()
	if err := run(ctx, opts); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}

		fmt.Printf("failed, %v\n", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, opts options) error {
	exit := make(chan error)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		<-stop
		cancel()

		<-time.After(1 * time.Second)
		exit <- nil
	}()

	eg, ctx := errgroup.WithContext(ctx)

	proxiesCh := runReading(ctx, opts.Input, eg)
	resultCh := runChecking(ctx, proxiesCh, eg)
	runWriting(ctx, opts.Output, resultCh, eg)

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

func runChecking(ctx context.Context, proxiesCh <-chan string, eg *errgroup.Group) <-chan string {
	res := make(chan string)

	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checker := proxy.Checker{Target: "http://checkip.amazonaws.com/", Timeout: 5 * time.Second}

			for {
				select {
				case ch, ok := <-proxiesCh:
					if !ok {
						return
					}

					if err := checker.Check(ctx, ch); err != nil {
						slog.Debug(err.Error())
					} else {
						res <- ch
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	eg.Go(func() error {
		wg.Wait()
		close(res)

		return nil
	})

	return res
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

func setupLogger(o *options) {
	level := slog.LevelInfo

	if o.Debug {
		level = slog.LevelDebug
	}

	logger := slog.New(
		slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
			ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
				if a.Key != slog.TimeKey {
					return a
				}

				a.Value = slog.StringValue(a.Value.Time().Format(time.DateTime))

				return a
			},
		}),
	)

	slog.SetDefault(logger)
}
