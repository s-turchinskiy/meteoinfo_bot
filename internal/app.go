package internal

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/s-turchinskiy/meteoinfo_bot/internal/config"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/handlers/telegram_updates"
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
	var options []telegram_updates.OptionBotViaUpdates
	if cfg.URLProxy != nil {
		options = append(options, telegram_updates.WithProxy(cfg.URLProxy))
	} else {
		options = append(options, telegram_updates.WithoutProxy())
	}

	srvc, err := service.NewService(log, receiver)
	if err != nil {
		log.Fatalw("Error init service", "error", err.Error())
	}

	bot, err := telegram_updates.NewBotViaUpdates(srvc, cfg.TelegramBotToken, cfg.Timeout, log, options...)
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

func (a *App) Stop(timeout time.Duration) error {
	closer := closerutil.New(timeout)
	closer.Add(a.waitCloseGoroutines)
	closer.Add(a.bot.Close)
	err := closer.Shutdown()

	return err
}

func (a *App) waitCloseGoroutines(ctx context.Context) error {
	a.wg.Wait()
	return nil
}
