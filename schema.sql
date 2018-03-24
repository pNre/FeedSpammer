DROP TABLE IF EXISTS "subscription";

CREATE TABLE "subscription" (
    "id" INTEGER NOT NULL PRIMARY KEY,
    "subscriber_id" VARCHAR(64) NOT NULL,
    "type_id" VARCHAR(32) NOT NULL,
    "feed_url" TEXT DEFAULT NULL,
    CONSTRAINT subscription_un UNIQUE (subscriber_id, type_id, feed_url)
);
CREATE INDEX subscription_subscriber ON subscription (subscriber_id);
CREATE INDEX subscription_feed ON subscription (feed_url);

DROP TABLE IF EXISTS "sent_item";

CREATE TABLE "sent_item" (
    "subscription_id" INTEGER NOT NULL REFERENCES subscription(id) ON DELETE CASCADE,
    "last_item_url" TEXT DEFAULT NULL,
    CONSTRAINT sent_item_un UNIQUE (subscription_id, last_item_url)
);
CREATE INDEX sent_item_subscription ON sent_item (subscription_id);
