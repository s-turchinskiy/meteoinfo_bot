package main

import (
	"context"
	"fmt"
	systemlog "log"
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

func main() {

	err := godotenv.Load("./cmd/.env")
	if err != nil {
		systemlog.Fatal("Error loading .env file", "error", err.Error())
	}

	cfg, err := config.GetConfig()
	if err != nil {
		systemlog.Fatal("Error get config", "error", err.Error())
	}

	log, err := logger.Initialize(cfg.OutputPathsLog)
	if err != nil {
		systemlog.Fatal(err)
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

	srvc, err := service.NewService(log)
	if err != nil {
		log.Fatalw("Error init service", "error", err.Error())
	}

	bot, err = telegram_updates.NewBotViaUpdates(srvc, cfg.TelegramBotToken, cfg.Timeout, log, options...)
	if err != nil {
		log.Fatalw(
			fmt.Errorf("connect to telegram wrong, error: %w", err).Error(),
			"proxy", cfg.URLProxy.Scheme+"://"+cfg.URLProxy.Host)
		return
	}

	log.Info("connect to telegram successful")

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
