package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/s-turchinskiy/meteoinfo_bot/internal"

	"github.com/s-turchinskiy/meteoinfo_bot/internal/config"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/logger"
)

func main() {
	err := godotenv.Load("./cmd/.env")
	if err != nil {
		log.Fatal("Error loading .env file", "error", err.Error())
	}

	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal("Error get config", "error", err.Error())
	}

	loggingSystem, err := logger.Initialize(cfg.OutputPathsLog)
	if err != nil {
		log.Fatal(err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	app, err := internal.NewApp(cfg, loggingSystem)
	if err != nil {
		log.Fatal(err)
	}

	app.Run(ctx)

	<-ctx.Done()

	err = app.Stop(20 * time.Second)
	stop()

	if err != nil {
		log.Fatal(err)
	}
}
