package main

import (
	"context"
	"flag"

	"github.com/lvlcn-t/loggerhead/logger"
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
	flag.Parse()

	cfg, err := config.Load(cfgPath)
	if err != nil {
		log.FatalContext(ctx, "Failed to load configuration", "error", err)
	}

	err = cfg.Validate(ctx)
	if err != nil {
		log.FatalContext(ctx, "Invalid configuration", "error", err)
	}

	log.InfoContext(ctx, "Configuration loaded", "config", cfg)
	// TODO: Implement the rest of the application
}
