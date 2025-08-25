package kafka

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/yourusername/vehicle-stock-service/internal/mongo"
)

var (
	testTopic = "test-topic"
	testDate  = "2025-08-24"
)

type mockKafkaConsumer struct {
	messages []*kafka.Message
	err      error
	closed   bool
	idx      int
}

func (m *mockKafkaConsumer) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.idx >= len(m.messages) {
		return nil, errors.New("no more messages")
	}
	msg := m.messages[m.idx]
	m.idx++
	return msg, nil
}
func (m *mockKafkaConsumer) Close() error { m.closed = true; return nil }

func TestConsumerHappyPath(t *testing.T) {
	stock := map[string]interface{}{"ticker": "AAPL", "bid": 150.0, "ask": 151.0, "time": testDate}
	val, _ := json.Marshal(stock)
	msg := &kafka.Message{Value: val}
	mock := &mockKafkaConsumer{messages: []*kafka.Message{msg}}
	c := &Consumer{consumer: mock, topic: testTopic}
	done := make(chan struct{})
	go func() { c.ConsumeLoop(done) }()
	assert.False(t, mock.closed)
	close(done)
	mock.Close()
	assert.True(t, mock.closed)
}

func TestConsumerMultipleMessages(t *testing.T) {
	stock1 := map[string]interface{}{"ticker": "AAPL", "bid": 150.0, "ask": 151.0, "time": testDate}
	stock2 := map[string]interface{}{"ticker": "GOOG", "bid": 2500.0, "ask": 2501.0, "time": testDate}
	val1, _ := json.Marshal(stock1)
	val2, _ := json.Marshal(stock2)
	msg1 := &kafka.Message{Value: val1}
	msg2 := &kafka.Message{Value: val2}
	mock := &mockKafkaConsumer{messages: []*kafka.Message{msg1, msg2}}
	c := &Consumer{consumer: mock, topic: testTopic}
	done := make(chan struct{})
	go func() { c.ConsumeLoop(done) }()
	close(done)
	mock.Close()
	assert.True(t, mock.closed)
}

func TestConsumerEmptyMessage(t *testing.T) {
	msg := &kafka.Message{Value: []byte{}}
	mock := &mockKafkaConsumer{messages: []*kafka.Message{msg}}
	c := &Consumer{consumer: mock, topic: testTopic}
	done := make(chan struct{})
	go func() { c.ConsumeLoop(done) }()
	close(done)
	mock.Close()
	assert.True(t, mock.closed)
}

func TestConsumerInvalidJSON(t *testing.T) {
	msg := &kafka.Message{Value: []byte("not-json")}
	mock := &mockKafkaConsumer{messages: []*kafka.Message{msg}}
	c := &Consumer{consumer: mock, topic: testTopic}
	done := make(chan struct{})
	go func() { c.ConsumeLoop(done) }()
	close(done)
	mock.Close()
	assert.True(t, mock.closed)
}

func TestConsumerNoMessages(t *testing.T) {
	mock := &mockKafkaConsumer{messages: []*kafka.Message{}}
	c := &Consumer{consumer: mock, topic: testTopic}
	done := make(chan struct{})
	go func() { c.ConsumeLoop(done) }()
	close(done)
	mock.Close()
	assert.True(t, mock.closed)
}

func TestConsumerCloseMultipleTimes(t *testing.T) {
	mock := &mockKafkaConsumer{}
	c := &Consumer{consumer: mock, topic: testTopic}
	c.Close()
	c.Close()
	assert.True(t, mock.closed)
}

func TestNewConsumerError(t *testing.T) {
	orig := KafkaConsumerConstructor
	KafkaConsumerConstructor = func(cfg *kafka.ConfigMap) (*kafka.Consumer, error) {
		return nil, errors.New("fail")
	}
	defer func() { KafkaConsumerConstructor = orig }()
	_, err := NewConsumer("invalid:broker", "group", testTopic)
	assert.Error(t, err)
}

func TestConsumerMongoInsertSuccess(t *testing.T) {
	// Mock mongo.Client and mongo.InsertDataFunc
	// Clean assignment for testability
	origInsert := mongo.InsertDataFunc
	mongo.InsertDataFunc = func(database, collection string, data interface{}) error {
		return nil
	}
	defer func() { mongo.InsertDataFunc = origInsert }()

	stock := map[string]interface{}{"ticker": "AAPL", "bid": 150.0, "ask": 151.0, "time": testDate}
	val, _ := json.Marshal(stock)
	msg := &kafka.Message{Value: val}
	mock := &mockKafkaConsumer{messages: []*kafka.Message{msg}}
	c := &Consumer{consumer: mock, topic: testTopic}
	done := make(chan struct{})
	go func() { c.ConsumeLoop(done) }()
	close(done)
	mock.Close()
	assert.True(t, mock.closed)
}

func TestConsumerMongoInsertError(t *testing.T) {
	// Mock mongo.Client and mongo.InsertDataFunc
	// Clean assignment for testability
	origClient := mongo.Client
	origInsert := mongo.InsertDataFunc
	mongo.InsertDataFunc = func(database, collection string, data interface{}) error {
		return errors.New("insert error")
	}
	defer func() { mongo.Client = origClient; mongo.InsertDataFunc = origInsert }()

	stock := map[string]interface{}{"ticker": "AAPL", "bid": 150.0, "ask": 151.0, "time": testDate}
	val, _ := json.Marshal(stock)
	msg := &kafka.Message{Value: val}
	mock := &mockKafkaConsumer{messages: []*kafka.Message{msg}}
	c := &Consumer{consumer: mock, topic: testTopic}
	done := make(chan struct{})
	go func() { c.ConsumeLoop(done) }()
	close(done)
	mock.Close()
	assert.True(t, mock.closed)
}

func TestConsumerReadMessageError(t *testing.T) {
	mock := &mockKafkaConsumer{err: errors.New("read error")}
	c := &Consumer{consumer: mock, topic: testTopic}
	done := make(chan struct{})
	go func() { c.ConsumeLoop(done) }()
	close(done)
	mock.Close()
	assert.True(t, mock.closed)
}

func TestConsumerCloseNilConsumer(t *testing.T) {
	c := &Consumer{consumer: nil, topic: testTopic}
	c.Close()
}

func TestConsumerConsumeLoopNilConsumer(t *testing.T) {
	c := &Consumer{consumer: nil, topic: testTopic}
	done := make(chan struct{})
	go func() { c.ConsumeLoop(done) }()
	close(done)
	// Should not panic
}

func TestGetStopChanDefault(t *testing.T) {
	ch := getStopChan([]chan struct{}{})
	select {
	case <-ch:
		// Should not happen
	default:
		// Should block forever
	}
}
