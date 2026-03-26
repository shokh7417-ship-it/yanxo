package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"yanxo/internal/bot"
	"yanxo/internal/config"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.FromEnv()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	app, err := bot.NewApp(cfg)
	if err != nil {
		log.Fatalf("app init: %v", err)
	}

	if err := app.Run(ctx); err != nil {
		log.Fatalf("run: %v", err)
	}
}

