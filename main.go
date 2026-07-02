package main

import (
	"context"
	"log"
	"os"

	"github.com/Tembo08/my-weather-bot/clients/openweather"
	"github.com/Tembo08/my-weather-bot/clients/postgres" // ← Добавляем импорт репозитория
	"github.com/Tembo08/my-weather-bot/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	// Загружаем .env
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Подключаемся к PostgreSQL через pgx
	conn, err := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err) // ← исправлено: Fatalf
	}
	defer conn.Close(context.Background())

	// Проверяем подключение
	err = conn.Ping(context.Background())
	if err != nil {
		log.Fatal("Error ping db: ", err)
	}
	log.Println("✅ Connected to PostgreSQL")

	// 🔥 СОЗДАЁМ РЕПОЗИТОРИЙ (НОВОЕ!)
	userRepo := postgres.NewRepository(conn)

	// Создаём бота
	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Fatal(err)
	}
	bot.Debug = true
	log.Printf("✅ Authorized on account %s", bot.Self.UserName)

	// Создаём клиент OpenWeather
	owClient := openweather.New(os.Getenv("OPENWEATHERAPI_KEY"))

	// 🔥 ПЕРЕДАЁМ РЕПОЗИТОРИЙ В ХЕНДЛЕР (НОВОЕ!)
	botHandler := handler.New(bot, owClient, userRepo)

	// Запускаем бота
	log.Println("🚀 Bot started")
	botHandler.Start()
}
