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
	"syscall"
)

type BotCommand struct {
	fs  *flag.FlagSet
	cfg *config.Config

	verbose     bool
	concurrency uint
}

func NewBotCommand() *BotCommand {
	gc := &BotCommand{
		fs: flag.NewFlagSet("bot", flag.ContinueOnError),
	}

	gc.fs.BoolVar(&gc.verbose, "v", false, "verbosity mode")
	gc.fs.UintVar(&gc.concurrency, "c", 0, "concurrency limit")

	return gc
}

func (g *BotCommand) Name() string {
	return g.fs.Name()
}

func (g *BotCommand) Init(args []string) error {
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

	setupLogger(g.cfg)
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

	bot.Debug = g.verbose

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
		go handleUpdate(ctx, bot, g.cfg, update)
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

	resp, _ := proxy.NewChecker(cfg.ProxyChecker).AwaitCheck(ctx, proxiesCh)

	if _, err := bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, strings.Join(resp, "\n"))); err != nil {
		slog.Error("sending message failed", slog.String("error", err.Error()))
	}
}
