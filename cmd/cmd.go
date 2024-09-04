package cmd

import (
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/exec"
	"proxy-checker/internal/config"
	"proxy-checker/internal/logger"
	"runtime"
	"time"
)

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

func checkInternetConnection() {
	dialer := &net.Dialer{Timeout: 1 * time.Second}

	conn, err := dialer.Dial("tcp", "8.8.8.8:53")
	if err != nil {
		log.Panicf("no internet connection: %s", err)
	}
	defer conn.Close()
}
