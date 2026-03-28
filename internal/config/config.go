package config

import (
	"errors"
	"net/url"
	"reflect"
	"strings"

	"github.com/caarlos0/env/v11"
)

type URLProxy1 struct{}

type (
	URLProxy       *url.URL
	OutputPathsLog []string
)

type Config struct {
	TelegramBotToken string         `env:"TOKEN"`            // Токен бота телеграмма
	URLProxy         URLProxy       `env:"PROXY"`            // Прокси для работы телеграмма
	Timeout          int            `env:"TIMEOUT"`          // Таймаут проверки сообщений в секундах
	OutputPathsLog   OutputPathsLog `env:"OUTPUT_PATHS_LOG"` // Куда будет выводиться лог
}

var ErrTokenIsEmpty = errors.New("token is empty")

func GetConfig() (*Config, error) {
	config := &Config{
		Timeout:        10,
		OutputPathsLog: []string{"bot.log", "stdout"},
	}

	err := env.ParseWithOptions(config, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(new(URLProxy)): func(incomingData string) (any, error) {
				return url.Parse(incomingData)
			},
			reflect.TypeOf(new(OutputPathsLog)): func(incomingData string) (any, error) {
				return strings.Split(incomingData, ","), nil
			},
		},
	})
	if err != nil {
		return nil, err
	}

	if config.TelegramBotToken == "" {
		return nil, ErrTokenIsEmpty
	}
	return config, nil
}
