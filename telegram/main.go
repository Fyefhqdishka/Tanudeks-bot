package main

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type Rates struct {
	AED float64 `json:"AED"`
	GBP float64 `json:"GBP"`
	EUR float64 `json:"EUR"`
	RUB float64 `json:"RUB"`
	KZT float64 `json:"KZT"`
}

type Currency struct {
	Disclaimer string `json:"disclaimer"`
	License    string `json:"license"`
	Timestamp  int64  `json:"timestamp"`
	Base       string `json:"base"`
	Rates      Rates  `json:"rates"`
}

func main() {
	err := godotenv.Load("index.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	api_key := os.Getenv("TELEGRAM_BOT_API_TOKEN")

	bot, err := tgbotapi.NewBotAPI(api_key)
	if err != nil {
		log.Fatal("Error creating bot client")
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			var msg tgbotapi.MessageConfig

			switch update.Message.Command() {
			case "start":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Привет"+
					"\nЯ помогу отследить актуальные курсы валют.")

			case "help":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Список команд:"+
					"\n/start"+
					"\n/help"+
					"\n/news"+
					"\nДля моей любимой женщины:"+
					"\n/formywife")

			case "news":
				url := os.Getenv("URL_ID")
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					log.Printf("Ошибка при получении запроса")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка")
				}
				req.Header.Set("accept", "application/json")

				r, err := http.DefaultClient.Do(req)
				if err != nil {
					log.Printf("Ошибка при получении данных")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка")
					bot.Send(msg)
					continue
				}
				defer r.Body.Close()

				if r.StatusCode != http.StatusOK {
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Ошибка при запросе данных.")
					bot.Send(msg)
					continue
				}

				Bodybytes, err := io.ReadAll(r.Body)
				if err != nil {
					log.Printf("Ошибка в чтении тела запроса")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Произошла ошибка")
					bot.Send(msg)
					continue
				}

				currency := Currency{}
				err = json.Unmarshal(Bodybytes, &currency)
				if err != nil {
					log.Printf("Ошибка в анмаршеле")
					msg = tgbotapi.NewMessage(update.Message.Chat.ID, "ошибка при обработке данных")
					bot.Send(msg)
					continue
				}

				log.Println(string(Bodybytes))

				time := time.Unix(currency.Timestamp, 0)
				formattedTime := time.Format("02-01-2006 15:04")

				msg = tgbotapi.NewMessage(update.Message.Chat.ID,
					fmt.Sprintf("Курсы валют (Базовая валюта: %s):\nAED: %.2f\nGBP: %.2f\nEUR: %.2f\nRUB: %.2f\nKZT: %.2f\nВремя обновления: \n%s",
						currency.Base, currency.Rates.AED, currency.Rates.GBP, currency.Rates.EUR, currency.Rates.RUB, currency.Rates.KZT, formattedTime))

			case "formywife":
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Люблю Раниду"+
					"\nс Ануар")

			default:
				msg = tgbotapi.NewMessage(update.Message.Chat.ID, "Я не знаю такой команды"+
					"\nВведите: /help")

			}
			if _, err := bot.Send(msg); err != nil {
				log.Printf("ошибка при получении сообщения: %v", err)
			}
		}
	}
}
