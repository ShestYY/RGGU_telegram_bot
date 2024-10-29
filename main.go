package main

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

func main() {
	// Загрузка переменных окружения с токеном
	err := godotenv.Load("go.env")
	if err != nil {
		fmt.Println("Ошибка загрузки файла .env")
		return
	}
	// загрузка токена
	botToken := os.Getenv("TOKEN")
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Установка вебхука
	_ = bot.SetWebhook(&telego.SetWebhookParams{
		URL: "https://468d-95-165-158-59.ngrok-free.app/bot/" + bot.Token(),
	})

	updates, _ := bot.UpdatesViaWebhook("/bot/" + bot.Token())

	go func() {
		_ = bot.StartWebhook("localhost:80")
	}()

	defer func() {
		_ = bot.StopWebhook()
	}()

	// логика сообщения
	for update := range updates {
		if update.Message != nil {

			chatID := update.Message.Chat.ID

			responseText := fmt.Sprintf(update.Message.Text)

			sentMessage, _ := bot.SendMessage(
				tu.Message(
					tu.ID(chatID),
					responseText,
				),
			)

			fmt.Printf("Sent Message: %v\n", sentMessage)
		}
	}
}
