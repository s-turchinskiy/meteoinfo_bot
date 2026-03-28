package telegram_updates

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"go.uber.org/zap"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/handlers"
	"github.com/s-turchinskiy/meteoinfo_bot/internal/service"
)

type BotViaUpdates struct {
	token   string
	bot     *tgbotapi.BotAPI
	updates tgbotapi.UpdatesChannel
	srvc    handlers.Servicer
	log     *zap.SugaredLogger
}

type OptionBotViaUpdates func(*BotViaUpdates) error

var ErrBotIsNotSpecified = errors.New("bot is not specified")

func NewBotViaUpdates(
	srvc handlers.Servicer,
	token string,
	timeout int,
	log *zap.SugaredLogger,
	opts ...OptionBotViaUpdates,
) (*BotViaUpdates, error) {
	b := &BotViaUpdates{
		srvc:  srvc,
		token: token,
		log:   log,
	}

	var err error
	for _, opt := range opts {
		err = opt(b)
		if err != nil {
			return nil, err
		}
	}

	if b.bot == nil {
		return nil, ErrBotIsNotSpecified
	}

	log.Infof("Authorized on account %s", b.bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeout

	b.updates = b.bot.GetUpdatesChan(u)

	return b, nil
}

func WithProxy(urlProxy *url.URL) OptionBotViaUpdates {
	return func(b *BotViaUpdates) error {
		var err error
		b.bot, err = tgbotapi.NewBotAPIWithClient(
			b.token,
			tgbotapi.APIEndpoint,
			&http.Client{
				Transport: &http.Transport{Proxy: http.ProxyURL(urlProxy)},
				Timeout:   5 * time.Second,
			},
		)
		if err != nil {
			return fmt.Errorf("error NewBotAPIWithClient WithProxy %w", err)
		}

		return nil
	}
}

func WithoutProxy() OptionBotViaUpdates {
	return func(b *BotViaUpdates) error {
		var err error
		b.bot, err = tgbotapi.NewBotAPI(b.token)

		return err
	}
}

func (b BotViaUpdates) Do(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case update := <-b.updates:
			b.processMessage(update)
		}
	}
}

func (b BotViaUpdates) processMessage(update tgbotapi.Update) {
	if update.Message != nil {
		b.log.Debugf("incoming message %v", update.Message)

		msg := b.getMsg(update.Message.Chat.ID, update.Message.MessageID, update.Message.Text)

		b.log.Debugf("sending message %v", update.Message)
		_, err := b.bot.Send(msg)
		if err != nil {
			b.log.Errorw("error sending message",
				"error", err.Error(),
				"incoming_msg", update.Message,
				"send_msg", msg,
			)
		}
	}
}

func (b BotViaUpdates) getMsg(chatID int64, messageID int, text string) tgbotapi.Chattable {
	switch text {
	case "/start":
		msg := tgbotapi.NewMessage(chatID, "Выберите город из списка ниже")
		msg.ReplyMarkup = handlers.Buttons
		return msg
	case "close":
		msg := tgbotapi.NewMessage(chatID, "Для дальнейшего использования напишите /start")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		return msg
	default:

		msg := tgbotapi.NewMessage(chatID, "")
		msg.ReplyToMessageID = messageID

		result, err := b.srvc.GetWeather(text)
		if err != nil {
			if errors.Is(err, service.ErrNoCity) {
				b.log.Infow(
					"error get data for city",
					"city", text,
					"chatID", chatID,
					"error", err.Error(),
				)

				msg.Text = "Для этого города не могу предоставить информацию, выберите город из списка ниже"
				msg.ReplyMarkup = handlers.Buttons
				return msg
			}

			b.log.Warnw(
				"error get data for city",
				"city", text,
				"chatID", chatID,
				"error", err.Error(),
			)
			msg.Text = "Ошибка получения данных"
			return msg
		}

		msg.Text = result
		return msg
	}
}

func (b BotViaUpdates) Close(ctx context.Context) error {
	b.bot.StopReceivingUpdates()
	return nil
}
