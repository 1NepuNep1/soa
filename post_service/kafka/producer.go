package kafka

import (
	"context"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/segmentio/kafka-go"
)

var writers map[string]*kafka.Writer

func InitKafkaWriters() {
	broker := os.Getenv("KAFKA_BROKER_URL")

	waitForKafka(broker)

	topics := []string{
		"client_registered",
		"post_liked",
		"post_viewed",
		"post_commented",
	}

	for _, topic := range topics {
		createTopic(topic, broker)
	}

	writers = map[string]*kafka.Writer{
		"client_registered": newWriter([]string{broker}, "client_registered"),
		"post_liked":        newWriter([]string{broker}, "post_liked"),
		"post_viewed":       newWriter([]string{broker}, "post_viewed"),
		"post_commented":    newWriter([]string{broker}, "post_commented"),
	}
}

func newWriter(brokers []string, topic string) *kafka.Writer {
	return kafka.NewWriter(kafka.WriterConfig{
		Brokers:  brokers,
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	})
}

func SendMessage(topic, key string, value []byte) {
	writer, ok := writers[topic]
	if !ok {
		log.Printf("Kafka topic %s not found", topic)
		return
	}

	err := writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(key),
		Value: value,
		Time:  time.Now(),
	})
	if err != nil {
		log.Printf("Failed to send message to %s: %v", topic, err)
	}
}

func createTopic(topic string, brokerAddr string) {
	conn, err := kafka.Dial("tcp", brokerAddr)
	if err != nil {
		log.Printf("❌ Failed to dial Kafka broker: %v", err)
		return
	}
	defer conn.Close()

	controller, err := conn.Controller()
	if err != nil {
		log.Printf("❌ Failed to get controller: %v", err)
		return
	}
	conn.Close()

	controllerAddr := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	conn, err = kafka.Dial("tcp", controllerAddr)
	if err != nil {
		log.Printf("❌ Failed to dial controller %s: %v", controllerAddr, err)
		return
	}
	defer conn.Close()

	topicConfigs := []kafka.TopicConfig{
		{
			Topic:             topic,
			NumPartitions:     1,
			ReplicationFactor: 1,
		},
	}
	err = conn.CreateTopics(topicConfigs...)
	if err != nil {
		log.Printf("⚠️ Error creating topic %s (might already exist): %v", topic, err)
	} else {
		log.Printf("✅ Topic %s created successfully", topic)
	}
}

func waitForKafka(host string) {
	for {
		conn, err := net.DialTimeout("tcp", host, 2*time.Second)
		if err == nil {
			log.Println("✅ Kafka is reachable")
			conn.Close()
			return
		}
		log.Println("⏳ Waiting for Kafka to become reachable...")
		time.Sleep(2 * time.Second)
	}
}
