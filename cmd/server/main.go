package main

import (
	"context"
	"errors"
	"fmt"
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

func main() {

	cfg := config.MustLoadFile()
	temp := template.Must(template.ParseGlob("web/templates/*"))
	setupLogger(cfg.Env)

	slog.Info("starting", slog.String("env", cfg.Env))
	slog.Debug("debug enabled")

	ctx := context.Background()
	if err := run(ctx, cfg, temp); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}

		fmt.Printf("failed, %v\n", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config, temp *template.Template) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      http_server.New(cfg, temp),
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	var eg errgroup.Group

	eg.Go(func() error {
		return open(cfg)
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

func setupLogger(env string) {
	var logger *slog.Logger

	switch env {
	case "local":
		logger = slog.Default()
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "dev":
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case "prod":
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	default:
		logger = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
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
