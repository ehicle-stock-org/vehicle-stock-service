package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig_HappyPath(t *testing.T) {
	os.Setenv("KAFKA_BROKERS", "test-broker:9092")
	os.Setenv("KAFKA_TOPIC", "test-topic")
	os.Setenv("MONGO_URI", "mongodb://test:27017")
	os.Setenv("MONGO_DB", "test_db")
	os.Setenv("MONGO_COLLECTION", "test_collection")
	os.Setenv("STRIPE_KEY", "sk_test_123")

	LoadConfig()

	assert.Equal(t, []string{"test-broker:9092"}, AppConfig.KafkaBrokers)
	assert.Equal(t, "test-topic", AppConfig.KafkaTopic)
	assert.Equal(t, "mongodb://test:27017", AppConfig.MongoURI)
	assert.Equal(t, "test_db", AppConfig.MongoDB)
	assert.Equal(t, "test_collection", AppConfig.MongoColl)
	assert.Equal(t, "sk_test_123", AppConfig.StripeKey)
}

func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("KAFKA_BROKERS")
	os.Unsetenv("KAFKA_TOPIC")
	os.Unsetenv("MONGO_URI")
	os.Unsetenv("MONGO_DB")
	os.Unsetenv("MONGO_COLLECTION")
	os.Unsetenv("STRIPE_KEY")

	LoadConfig()

	assert.Equal(t, []string{"localhost:9092"}, AppConfig.KafkaBrokers)
	assert.Equal(t, "vehicle-stock", AppConfig.KafkaTopic)
	assert.Equal(t, "mongodb://localhost:27017", AppConfig.MongoURI)
	assert.Equal(t, "vehicle_stock_db", AppConfig.MongoDB)
	assert.Equal(t, "stock_data", AppConfig.MongoColl)
	assert.Equal(t, "", AppConfig.StripeKey)
}
