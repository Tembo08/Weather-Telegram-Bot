package handler

import (
	"fmt"
	"log"
	"math"

	"github.com/Tembo08/my-weather-bot/clients/openweather"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Handler struct {
	bot        *tgbotapi.BotAPI
	owClient   *openweather.OpenWeatherClient
	userCities map[int64]string
}

func New(bot *tgbotapi.BotAPI, owClient *openweather.OpenWeatherClient) *Handler {
	return &Handler{
		bot:        bot,
		owClient:   owClient,
		userCities: make(map[int64]string),
	}
}

func (h Handler) Start() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := h.bot.GetUpdatesChan(u)

	for update := range updates {
		h.handleUpdate(update)
	}
}

func (h *Handler) handleUpdate(update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	if update.Message.IsCommand() {
		switch update.Message.Command() {
		case "city":
			h.handleSetCity(update)
			return
		case "weather":
			h.handleSendWeather(update)
			return
		default:
			h.handleUnknownCommand(update)
			return
		}
	}
}

func (h Handler) handleSetCity(update tgbotapi.Update) {
	city := update.Message.CommandArguments()
	h.userCities[update.Message.From.ID] = city
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Город %s сохранен", city))
	msg.ReplyToMessageID = update.Message.MessageID
	h.bot.Send(msg)

}

func (h *Handler) handleSendWeather(update tgbotapi.Update) {
	city, ok := h.userCities[update.Message.From.ID]
	if !ok {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Сначала установите город с помощью команды /city\nВот так '/city Москва'")
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	coordinates, err := h.owClient.Coordinates(city)
	if err != nil {
		log.Printf("error owClient.Coordinates: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не смогли получить координаты")
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	weather, err := h.owClient.Weather(coordinates.Lat, coordinates.Lon)
	if err != nil {
		log.Printf("error owClient.Weather: %v", err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не смогли получить погоду в этой местности")
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("Температура в %s: %d°C", city, int(math.Round(weather.Temp))),
	)
	msg.ReplyToMessageID = update.Message.MessageID

	h.bot.Send(msg)

}

func (h *Handler) handleUnknownCommand(update tgbotapi.Update) {
	log.Printf("Unrnown command [%s] %s", update.Message.From.UserName, update.Message.Text)
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такая команда не доступна")
	msg.ReplyToMessageID = update.Message.MessageID
	h.bot.Send(msg)
}
