package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/mymmrac/telego"
	tu "github.com/mymmrac/telego/telegoutil"
)

// Глобальная переменная для хранения расписания
var raspisanie []Raspisanie

type Raspisanie struct {
	WeekDay   string `json:"week_day"`
	Nomerpari int    `json:"number"`
	Date      string `json:"date"`
	Kabinet   string `json:"cabinet_name"`
	Para      string `json:"name"`
	Type      string `json:"type"`
	Prepod    string `json:"teacher_name"`
	GroupID   string `json:"group_id"`
	ID        string `json:"id"`
}

var weekDays = map[int]string{
	0: "Воскресенье",
	1: "Понедельник",
	2: "Вторник",
	3: "Среда",
	4: "Четверг",
	5: "Пятница",
	6: "Суббота",
}

type ApiResponse struct {
	Status string       `json:"status"`
	Data   []Raspisanie `json:"data"`
}

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

			if text == "Показать расписание" {
				keyboard2 := tu.Keyboard(
					tu.KeyboardRow(
						tu.KeyboardButton("сегодня-завтра"),
					),
					tu.KeyboardRow(
						tu.KeyboardButton("текущая неделя"),
					),
					tu.KeyboardRow(
						tu.KeyboardButton("следующая неделя"),
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

				parse() // Запрашиваем расписание
				responseText := getTodayTomorrowSchedule()
				_, _ = bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						responseText,
					),
				)

				continue
			}

			if text == "текущая неделя" {

				parse() // Запрашиваем расписание
				responseText := getThisWeekSchedule()
				_, _ = bot.SendMessage(
					tu.Message(
						tu.ID(chatID),
						responseText,
					),
				)

				continue
			}

			if text == "следующая неделя" {

				parse() // Запрашиваем расписание
				responseText := getNextWeekSchedule()
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
	getURL := "https://functions.yandexcloud.net/d4e3jfkud94j5ovcurg2?method=/proxy/timetables&group_id=%D0%98%D0%9F_%D0%9F%D0%9F%D0%9E_%D0%9F%D0%9F%D0%9A%20(%D0%93%D1%80%D1%83%D0%BF%D0%BF%D0%B0:%201)%20[%D0%94:3]"

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

	var apiResponse ApiResponse
	err = json.Unmarshal(body, &apiResponse)
	if err != nil {
		fmt.Println("Ошибка при декодировании JSON:", err)
		return
	}

	raspisanie = apiResponse.Data
}

func getTodayTomorrowSchedule() string {
	var responseText string
	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1)

	responseText += "\n" // Добавляем новую строку в начале ответа

	for _, item := range raspisanie {
		// Преобразуем строку даты в time.Time
		date, err := time.Parse("2006-01-02T15:04:05.000000Z", item.Date)
		if err != nil {
			fmt.Println("Ошибка при парсинге даты:", err)
			continue
		}

		// Проверяем, попадает ли дата в сегодня или завтра
		if (date.Year() == today.Year() && date.YearDay() == today.YearDay()) ||
			(date.Year() == tomorrow.Year() && date.YearDay() == tomorrow.YearDay()) {

			// Получаем день недели и форматируем дату
			weekdayName := weekDays[int(date.Weekday())]
			formattedDate := date.Format("2006-01-02")

			responseText += fmt.Sprintf("\n%s\n %d Пара \nДата: %s\nКабинет: %s\nПара: %s\nТип: %s\nПреподаватель: %s\n\n",
				weekdayName, item.Nomerpari, formattedDate, item.Kabinet, item.Para, item.Type, item.Prepod)
		}
	}

	if responseText == "\n" {
		responseText = "Нет расписания на сегодня или завтра."
	}
	return responseText
}

func getThisWeekSchedule() string {
	var responseText string

	// Определяем начало и конец текущей недели
	today := time.Now()
	weekday := int(today.Weekday())
	startOfWeek := today.AddDate(0, 0, -weekday+1) // Понедельник текущей недели
	endOfWeek := startOfWeek.AddDate(0, 0, 6)      // Воскресенье текущей недели

	responseText += "\n" // Начинаем с новой строки

	for _, item := range raspisanie {
		// Преобразуем строку даты в time.Time
		date, err := time.Parse("2006-01-02T15:04:05.000000Z", item.Date)
		if err != nil {
			fmt.Println("Ошибка при парсинге даты:", err)
			continue
		}

		// Проверяем, попадает ли дата в текущую неделю
		if date.After(startOfWeek.Add(-time.Second)) && date.Before(endOfWeek.Add(24*time.Hour)) {
			weekdayName := weekDays[int(date.Weekday())]
			formattedDate := date.Format("2006-01-02")

			responseText += fmt.Sprintf("День: %s\n %d Пара \nДата: %s\nКабинет: %s\nПара: %s\nТип: %s\nПреподаватель: %s\n\n",
				weekdayName, item.Nomerpari, formattedDate, item.Kabinet, item.Para, item.Type, item.Prepod)
		}
	}

	if responseText == "\n" {
		responseText = "Нет расписания на текущую неделю."
	}
	return responseText
}

func getNextWeekSchedule() string {
	var responseText string

	// Определяем начало и конец следующей недели
	today := time.Now()
	weekday := int(today.Weekday())
	startOfNextWeek := today.AddDate(0, 0, -weekday+8) // Понедельник следующей недели
	endOfNextWeek := startOfNextWeek.AddDate(0, 0, 6)  // Воскресенье следующей недели

	responseText += "\n" // Начинаем с новой строки

	for _, item := range raspisanie {
		// Преобразуем строку даты в time.Time
		date, err := time.Parse("2006-01-02T15:04:05.000000Z", item.Date)
		if err != nil {
			fmt.Println("Ошибка при парсинге даты:", err)
			continue
		}

		// Проверяем, попадает ли дата в следующую неделю
		if date.After(startOfNextWeek.Add(-time.Second)) && date.Before(endOfNextWeek.Add(24*time.Hour)) {
			weekdayName := weekDays[int(date.Weekday())]
			formattedDate := date.Format("2006-01-02")

			responseText += fmt.Sprintf("День: %s\n%d Пара \nДата: %s\nКабинет: %s\nПара: %s\nТип: %s\nПреподаватель: %s\n\n",
				weekdayName, item.Nomerpari, formattedDate, item.Kabinet, item.Para, item.Type, item.Prepod)
		}
	}

	if responseText == "\n" {
		responseText = "Нет расписания на следующую неделю."
	}
	return responseText
}
