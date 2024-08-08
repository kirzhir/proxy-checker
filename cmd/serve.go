package cmd

import (
	"context"
	"errors"
	"flag"
	"golang.org/x/sync/errgroup"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"proxy-checker/internal/config"
	http_server "proxy-checker/internal/http-server"
	"runtime"
	"syscall"
	"time"
)

type ServerCommand struct {
	fs   *flag.FlagSet
	cfg  *config.Config
	temp *template.Template

	verbose bool
}

func NewServerCommand() *ServerCommand {
	gc := &ServerCommand{
		fs: flag.NewFlagSet("serve", flag.ContinueOnError),
	}

	gc.fs.BoolVar(&gc.verbose, "v", false, "verbosity mode")

	return gc
}

func (g *ServerCommand) Name() string {
	return g.fs.Name()
}

func (g *ServerCommand) Init(args []string) error {
	if err := g.fs.Parse(args); err != nil {
		return err
	}

	g.cfg = config.MustLoad()
	g.temp = template.Must(template.ParseGlob("web/templates/*"))

	g.setupLogger()
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
		log.Printf("listening on %s\n", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}

		return nil
	})

	eg.Go(func() error {
		<-stop

		shutdownCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		return srv.Shutdown(shutdownCtx)
	})

	return eg.Wait()
}

func (g *ServerCommand) setupLogger() {
	var logger *slog.Logger

	level := slog.LevelInfo

	if g.verbose {
		level = slog.LevelDebug
		slog.SetLogLoggerLevel(level)
	}

	switch g.cfg.Env {
	case "local":
		logger = slog.Default()
	default:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level}),
		)
	}

	slog.SetDefault(logger)
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
