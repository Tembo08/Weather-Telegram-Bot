package main

import (
	"log"
	"os"

	"github.com/Tembo08/my-weather-bot/clients/openweather"
	"github.com/Tembo08/my-weather-bot/handler"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	owClient := openweather.New(os.Getenv("OPENWEATHERAPI_KEY"))

	botHandler := handler.New(bot, owClient)

	botHandler.Start()

}
