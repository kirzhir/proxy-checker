package main

import (
	"context"
	"flag"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"proxy-checker/internal/proxy"
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

	if err := run(opts); err != nil {
		if opts.Debug {
			log.Panicf("[ERROR] %v", err)
		}
		fmt.Printf("failed, %v\n", err.Error())
		os.Exit(1)
	}
}

func run(opts options) error {
	exit := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	proxiesCh := runReading(ctx, opts.Input)
	resultCh := runChecking(ctx, proxiesCh)

	runWriting(ctx, opts.Output, resultCh)

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
		slog.Warn("interrupt signal")
		<-stop
		cancel()

		<-time.After(1 * time.Second)
		exit <- nil
	}()

	return <-exit
}

func runWriting(ctx context.Context, out string, proxiesCh <-chan string) {

	writer := proxy.NewStdoutWriter()

	go func() {
		err := writer.Write(ctx, proxiesCh)
		if err != nil {
			slog.Error(err.Error())
		}
	}()
}

func runChecking(ctx context.Context, proxiesCh <-chan string) <-chan string {
	res := make(chan string)
	group, ctx := errgroup.WithContext(ctx)
	group.SetLimit(10)

	for i := 0; i < 10; i++ {
		group.Go(func() error {
			checker := proxy.Checker{Target: "http://checkip.amazonaws.com/", Timeout: 5 * time.Second}

			for {
				select {
				case ch, ok := <-proxiesCh:
					if !ok {
						return nil
					}

					if err := checker.Check(ctx, ch); err != nil {
						slog.Debug(err.Error())
					} else {
						res <- ch
					}
				case <-ctx.Done():
					return nil
				}
			}
		})
	}

	go func() {
		if err := group.Wait(); err != nil {
			slog.Error(err.Error())
		}

		close(res)
	}()

	return res
}

func runReading(ctx context.Context, in string) <-chan string {
	var reader proxy.Reader

	if in == "stdin" {
		reader = proxy.NewStdinReader()
	} else {
		reader = proxy.NewFileReader(in)
	}

	proxiesCh := make(chan string)
	go func() {
		defer close(proxiesCh)

		if err := reader.Read(ctx, proxiesCh); err != nil {
			slog.Error(err.Error())
		}
	}()

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
