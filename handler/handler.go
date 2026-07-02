package handler

import (
	"context"
	"fmt"
	"log"
	"math"

	"github.com/Tembo08/my-weather-bot/clients/openweather"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type userRepository interface {

	//получить город юзера
	GetUserCity(ctx context.Context, userID int64) (string, error)
	//создать юзера
	CreateUser(ctx context.Context, userID int64) error
	//изменить город юзера
	UpdateCity(ctx context.Context, userID int64, city string) error
}

type Handler struct {
	bot      *tgbotapi.BotAPI
	owClient *openweather.OpenWeatherClient
	userRepo userRepository
}

func New(
	bot *tgbotapi.BotAPI,
	owClient *openweather.OpenWeatherClient,
	userRepo userRepository,
) *Handler {

	return &Handler{
		bot:      bot,
		owClient: owClient,
		userRepo: userRepo,
	}
}

func (h *Handler) Start() {
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
		case "start":
			h.handleStart(update)
			return
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

//Приветствие и создание пользователя в БД

func (h *Handler) handleStart(update tgbotapi.Update) {
	ctx := context.Background()
	userID := update.Message.From.ID

	err := h.userRepo.CreateUser(ctx, userID)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
	}

	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"👋 Привет! Я бот погоды.\n\n"+
			"Доступные команды:\n"+
			"/city <город> — установить город\n"+
			"/weather — показать погоду",
	)
	msg.ReplyToMessageID = update.Message.MessageID
	h.bot.Send(msg)
}

func (h *Handler) handleSetCity(update tgbotapi.Update) {
	ctx := context.Background()
	userID := update.Message.From.ID
	city := update.Message.CommandArguments()

	if city == "" {
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"❌ Укажите город. Например: /city Москва",
		)
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	//Проверяем существует ли город

	_, err := h.owClient.Coordinates(city)
	if err != nil {
		log.Printf("City not found: %s,%v", city, err)
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			fmt.Sprintf("❌ Город '%s' не найден. Проверьте название.", city),
		)
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	err = h.userRepo.UpdateCity(ctx, userID, city)
	if err != nil {
		log.Printf("Falied to update city: %v", err)
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"❌ Произошла ошибка при сохранении города",
		)
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("✅ Город установлен: %s", city),
	)
	msg.ReplyToMessageID = update.Message.MessageID
	h.bot.Send(msg)

}

// handleSendWeather — отправка погоды (город из БД)
func (h *Handler) handleSendWeather(update tgbotapi.Update) {
	ctx := context.Background()
	userID := update.Message.From.ID

	// Получаем город из БД
	city, err := h.userRepo.GetUserCity(ctx, userID)
	if err != nil {
		log.Printf("Failed to get user city: %v", err)
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"❌ Произошла ошибка при получении города",
		)
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	// Если город не установлен
	if city == "" {
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"❌ Сначала установите город с помощью команды /city\nНапример: /city Москва",
		)
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	// Получаем координаты города
	coordinates, err := h.owClient.Coordinates(city)
	if err != nil {
		log.Printf("error owClient.Coordinates: %v", err)
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"❌ Не удалось найти город",
		)
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	// Получаем погоду
	weather, err := h.owClient.Weather(coordinates.Lat, coordinates.Lon)
	if err != nil {
		log.Printf("error owClient.Weather: %v", err)
		msg := tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"❌ Не удалось получить погоду",
		)
		msg.ReplyToMessageID = update.Message.MessageID
		h.bot.Send(msg)
		return
	}

	// Формируем ответ
	emoji := getWeatherEmoji(weather.Temp)
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		fmt.Sprintf("🌤️ Погода в %s:\n%s\n🌡️ Температура: %d°C",
			city,
			emoji,
			int(math.Round(weather.Temp)),
		),
	)
	msg.ReplyToMessageID = update.Message.MessageID
	h.bot.Send(msg)
}

// getWeatherEmoji — возвращает эмодзи в зависимости от температуры
func getWeatherEmoji(temp float64) string {
	switch {
	case temp < -10:
		return "❄️ Очень холодно"
	case temp < 0:
		return "🥶 Холодно"
	case temp < 15:
		return "🌥️ Прохладно"
	case temp < 25:
		return "☀️ Тепло"
	default:
		return "🔥 Жарко"
	}
}

// handleUnknownCommand — обработка неизвестных команд
func (h *Handler) handleUnknownCommand(update tgbotapi.Update) {
	log.Printf("Unknown command from %s: %s", update.Message.From.UserName, update.Message.Text)
	msg := tgbotapi.NewMessage(
		update.Message.Chat.ID,
		"❌ Неизвестная команда. Доступные команды:\n/city <город>\n/weather",
	)
	msg.ReplyToMessageID = update.Message.MessageID
	h.bot.Send(msg)
}
