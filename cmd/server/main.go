package main

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/sync/errgroup"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"proxy-checker/internal/config"
	"syscall"
	"time"
)

func main() {

	cfg := config.MustLoad()
	setupLogger(cfg.Env)

	slog.Info("starting", slog.String("env", cfg.Env))
	slog.Debug("debug enabled")

	ctx := context.Background()
	if err := run(ctx, cfg); err != nil {
		if errors.Is(err, context.Canceled) {
			os.Exit(0)
		}

		fmt.Printf("failed, %v\n", err.Error())
		os.Exit(1)
	}
}

func run(ctx context.Context, cfg *config.Config) error {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      NewServer(cfg),
		ReadTimeout:  cfg.HTTPServer.Timeout,
		WriteTimeout: cfg.HTTPServer.Timeout,
		IdleTimeout:  cfg.HTTPServer.IdleTimeout,
	}

	var eg errgroup.Group

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

func NewServer(cfg *config.Config) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		cfg,
	)
	var handler http.Handler = mux
	//handler = someMiddleware(handler)
	//handler = someMiddleware2(handler)
	//handler = someMiddleware3(handler)
	return handler
}

func addRoutes(
	mux *http.ServeMux,
	cfg *config.Config,
	// tenantsStore        *TenantsStore,
	// commentsStore       *CommentsStore,
	// conversationService *ConversationService,
	// chatGPTService      *ChatGPTService,
	// authProxy           *authProxy
) {
	//mux.Handle("/api/v1/", handleTenantsGet(logger, tenantsStore))
	//mux.Handle("/oauth2/", handleOAuth2Proxy(logger, authProxy))
	//mux.HandleFunc("/healthz", handleHealthzPlease(logger))
	mux.Handle("/", http.NotFoundHandler())
	mux.HandleFunc("/test", func(w http.ResponseWriter, req *http.Request) {
		time.Sleep(2 * time.Second)
		fmt.Fprintln(w, "Hello world!")
	})
}

func setupLogger(env string) *slog.Logger {
	var logger *slog.Logger

	switch env {
	case "local":
		logger = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
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

	return logger
}
