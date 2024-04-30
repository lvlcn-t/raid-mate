package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/bot"
	"github.com/lvlcn-t/raid-mate/config"
)

// version is set on build time
var version string

func main() {
	_ = version
	log := logger.NewLogger()
	ctx := logger.IntoContext(context.Background(), log)

	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "Path to the configuration file")
	flag.ErrHelp = errors.New("usage: raid-mate -config <path>")
	flag.Parse()

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.FatalContext(ctx, "Failed to load configuration", "error", err)
	}

	err = cfg.Validate(ctx)
	if err != nil {
		log.FatalContext(ctx, "Invalid configuration", "error", err)
	}

	sigChan := make(chan os.Signal, 2)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	bot, err := bot.New(cfg.Bot)
	if err != nil {
		log.FatalContext(ctx, "Failed to create bot", "error", err)
	}

	cErr := make(chan error, 1)
	go func() {
		cErr <- bot.Run(ctx)
	}()

	select {
	case <-sigChan:
		log.InfoContext(ctx, "Received signal, shutting down")
		err = bot.Shutdown(ctx)
		<-cErr
	case err = <-cErr:
	}
	if err != nil {
		log.FatalContext(ctx, "Failed to run bot", "error", err)
	}
}
