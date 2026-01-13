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
			city := update.Message.CommandArguments()
			h.userCities[update.Message.From.ID] = city
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, fmt.Sprintf("Город %s сохранен", city))
			msg.ReplyToMessageID = update.Message.MessageID
			h.bot.Send(msg)
			return
		default:
			log.Printf("New comand [%s] %s", update.Message.From.UserName, update.Message.Text)
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Такая команда не доступна")
			msg.ReplyToMessageID = update.Message.MessageID
			h.bot.Send(msg)
			return
		}
	}

	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	coordinates, err := h.owClient.Coordinates(update.Message.Text)
	if err != nil {
		log.Println(err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не смогли получить координаты")
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	weather, err := h.owClient.Weather(coordinates.Lat, coordinates.Lon)
	if err != nil {
		log.Println(err)
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не смогли получить погоду в этой местности")
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("Температура в %s: %d°C", update.Message.Text, int(math.Round(weather.Temp))),
	)
	msg.ReplyToMessageID = update.Message.MessageID

	h.bot.Send(msg)

}
