package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
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
	_, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	fmt.Println(expandPath(opts.Input))

	return nil
}

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~") {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, path[1:]), nil
	}
	return path, nil
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
