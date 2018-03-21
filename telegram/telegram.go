package telegram

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

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

	text := update.Message.Text
	id := strconv.FormatInt(update.Message.Chat.Id, 10)

	if strings.HasPrefix(text, "/") {
		//	text begins with a /, consider it a command
		command := strings.ToLower(strings.TrimPrefix(text, "/"))
		go handleCommand(strings.Split(command, " "), manager, id)
	} else if _, err := url.ParseRequestURI(text); err == nil {
		//	text is an uri -> try to subscribe
		go manager.Subscribe(text, id, "Telegram")
	}

	responseWriter.WriteHeader(http.StatusNoContent)
}

func handleCommand(args []string, manager *subscription.SubscriptionManager, chatId string) {
	if len(args) == 0 {
		return
	}

	switch args[0] {
	case "subscriptions":
		sendSubscriptions(manager, chatId)
	case "unsubscribe":
		if len(args) > 1 {
			log.Printf("Deleting %v", args[1])
			manager.Unsubscribe(args[1], chatId)
			sendSubscriptions(manager, chatId)
		}
	}
}

func sendSubscriptions(manager *subscription.SubscriptionManager, chatId string) {
	log.Print("Listing subscriptions")
	subscriptions, err := manager.Subscriptions(chatId)
	if err != nil {
		log.Print(err)
		SendMessage(chatId, "Error reading subscriptions")
		return
	}

	if len(subscriptions) == 0 {
		SendMessage(chatId, "No subscriptions")
		return
	}

	var messageBuffer bytes.Buffer
	for _, subscription := range subscriptions {
		messageBuffer.WriteString(subscription.FeedUrl)
		messageBuffer.WriteString("\n")
	}

	SendMessage(chatId, messageBuffer.String())
}
