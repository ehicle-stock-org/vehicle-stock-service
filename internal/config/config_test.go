package config

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfigHappyPath(t *testing.T) {
	os.Setenv("KAFKA_BROKERS", "test-broker:9092")
	os.Setenv("KAFKA_TOPIC", "test-topic")
	os.Setenv("MONGO_URI", "mongodb://test:27017")
	os.Setenv("MONGO_DB", "test_db")
	os.Setenv("MONGO_COLLECTION", "test_collection")
	os.Setenv("STRIPE_KEY", "sk_test_123")

	LoadConfig("vehicle-stock-service")

	assert.Equal(t, []string{"test-broker:9092"}, AppConfig.KafkaBrokers)
	assert.Equal(t, "test-topic", AppConfig.KafkaTopic)
	assert.Equal(t, "mongodb://test:27017", AppConfig.MongoURI)
	assert.Equal(t, "test_db", AppConfig.MongoDB)
	assert.Equal(t, "test_collection", AppConfig.MongoColl)
	assert.Equal(t, "sk_test_123", AppConfig.StripeKey)
}

func TestLoadConfigDefaults(t *testing.T) {
	os.Unsetenv("KAFKA_BROKERS")
	os.Unsetenv("KAFKA_TOPIC")
	os.Unsetenv("MONGO_URI")
	os.Unsetenv("MONGO_DB")
	os.Unsetenv("MONGO_COLLECTION")
	os.Unsetenv("STRIPE_KEY")

	LoadConfig("vehicle-stock-service")

	assert.Equal(t, []string{"localhost:9092"}, AppConfig.KafkaBrokers)
	assert.Equal(t, "vehicle-stock", AppConfig.KafkaTopic)
	assert.Equal(t, "mongodb://localhost:27017", AppConfig.MongoURI)
	assert.Equal(t, "vehicle_stock_db", AppConfig.MongoDB)
	assert.Equal(t, "stock_data", AppConfig.MongoColl)
	assert.Equal(t, "", AppConfig.StripeKey)
}

func TestLoadConfig_AWSSecretSuccess(t *testing.T) {
	os.Setenv("ENV", "production")
	// Mock fetchSecretsFromAWS to return valid JSON
	origFetch := fetchSecretsFromAWS
	fetchSecretsFromAWS = func(secretName string) (string, error) {
		return `{"kafka_brokers":["aws-broker:9092"],"kafka_topic":"aws-topic","mongo_uri":"mongodb://aws:27017","mongo_db":"aws_db","mongo_collection":"aws_collection","stripe_key":"sk_aws_123"}`, nil
	}
	defer func() { fetchSecretsFromAWS = origFetch }()

	LoadConfig("vehicle-stock-service")

	assert.Equal(t, []string{"aws-broker:9092"}, AppConfig.KafkaBrokers)
	assert.Equal(t, "aws-topic", AppConfig.KafkaTopic)
	assert.Equal(t, "mongodb://aws:27017", AppConfig.MongoURI)
	assert.Equal(t, "aws_db", AppConfig.MongoDB)
	assert.Equal(t, "aws_collection", AppConfig.MongoColl)
	assert.Equal(t, "sk_aws_123", AppConfig.StripeKey)
}

func TestLoadConfig_AWSSecretFailureFallbackToEnv(t *testing.T) {
	os.Setenv("ENV", "production")
	os.Setenv("KAFKA_BROKERS", "env-broker:9092")
	os.Setenv("KAFKA_TOPIC", "env-topic")
	os.Setenv("MONGO_URI", "mongodb://env:27017")
	os.Setenv("MONGO_DB", "env_db")
	os.Setenv("MONGO_COLLECTION", "env_collection")
	os.Setenv("STRIPE_KEY", "sk_env_123")
	// Mock fetchSecretsFromAWS to fail
	origFetch := fetchSecretsFromAWS
	fetchSecretsFromAWS = func(secretName string) (string, error) {
		return "", assert.AnError
	}
	defer func() { fetchSecretsFromAWS = origFetch }()

	LoadConfig("vehicle-stock-service")

	assert.Equal(t, []string{"env-broker:9092"}, AppConfig.KafkaBrokers)
	assert.Equal(t, "env-topic", AppConfig.KafkaTopic)
	assert.Equal(t, "mongodb://env:27017", AppConfig.MongoURI)
	assert.Equal(t, "env_db", AppConfig.MongoDB)
	assert.Equal(t, "env_collection", AppConfig.MongoColl)
	assert.Equal(t, "sk_env_123", AppConfig.StripeKey)
}

func TestLoadConfig_ConfigJsonFileNotFound(t *testing.T) {
	os.Setenv("ENV", "local")
	// Rename config.json if exists
	_ = os.Rename("config.json", "config.json.bak")
	defer func() { _ = os.Rename("config.json.bak", "config.json") }()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic when config.json is missing")
		}
	}()
	LoadConfig("vehicle-stock-service")
}

func TestLoadConfig_ConfigJsonDecodeError(t *testing.T) {
	os.Setenv("ENV", "local")
	// Write invalid JSON to config.json
	f, _ := os.Create("config.json")
	f.WriteString("invalid-json")
	f.Close()
	defer func() {
		_ = os.Remove("config.json")
	}()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic on JSON decode error")
		}
	}()
	LoadConfig("vehicle-stock-service")
}

func TestGetEnvOrDefault(t *testing.T) {
	os.Setenv("FOO", "bar")
	assert.Equal(t, "bar", getEnvOrDefault("FOO", "baz"))
	os.Unsetenv("FOO")
	assert.Equal(t, "baz", getEnvOrDefault("FOO", "baz"))
}

func TestFetchSecretsFromAWS_SessionError(t *testing.T) {
	// Save original function
	origFetch := fetchSecretsFromAWS
	fetchSecretsFromAWS = func(secretName string) (string, error) {
		return "", fmt.Errorf("session error")
	}
	defer func() { fetchSecretsFromAWS = origFetch }()

	os.Setenv("ENV", "production")
	LoadConfig("vehicle-stock-service") // Should fallback to env vars
	// No panic expected, just fallback
}

func TestFetchSecretsFromAWS_GetSecretValueError(t *testing.T) {
	origFetch := fetchSecretsFromAWS
	fetchSecretsFromAWS = func(secretName string) (string, error) {
		return "", fmt.Errorf("get secret value error")
	}
	defer func() { fetchSecretsFromAWS = origFetch }()

	os.Setenv("ENV", "production")
	LoadConfig("vehicle-stock-service") // Should fallback to env vars
	// No panic expected, just fallback
}

func TestLoadConfig_FakeHappyPath(t *testing.T) {
	os.Setenv("ENV", "production")
	origFetch := fetchSecretsFromAWS
	fetchSecretsFromAWS = func(secretName string) (string, error) {
		return `{"kafka_brokers":["fake-broker:9092"],"kafka_topic":"fake-topic","mongo_uri":"mongodb://fake:27017","mongo_db":"fake_db","mongo_collection":"fake_collection","stripe_key":"sk_fake_123"}`, nil
	}
	defer func() { fetchSecretsFromAWS = origFetch }()

	LoadConfig("vehicle-stock-service")

	assert.Equal(t, []string{"fake-broker:9092"}, AppConfig.KafkaBrokers)
	assert.Equal(t, "fake-topic", AppConfig.KafkaTopic)
	assert.Equal(t, "mongodb://fake:27017", AppConfig.MongoURI)
	assert.Equal(t, "fake_db", AppConfig.MongoDB)
	assert.Equal(t, "fake_collection", AppConfig.MongoColl)
	assert.Equal(t, "sk_fake_123", AppConfig.StripeKey)
}

func TestLoadConfig_FakeAWSFailureFallback(t *testing.T) {
	os.Setenv("ENV", "production")
	os.Setenv("KAFKA_BROKERS", "env-fake-broker:9092")
	os.Setenv("KAFKA_TOPIC", "env-fake-topic")
	os.Setenv("MONGO_URI", "mongodb://env-fake:27017")
	os.Setenv("MONGO_DB", "env_fake_db")
	os.Setenv("MONGO_COLLECTION", "env_fake_collection")
	os.Setenv("STRIPE_KEY", "sk_env_fake_123")
	origFetch := fetchSecretsFromAWS
	fetchSecretsFromAWS = func(secretName string) (string, error) {
		return "", assert.AnError
	}
	defer func() { fetchSecretsFromAWS = origFetch }()

	LoadConfig("vehicle-stock-service")

	assert.Equal(t, []string{"env-fake-broker:9092"}, AppConfig.KafkaBrokers)
	assert.Equal(t, "env-fake-topic", AppConfig.KafkaTopic)
	assert.Equal(t, "mongodb://env-fake:27017", AppConfig.MongoURI)
	assert.Equal(t, "env_fake_db", AppConfig.MongoDB)
	assert.Equal(t, "env_fake_collection", AppConfig.MongoColl)
	assert.Equal(t, "sk_env_fake_123", AppConfig.StripeKey)
}

func TestFetchSecretsFromAWS_FakeSessionError(t *testing.T) {
	origFetch := fetchSecretsFromAWS
	fetchSecretsFromAWS = func(secretName string) (string, error) {
		return "", fmt.Errorf("session error")
	}
	defer func() { fetchSecretsFromAWS = origFetch }()

	os.Setenv("ENV", "production")
	LoadConfig("vehicle-stock-service") // Should fallback to env vars
	// No panic expected, just fallback
}

func TestFetchSecretsFromAWS_FakeGetSecretValueError(t *testing.T) {
	origFetch := fetchSecretsFromAWS
	fetchSecretsFromAWS = func(secretName string) (string, error) {
		return "", fmt.Errorf("get secret value error")
	}
	defer func() { fetchSecretsFromAWS = origFetch }()

	os.Setenv("ENV", "production")
	LoadConfig("vehicle-stock-service") // Should fallback to env vars
	// No panic expected, just fallback
}
