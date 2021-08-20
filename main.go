package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-telegram-bot-api/telegram-bot-api"
)

type bResponse = struct {
	Symbol string  `json:"symbol"`
	Price  float64 `json:"price,string"`
}

type wallet map[string]float64

var db = map[int]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI(getToken())
	if err != nil {
		log.Panic(err)
	}
	//bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}
		command := strings.Split(update.Message.Text, " ")
		userID := update.Message.From.ID

		switch strings.ToUpper(command[0]) {
		case "ADD":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы!"))
				continue
			}
			command[1] = strings.ToUpper(command[1])
			_, err := getPrice(command[1] , "")
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверная валюта!"))
				continue
			}

			money, err := strconv.ParseFloat(command[2], 64)
			fmt.Println(money)

			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}

			if _, ok := db[userID]; !ok {
				db[userID] = make(wallet)
			}

			db[userID][command[1]] += money
			fmt.Println(db)
		case "SUB":
			if len(command) != 3 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы!"))
				continue
			}
			command[1] = strings.ToUpper(command[1])
			money, err := strconv.ParseFloat(command[2], 64)
			if err != nil {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
				continue
			}

			if _, ok := db[userID]; !ok {
				db[userID] = make(wallet)
			}

			db[userID][command[1]] -= money
		case "DEL":
			if len(command) != 2 {
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неверные аргументы!"))
				continue
			}
			command[1] = strings.ToUpper(command[1])
			delete(db[userID], command[1])
		case "SHOW":
			res := ""

			var usd float64
			var rub float64
			var sum float64

			for key, value := range db[userID] {
				usd, err = getPrice(key , "USD")
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					continue
				}
				rub, err = getPrice(key, "RUB")
				if err != nil {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, err.Error()))
					continue
				}
				sum += value
				res += fmt.Sprintf("%s: $%.2f (₽%.2f)\n", key, value*usd, value*usd*rub)
			}
			res += fmt.Sprintf("Сумма: $%.2f (₽%.2f)\n", sum*usd, sum*usd*rub)
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, res))
		default:
			bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "Неизвестная команда!"))
		}

		log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
	}
}

func getPrice(symbol string, currency string) (float64, error) {
	var url string
	if currency == "" || strings.ToUpper(currency) == "USD" {
		url = fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol)
	} else {
		url = "https://api.binance.com/api/v3/ticker/price?symbol=USDTRUB"
	}
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}

	var res bResponse
	err = json.NewDecoder(resp.Body).Decode(&res)

	if err != nil {
		return 0, err
	}

	if res.Symbol == "" {
		return 0, errors.New("неверная валюта")
	}

	return res.Price, nil
}


func getToken() string {
	return "YorKey"
}
