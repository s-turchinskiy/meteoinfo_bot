package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type WeatherService interface {
	GetWeather(city string) (result string, err error)
}

var Buttons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Москва"),
		tgbotapi.NewKeyboardButton("Сочи"),
	),
)
