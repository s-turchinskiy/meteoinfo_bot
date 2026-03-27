package handlers

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

var Buttons = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("Москва"),
		tgbotapi.NewKeyboardButton("Сочи"),
	),
)
