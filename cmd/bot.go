package cmd

import (
	"context"
	"flag"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"proxy-checker/internal/config"
	"proxy-checker/internal/proxy"
	"strings"
	"sync"
	"syscall"
)

type BotCommand struct {
	fs  *flag.FlagSet
	cfg *config.Config

	debug bool
}

func NewBotCommand() *BotCommand {
	gc := &BotCommand{
		fs: flag.NewFlagSet("bot", flag.ContinueOnError),
	}

	gc.fs.BoolVar(&gc.debug, "debug", false, "enable debug mode")

	return gc
}

func (g *BotCommand) Name() string {
	return g.fs.Name()
}

func (g *BotCommand) Init(args []string) error {
	if err := g.fs.Parse(args); err != nil {
		return err
	}

	g.setupLogger()

	g.cfg = config.MustLoadEnv()
	if g.cfg.APIToken == "" {
		return fmt.Errorf("TELEGRAM_API_TOKEN is not set")
	}

	slog.Info("starting bot...")
	slog.Debug("debug enabled")

	return nil
}

func (g *BotCommand) Run(ctx context.Context) error {
	bot, err := tgbotapi.NewBotAPI(g.cfg.APIToken)
	if err != nil {
		return err
	}

	bot.Debug = g.debug

	go func() {
		stop := make(chan os.Signal, 1)
		signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

		<-stop
		bot.StopReceivingUpdates()
	}()

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	for update := range bot.GetUpdatesChan(u) {
		if update.Message == nil {
			continue
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
		go handleUpdate(context.Background(), bot, g.cfg, update)
	}

	return err
}

func handleUpdate(ctx context.Context, bot *tgbotapi.BotAPI, cfg *config.Config, update tgbotapi.Update) {
	proxiesCh := make(chan string, cfg.Concurrency)

	go func() {
		req := strings.Split(update.Message.Text, "\n")

		for _, p := range req {
			proxiesCh <- p
		}
		close(proxiesCh)
	}()

	res := make(chan string, cfg.Concurrency)
	wg := &sync.WaitGroup{}
	for i := 0; i < cfg.Concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			checker := proxy.NewChecker(cfg.ProxyChecker)

			for {
				select {
				case ch, ok := <-proxiesCh:
					if !ok {
						return
					}

					if p, err := checker.Check(ctx, ch); err != nil {
						slog.Debug(err.Error())
					} else {
						res <- p
					}
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	var resp []string
	for p := range res {
		resp = append(resp, p)
	}

	if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, strings.Join(resp, "\n"))); err != nil {
		slog.Error("sending message failed", slog.String("error", err.Error()))
	}
}

func (g *BotCommand) setupLogger() {
	level := slog.LevelInfo

	if g.debug {
		level = slog.LevelDebug
	}

	slog.SetLogLoggerLevel(level)
}
