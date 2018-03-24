package subscription

import (
	"log"
	"strings"
	"strconv"
	"time"

	"github.com/gilliek/go-opml/opml"
	"github.com/mmcdole/gofeed"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

type Subscription struct {
	Id 			 int64
	SubscriberId string
	TypeId       string
	FeedUrl      string
}

type SentItem struct {
	SubscriptionId int64
	LastItemUrl    string
}

type SubscriptionUpdate struct {
	Subscription Subscription
	Item         *gofeed.Item
}

type SubscriptionManager struct {
	Engine        *xorm.Engine
	UpdateHandler func(SubscriptionUpdate)
}

func NewSubscriptionManager(dbPath string) (*SubscriptionManager, error) {
	engine, err := xorm.NewEngine("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	//	enable foreign keys
	engine.Query("PRAGMA foreign_keys = ON")

	manager := SubscriptionManager{
		engine,
		func(u SubscriptionUpdate) {},
	}

	return &manager, nil
}

func (manager *SubscriptionManager) checkFeed(feed string) {
	feedParser := gofeed.NewParser()
	ticker := time.NewTicker(time.Millisecond * 60000)
	for ; true; <-ticker.C {
		log.Printf("Checking %v", feed)

		result, err := feedParser.ParseURL(feed)
		if err != nil || len(result.Items) == 0 {
			log.Print(err)
			continue
		}

		latestItem := result.Items[0]

		query := `
		SELECT subscription.*
		FROM subscription
		WHERE subscription.feed_url = ?
			AND (SELECT COUNT(1)
			     FROM sent_item
			     WHERE subscription_id = subscription.id AND last_item_url = ?) = 0
		`
		subscriptions, err := manager.Engine.Query(query, feed, latestItem.Link)
		if err != nil {
			log.Printf("Checking %v: %v", feed, err)
			continue
		}

		for _, subscriptionData := range subscriptions {
			id, _ := strconv.ParseInt(string(subscriptionData["id"]), 10, 64)
			subscription := Subscription{
				id,
				string(subscriptionData["subscriber_id"]),
				string(subscriptionData["type_id"]),
				string(subscriptionData["feed_url"])}
			manager.UpdateHandler(SubscriptionUpdate{subscription, latestItem})
			manager.Engine.Insert(&SentItem{id, latestItem.Link})
		}
	}
}

func (manager *SubscriptionManager) addSubscription(feed string, subscriberId string, subscriptionType string) {
	feedParser := gofeed.NewParser()
	if _, err := feedParser.ParseURL(feed); err != nil {
		log.Print(err)
		return
	}

	subscription := Subscription{SubscriberId: subscriberId, TypeId: subscriptionType, FeedUrl: feed}

	log.Printf("Adding subscription to %v", feed)

	if _, err := manager.Engine.Insert(&subscription); err != nil {
		log.Printf("Couldn't add subscription to %v: %v", feed, err)
		return
	}

	go manager.checkFeed(feed)
}

func (manager *SubscriptionManager) Subscribe(feed string, subscriberId string, subscriptionType string) {
	if strings.HasSuffix(strings.ToLower(feed), "opml") {
		doc, err := opml.NewOPMLFromURL(feed)
		if err != nil {
			log.Print(err)
			return
		}

		for _, outline := range doc.Outlines() {
			manager.addSubscription(outline.XMLURL, subscriberId, subscriptionType)
		}
	} else {
		manager.addSubscription(feed, subscriberId, subscriptionType)
	}
}

func (manager *SubscriptionManager) LoadSubscriptions() error {
	var feeds []string
	if err := manager.Engine.Table("subscription").Distinct("feed_url").Find(&feeds); err != nil {
		log.Print(err)
		return err
	}

	for _, feed := range feeds {
		log.Printf("Resuming %v", feed)
		go manager.checkFeed(feed)
	}

	return nil
}

func (manager *SubscriptionManager) Subscriptions(subscriberId string) ([]Subscription, error) {
	var subscriptions []Subscription
	err := manager.Engine.Table("subscription").Where("subscriber_id = ?", subscriberId).Find(&subscriptions)
	return subscriptions, err
}

func (manager *SubscriptionManager) Unsubscribe(feed string, subscriberId string) error {
	var subscriptions []Subscription
	err := manager.Engine.Table("subscription").Where("subscriber_id = ? AND feed_url = ?", subscriberId, feed).Find(&subscriptions)
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		manager.Engine.Delete(&subscription)
	}

	return nil
}
