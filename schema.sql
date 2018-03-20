DROP TABLE IF EXISTS "subscription";

CREATE TABLE "subscription" (
    "id" VARCHAR(64) NOT NULL,
    "type_id" VARCHAR(32) NOT NULL,
    "feed_url" TEXT DEFAULT NULL,
    "last_item_url" TEXT DEFAULT NULL,
    CONSTRAINT pk PRIMARY KEY (id, type_id, feed_url)
);
