package cmd

import (
	"context"
	"errors"
	"flag"
	"golang.org/x/sync/errgroup"
	"html/template"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"proxy-checker/internal/config"
	http_server "proxy-checker/internal/http-server"
	"proxy-checker/internal/logger"
	"runtime"
	"syscall"
)

type ServerCommand struct {
	fs   *flag.FlagSet
	cfg  *config.Config
	temp *template.Template

	verbose     bool
	concurrency uint
}

func NewServerCommand() *ServerCommand {
	gc := &ServerCommand{
		fs: flag.NewFlagSet("serve", flag.ContinueOnError),
	}

	gc.fs.BoolVar(&gc.verbose, "v", false, "verbosity mode")
	gc.fs.UintVar(&gc.concurrency, "c", 0, "concurrency limit")

	return gc
}

func (g *ServerCommand) Name() string {
	return g.fs.Name()
}

func (g *ServerCommand) Init(args []string) error {
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
	g.temp = template.Must(template.ParseGlob("web/templates/*"))

	setupLogger(g.cfg)
	slog.Info("starting", slog.String("env", g.cfg.Env))
	slog.Debug("debug enabled")

	return nil
}

func (g *ServerCommand) Run(ctx context.Context) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	srv := &http.Server{
		Addr:         g.cfg.Address,
		Handler:      http_server.New(g.cfg, g.temp),
		ReadTimeout:  g.cfg.HTTPServer.Timeout,
		WriteTimeout: g.cfg.HTTPServer.Timeout,
		IdleTimeout:  g.cfg.HTTPServer.IdleTimeout,
	}

	var eg errgroup.Group

	eg.Go(func() error {
		return open(g.cfg)
	})

	eg.Go(func() error {
		slog.Info("listening on " + g.cfg.Address)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	eg.Go(func() error {
		<-stop

		shutdownCtx, cancel := context.WithTimeout(ctx, g.cfg.ShutdownTimeout)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	})

	return eg.Wait()
}

func setupLogger(cfg *config.Config) {
	var l *slog.Logger

	level := slog.LevelInfo

	if cfg.Verbose {
		level = slog.LevelDebug
	}

	switch cfg.Env {
	case "local":
		l = slog.New(logger.NewPrettyHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	default:
		l = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}))
	}

	slog.SetDefault(l)
}

func open(cfg *config.Config) error {
	if cfg.Env != "local" {
		return nil
	}

	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}

	return exec.Command(cmd, append(args, "http://"+cfg.Address)...).Start()
}
