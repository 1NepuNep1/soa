package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"statisticservice/database"
	"time"

	"github.com/segmentio/kafka-go"
)

func StartConsumers() {
	topics := map[string]int8{
		"post_viewed":    1,
		"post_liked":     2,
		"post_commented": 3,
	}

	for topic, eventType := range topics {
		go consumeTopic(topic, eventType)
	}
}

type InteractionEvent struct {
	ClientID uint32    `json:"client_id"`
	EntityID uint32    `json:"entity_id"`
	ActionAt time.Time `json:"action_at"`
}

func consumeTopic(topic string, eventType int8) {
	broker := os.Getenv("KAFKA_BROKER_URL")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:     []string{broker},
		GroupID:     "stat_service",
		Topic:       topic,
		StartOffset: kafka.LastOffset,
	})

	log.Printf("👂 Started Kafka consumer for topic: %s", topic)

	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			log.Printf("⚠️ Kafka read error on %s: %v", topic, err)
			continue
		}

		var event InteractionEvent
		if err := json.Unmarshal(msg.Value, &event); err != nil {
			log.Printf("❌ Failed to parse event on %s: %v", topic, err)
			continue
		}

		date := event.ActionAt.Format("2006-01-02")
		if err := insertEvent(event.EntityID, event.ClientID, eventType, date); err != nil {
			log.Printf("❌ Failed to insert event on %s: %v", topic, err)
		}
	}
}

func insertEvent(postID, userID uint32, eventType int8, date string) error {
	query := `
        INSERT INTO post_events (date, post_id, user_id, event_type, cnt)
        SELECT ?, ?, ?, ?, sumState(toUInt64(1))
    `
	_, err := database.Conn.ExecContext(context.Background(), query, date, postID, userID, eventType)
	return err
}
