package kafka

import (
	"log"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var KafkaProducerConstructor func(conf *kafka.ConfigMap) (*kafka.Producer, error) = kafka.NewProducer

// KafkaProducer is an interface for mocking
type KafkaProducer interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
	Flush(timeoutMs int) int
	Close()
	Events() chan kafka.Event
}

// Producer wraps a Kafka producer instance
type Producer struct {
	producer KafkaProducer
	topic    string
}

// NewProducer initializes a Kafka producer
func NewProducer(brokers, topic string) (*Producer, error) {
	p, err := KafkaProducerConstructor(&kafka.ConfigMap{
		"bootstrap.servers": brokers,
	})
	if err != nil {
		return nil, err
	}

	prod := &Producer{
		producer: p,
		topic:    topic,
	}

	// Start delivery report listener
	go func() {
		for e := range p.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v", ev.TopicPartition.Error)
				} else {
					log.Printf("Message delivered to %v", ev.TopicPartition)
				}
			}
		}
	}()

	return prod, nil
}

// Publish sends a message to Kafka
func (p *Producer) Publish(key string, value []byte) {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          value,
	}

	err := p.producer.Produce(msg, nil)
	if err != nil {
		log.Printf("Kafka publish error: %v", err)
	}
}

// Close the producer
func (p *Producer) Close() {
	p.producer.Flush(1000)
	p.producer.Close()
}
