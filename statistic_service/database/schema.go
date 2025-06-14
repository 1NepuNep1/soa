package database

import (
	"context"
	"log"
)

func EnsureSchema() {
	query := `
	CREATE TABLE IF NOT EXISTS post_events (
		date Date,
		post_id UInt32,
		user_id UInt32,
		event_type Enum8('view'=1, 'like'=2, 'comment'=3),
		cnt AggregateFunction(sum, UInt64)
	) ENGINE = AggregatingMergeTree()
	PARTITION BY toYYYYMM(date)
	ORDER BY (post_id, user_id, event_type, date)
	`
	if _, err := Conn.ExecContext(context.Background(), query); err != nil {
		log.Fatalf("failed to create post_events table: %v", err)
	}
	log.Println("✔ Ensured schema: post_events")
}
