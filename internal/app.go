package internal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/s-turchinskiy/meteoinfo_bot/internal/handlers"

	"github.com/s-turchinskiy/meteoinfo_bot/internal/config"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/service"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/utils/closerutil"
	"go.uber.org/zap"
)

type App struct {
	bot HandleProvider
	wg  sync.WaitGroup
}
type HandleProvider interface {
	Do(ctx context.Context)
	Close(ctx context.Context) error
}

func NewApp(cfg *config.Config, log *zap.SugaredLogger, receiver service.DataReceiver) (*App, error) {
	var options []handlers.OptionBotViaUpdates
	if cfg.URLProxy != nil {
		options = append(options, handlers.WithProxy(cfg.URLProxy))
	} else {
		options = append(options, handlers.WithoutProxy())
	}

	srvc, err := service.NewService(receiver, cfg.CitiesPath)
	if err != nil {
		log.Fatalw("Error init service", "error", err.Error())
	}

	bot, err := handlers.NewBotViaUpdates(srvc, cfg.TelegramBotToken, cfg.Timeout, log, options...)
	if err != nil {
		log.Fatalw(
			fmt.Errorf("connect to telegram wrong, error: %w", err).Error(),
			"proxy", cfg.URLProxy.Scheme+"://"+cfg.URLProxy.Host)
		return nil, fmt.Errorf("error run app %w", err)
	}

	log.Info("connect to telegram successful")

	return &App{
		bot: bot,
	}, nil
}

func (a *App) Run(ctx context.Context) {
	a.wg.Add(1)

	go func() {
		defer a.wg.Done()
		a.bot.Do(ctx)
	}()
}

func (a *App) Stop(timeout time.Duration, log *zap.SugaredLogger) error {
	closer := closerutil.New(timeout, log)
	closer.Add(a.waitCloseGoroutines)
	closer.Add(a.bot.Close)
	err := closer.Shutdown()

	return err
}

func (a *App) waitCloseGoroutines(ctx context.Context) error {
	a.wg.Wait()
	return nil
}
