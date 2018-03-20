package telegram

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"feedspammer/subscription"
)

var apiToken = os.Getenv("TELEGRAM_BOT_TOKEN")

type Chat struct {
	Id   int64  `json:"id"`
	Type string `json:"type"`
}

type User struct {
	Id        int64  `json:"id"`
	isBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

type Message struct {
	Id   int64  `json:"id"`
	From User   `json:"from"`
	Date uint64 `json:"date"`
	Chat Chat   `json:"chat"`
	Text string `json:"text"`
}

type Update struct {
	Id            int64   `json:"id"`
	Message       Message `json:"message"`
	EditedMessage Message `json:"edited_message"`
}

func buildURL(endpoint string, query url.Values) url.URL {
	u := url.URL{}
	u.Scheme = "https"
	u.Host = "api.telegram.org"
	u.Path = fmt.Sprintf("/bot%v/%v", apiToken, endpoint)
	u.RawQuery = query.Encode()

	return u
}

func SendMessage(chatId string, text string) {
	query := url.Values{}
	query.Set("chat_id", chatId)
	query.Set("text", text)
	url := buildURL("sendMessage", query)
	http.Get(url.String())
}

func HandleUpdate(responseWriter http.ResponseWriter, request *http.Request, manager *subscription.SubscriptionManager) {
	var update Update
	err := json.NewDecoder(request.Body).Decode(&update)
	if err != nil {
		http.Error(responseWriter, err.Error(), http.StatusInternalServerError)
		return
	}

	if _, err := url.ParseRequestURI(update.Message.Text); err == nil {
		id := strconv.FormatInt(update.Message.Chat.Id, 10)
		go manager.Subscribe(update.Message.Text, id, "Telegram")
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}
