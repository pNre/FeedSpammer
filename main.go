package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"feedspammer/subscription"
	"feedspammer/telegram"
)

func pushSubscriptionUpdate(update subscription.SubscriptionUpdate) {
	switch update.Subscription.TypeId {
	case "Telegram":
		message := fmt.Sprintf("%s\n%s", update.Item.Title, update.Item.Link)
		telegram.SendMessage(update.Subscription.Id, message)
	default:
		log.Printf("Unknown type %v for subscription", update.Subscription.TypeId)
		return
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Not enough args")
	}

	manager, err := subscription.NewSubscriptionManager(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	manager.RegisterUpdateHandler(pushSubscriptionUpdate)
	manager.LoadSubscriptions()

	http.HandleFunc("/tg", func(responseWriter http.ResponseWriter, request *http.Request) {
		telegram.HandleUpdate(responseWriter, request, manager)
	})

	log.Fatal(http.ListenAndServe(":8000", nil))
}
