package main

import (
	"log"
	"net/http"
	"os"

	"feedspammer/subscription"
	"feedspammer/telegram"
)

func pushSubscriptionUpdate(update subscription.SubscriptionUpdate) {
	switch update.Subscription.TypeId {
	case "Telegram":
		telegram.SendMessage(update.Subscription.Id, update.Item.Link)
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
