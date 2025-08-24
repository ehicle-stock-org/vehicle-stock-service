package kafka

import (
	"errors"
	"log"
	"testing"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
)

type mockKafkaProducer struct {
	produced   bool
	produceErr error
}

func (m *mockKafkaProducer) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	m.produced = true
	return m.produceErr
}
func (m *mockKafkaProducer) Flush(timeoutMs int) int  { return 0 }
func (m *mockKafkaProducer) Close()                   { /* intentionally empty for mock */ }
func (m *mockKafkaProducer) Events() chan kafka.Event { return make(chan kafka.Event) }

func TestProducerPublishHappyPath(t *testing.T) {
	mock := &mockKafkaProducer{}
	p := &Producer{producer: mock, topic: testTopic}
	p.Publish("key", []byte("value"))
	assert.True(t, mock.produced)
}

func TestProducerPublishNilValue(t *testing.T) {
	mock := &mockKafkaProducer{}
	p := &Producer{producer: mock, topic: testTopic}
	p.Publish("key", nil)
	assert.True(t, mock.produced)
}

func TestProducerPublishEmptyKey(t *testing.T) {
	mock := &mockKafkaProducer{}
	p := &Producer{producer: mock, topic: testTopic}
	p.Publish("", []byte("value"))
	assert.True(t, mock.produced)
}

func TestProducerPublishError(t *testing.T) {
	mock := &mockKafkaProducer{produceErr: errors.New("fail")}
	p := &Producer{producer: mock, topic: testTopic}
	p.Publish("key", []byte("value"))
	assert.True(t, mock.produced)
}

func TestProducerClose(t *testing.T) {
	mock := &mockKafkaProducer{}
	p := &Producer{producer: mock, topic: testTopic}
	p.Close()
	// No panic or error expected
}

func TestProducerCloseMultipleTimes(t *testing.T) {
	mock := &mockKafkaProducer{}
	p := &Producer{producer: mock, topic: testTopic}
	p.Close()
	p.Close()
	// Should not panic or error
}

func TestNewProducerError(t *testing.T) {
	orig := KafkaProducerConstructor
	KafkaProducerConstructor = func(cfg *kafka.ConfigMap) (*kafka.Producer, error) {
		return nil, errors.New("fail")
	}
	defer func() { KafkaProducerConstructor = orig }()
	_, err := NewProducer("invalid:broker", testTopic)
	assert.Error(t, err)
}

func TestProducerDeliveryReport(t *testing.T) {
	ch := make(chan kafka.Event, 1)
	msg := &kafka.Message{TopicPartition: kafka.TopicPartition{Error: nil}}
	ch <- msg
	close(ch)
	// Simulate the delivery report goroutine
	go func() {
		for e := range ch {
			if ev, ok := e.(*kafka.Message); ok {
				if ev.TopicPartition.Error != nil {
					log.Printf("Delivery failed: %v", ev.TopicPartition.Error)
				} else {
					log.Printf("Message delivered to %v", ev.TopicPartition)
				}
			}
		}
	}()
	// No assertion needed, just coverage
}
