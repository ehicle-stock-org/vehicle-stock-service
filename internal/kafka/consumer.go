package kafka

import (
	"encoding/json"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/yourusername/vehicle-stock-service/internal/mongo"
)

var KafkaConsumerConstructor func(conf *kafka.ConfigMap) (*kafka.Consumer, error) = kafka.NewConsumer

// KafkaConsumer is an interface for mocking
type KafkaConsumer interface {
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
	Close() error
}

// Consumer wraps a Kafka consumer
type Consumer struct {
	consumer KafkaConsumer
	topic    string
}

// NewConsumer initializes a Kafka consumer
func NewConsumer(brokers, groupID, topic string) (*Consumer, error) {
	c, err := KafkaConsumerConstructor(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          groupID,
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return nil, err
	}

	if err := c.SubscribeTopics([]string{topic}, nil); err != nil {
		return nil, err
	}

	return &Consumer{consumer: c, topic: topic}, nil
}

// ConsumeLoop continuously reads messages from Kafka and stores in MongoDB
func (c *Consumer) ConsumeLoop(stopChan ...chan struct{}) {
	for {
		select {
		case <-getStopChan(stopChan):
			return
		default:
			msg, err := c.consumer.ReadMessage(-1)
			if err != nil {
				log.Printf("Consumer error: %v", err)
				continue
			}

			log.Printf("Message received: %s", string(msg.Value))

			var stockData map[string]interface{}
			if err := json.Unmarshal(msg.Value, &stockData); err == nil {
				if mongo.Client != nil {
					err := mongo.InsertDataFunc("vehicle_stock_db", "stock_data", stockData)
					if err != nil {
						log.Println("MongoDB insert failed:", err)
					}
				}
			}
		}
	}
}

func getStopChan(stopChan []chan struct{}) chan struct{} {
	if len(stopChan) > 0 {
		return stopChan[0]
	}
	return make(chan struct{}) // never closed, so loop continues in prod
}

// Close the consumer
func (c *Consumer) Close() {
	_ = c.consumer.Close()
}
