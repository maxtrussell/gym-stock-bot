package telegram

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api"
)

func ListenAndServe(api_token string) {
	bot, err := tgbot.NewBotAPI(api_token)
	if err != nil {
		log.Fatal(err)
	}

	u := tgbot.NewUpdate(0)
	u.Timeout = 60

	updates, _ := bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		if !update.Message.IsCommand() {
			continue
		}

		msg := tgbot.NewMessage(update.Message.Chat.ID, "")
		switch update.Message.Command() {
		case "hi":
			msg.Text = "Howdy world!"
		case "latest":
			msg.Text = readLatestLog()
		default:
			msg.Text = "I don't know that command"
		}

		if _, err := bot.Send(msg); err != nil {
			log.Fatal(err)
		}
	}
}

func SendMessage(api_token, chat_id, msg string) {
	endpoint := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", api_token)
	data := url.Values{
		"chat_id": {chat_id},
		"text":    {msg},
	}
	_, err := http.PostForm(endpoint, data)
	if err != nil {
		log.Fatal(err)
	}
}

func readLatestLog() string {
	content, err := ioutil.ReadFile("latest_log.txt")
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}
