package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/app"
	"github.com/lvlcn-t/raid-mate/app/config"
)

// version is set on build time
var version string

func main() {
	log := logger.NewLogger()
	ctx, cancel := logger.NewContextWithLogger(logger.IntoContext(context.Background(), log))
	defer cancel()

	var cfgPath string
	flag.StringVar(&cfgPath, "config", "", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.FatalContext(ctx, "Failed to load configuration", "error", err)
	}
	cfg.Version = version

	sigChan := make(chan os.Signal, 2)
	defer close(sigChan)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	application, err := app.New(cfg)
	if err != nil {
		log.FatalContext(ctx, "Failed to create application", "error", err)
	}

	cErr := make(chan error, 1)
	go func() {
		cErr <- application.Run(ctx)
	}()

	select {
	case <-sigChan:
		log.InfoContext(ctx, "Received signal, shutting down")
		err = application.Shutdown(ctx)
		<-cErr
	case err = <-cErr:
	}
	if err != nil {
		log.FatalContext(ctx, "Failed to run bot", "error", err)
	}
}
