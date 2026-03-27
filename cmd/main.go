package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/s-turchinskiy/meteoinfo_bot/internal/config"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/handlers"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/handlers/telegram_updates"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/logger"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/service"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/utils/closerutil"
)

func init() {
	err := logger.Initialize()
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	err := godotenv.Load("./cmd/.env")
	if err != nil {
		logger.Log.Fatalw("Error loading .env file", "error", err.Error())
	}

	cfg, err := config.GetConfig()
	if err != nil {
		logger.Log.Fatalw("Error get config", "error", err.Error())
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	closer := closerutil.New(20 * time.Second)

	var bot handlers.Handlerer

	var options []telegram_updates.OptionBotViaUpdates
	if cfg.URLProxy != nil {
		options = append(options, telegram_updates.WithProxy(cfg.URLProxy))
	} else {
		options = append(options, telegram_updates.WithoutProxy())
	}

	srvc, err := service.NewService()
	if err != nil {
		logger.Log.Fatalw("Error init service", "error", err.Error())
	}

	bot, err = telegram_updates.NewBotViaUpdates(srvc, cfg.TelegramBotToken, cfg.Timeout, options...)
	if err != nil {
		logger.Log.Fatalw(
			fmt.Errorf("connect to telegram wrong, error: %w", err).Error(),
			"proxy", cfg.URLProxy.Scheme+"://"+cfg.URLProxy.Host)
		return
	}

	logger.Log.Info("connect to telegram successful")

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		bot.Do(ctx)
	}()

	<-ctx.Done()

	closer.Add(bot.Close)
	err = closer.Shutdown()

	wg.Wait()
	stop()

	if err != nil {
		log.Fatal(err)
	}
}
