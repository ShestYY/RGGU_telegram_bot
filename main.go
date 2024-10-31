package main

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// переменная для отслеживания активности бота
var isActive bool

func main() {
	err := godotenv.Load("go.env")
	if err != nil {
		fmt.Println("Ошибка загрузки файла .env")
		return
	}
	botToken := os.Getenv("TOKEN")
	bot, err := telego.NewBot(botToken, telego.WithDefaultDebugLogger())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	_ = bot.SetWebhook(&telego.SetWebhookParams{
		URL: "https://ec1e-95-165-158-59.ngrok-free.app/bot/" + bot.Token(),
	})

	updates, _ := bot.UpdatesViaWebhook("/bot/" + bot.Token())

	go func() {
		_ = bot.StartWebhook("localhost:80")
	}()

	defer func() {
		_ = bot.StopWebhook()
	}()

	for update := range updates {
		if update.Message != nil {
			chatID := update.Message.Chat.ID
			text := update.Message.Text

			if text == "/start" {
				isActive = true // Активируем бота
				keyboard := tu.Keyboard(
					tu.KeyboardRow(
						tu.KeyboardButton("Показать расписание"),
					),
				)

				_, _ = bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						"Добро пожаловать! Выберите опцию:",
					).WithReplyMarkup(keyboard),
				)
				continue
			}

			// Проверяем, активен ли бот
			if !isActive {
				continue
			}

			if text == "Показать расписание" {
				keyboard2 := tu.Keyboard(
					tu.KeyboardRow(
						tu.KeyboardButton("сегодня-завтра"),
					),
					tu.KeyboardRow(
						tu.KeyboardButton("неделя"),
					),
					tu.KeyboardRow(
						tu.KeyboardButton("семестр"),
					),
				)

				_, _ = bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						"На какой период нужно показать расписание?",
					).WithReplyMarkup(keyboard2.WithIsPersistent()),
				)
				continue
			}

			if text == "сегодня-завтра" {
				isActive = false
				parse()
				responseText := "da"

				_, _ = bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						responseText,
					),
				)

				continue
			}

			if text == "неделя" {
				isActive = false
				parse()
				responseText := "da"

				_, _ = bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						responseText,
					),
				)

				continue
			}

			if text == "семестр" {
				isActive = false
				parse()
				responseText := "da"

				_, _ = bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						responseText,
					),
				)

				continue
			}
		}
	}
}

func parse() {

	getURL := "https://functions.yandexcloud.net/d4e3jfkud94j5ovcurg2?method=/proxy/timetables&group_id=%D0%98%D0%9F_%D0%9F%D0%9F%D0%9E_%D0%9F%D0%9F%D0%9A%20(%D0%93%D1%80%D1%83%D0%BF%D0%BF%D0%B0:%201)%20[%D0%94:4]"

	resp, err := http.Get(getURL)
	if err != nil {
		fmt.Println("Ошибка при отправке запроса:", err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Ошибка при чтении ответа:", err)
		return
	}

	fmt.Println("Статус код:", resp.Status)
	fmt.Println("Ответ:", string(body))
}
