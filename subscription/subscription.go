package subscription

import (
	"log"
	"strings"
	"time"

	"github.com/gilliek/go-opml/opml"
	"github.com/mmcdole/gofeed"

	"github.com/go-xorm/xorm"
	_ "github.com/mattn/go-sqlite3"
)

type Subscription struct {
	Id          string
	TypeId      string
	FeedUrl     string
	LastItemUrl string
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

	manager := SubscriptionManager{
		engine,
		func(u SubscriptionUpdate) {},
	}

	return &manager, nil
}

func (manager *SubscriptionManager) CheckFeed(feed string) {
	feedParser := gofeed.NewParser()
	ticker := time.NewTicker(time.Millisecond * 60000)
	for ; true; <-ticker.C {
		log.Printf("Checking %v", feed)

		result, err := feedParser.ParseURL(feed)
		if err != nil || len(result.Items) == 0 {
			log.Print(err)
			continue
		}

		var subscriptions []Subscription
		if err = manager.Engine.Table("subscription").Where("feed_url = ?", feed).Find(&subscriptions); err != nil {
			log.Printf("Checking %v: %v", feed, err)
			continue
		}

		for _, subscription := range subscriptions {
			if result.Items[0].Link == subscription.LastItemUrl {
				continue
			}
			subscription.LastItemUrl = result.Items[0].Link
			search := Subscription{Id: subscription.Id, TypeId: subscription.TypeId, FeedUrl: subscription.FeedUrl}
			manager.Engine.Update(&subscription, &search)
			manager.UpdateHandler(SubscriptionUpdate{subscription, result.Items[0]})
		}
	}
}

func (manager *SubscriptionManager) addSubscription(feed string, subscriberId string, subscriptionType string) {
	feedParser := gofeed.NewParser()
	if _, err := feedParser.ParseURL(feed); err != nil {
		log.Print(err)
		return
	}

	subscription := Subscription{
		Id:      subscriberId,
		TypeId:  subscriptionType,
		FeedUrl: feed,
	}

	log.Printf("Adding subscription to %v", feed)

	if _, err := manager.Engine.Insert(&subscription); err != nil {
		log.Printf("Couldn't add subscription to %v: %v", feed, err)
		return
	}

	go manager.CheckFeed(feed)
}

func (manager *SubscriptionManager) resumeSubscription(feed string) {
	log.Printf("Resuming %v", feed)
	go manager.CheckFeed(feed)
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
		manager.resumeSubscription(feed)
	}

	return nil
}

func (manager *SubscriptionManager) Subscriptions(subscriberId string) ([]Subscription, error) {
	var subscriptions []Subscription
	err := manager.Engine.Table("subscription").Where("id = ?", subscriberId).Find(&subscriptions)
	return subscriptions, err
}

func (manager *SubscriptionManager) Unsubscribe(feed string, subscriberId string) error {
	var subscriptions []Subscription
	err := manager.Engine.Table("subscription").Where("id = ? AND feed_url = ?", subscriberId, feed).Find(&subscriptions)
	if err != nil {
		return err
	}

	for _, subscription := range subscriptions {
		manager.Engine.Delete(&subscription)
	}

	return nil
}
