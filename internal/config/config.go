package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
)

// Config holds all app-wide configuration
type Config struct {
	KafkaBrokers []string `json:"kafka_brokers"`
	KafkaTopic   string   `json:"kafka_topic"`
	MongoURI     string   `json:"mongo_uri"`
	MongoDB      string   `json:"mongo_db"`
	MongoColl    string   `json:"mongo_collection"`
	StripeKey    string   `json:"stripe_key"`
}

// AppConfig is the exported global configuration
var AppConfig Config

// LoadConfig loads configuration from config.json (local) or env/AWS Secrets Manager (non-local)
func LoadConfig(serviceName string) {
	env := os.Getenv("ENV")
	if env == "local" {
		file, err := os.Open("config.json")
		if err != nil {
			panic(fmt.Errorf("failed to open config.json: %w", err))
		}
		defer file.Close()
		decoder := json.NewDecoder(file)
		if err := decoder.Decode(&AppConfig); err != nil {
			panic(fmt.Errorf("failed to decode config.json: %w", err))
		}
	} else {
		// Try to load from AWS Secrets Manager
		secretName := serviceName // or your actual secret name
		secretStr, err := fetchSecretsFromAWS(secretName)
		if err == nil {
			json.Unmarshal([]byte(secretStr), &AppConfig)
		} else {
			// Fallback to env vars if AWS fails
			AppConfig = Config{
				KafkaBrokers: []string{getEnvOrDefault("KAFKA_BROKERS", "localhost:9092")},
				KafkaTopic:   getEnvOrDefault("KAFKA_TOPIC", "vehicle-stock"),
				MongoURI:     getEnvOrDefault("MONGO_URI", "mongodb://localhost:27017"),
				MongoDB:      getEnvOrDefault("MONGO_DB", "vehicle_stock_db"),
				MongoColl:    getEnvOrDefault("MONGO_COLLECTION", "stock_data"),
				StripeKey:    os.Getenv("STRIPE_KEY"),
			}
		}
	}
}

// getEnvOrDefault returns the value of the environment variable or the default if not set
func getEnvOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}

var fetchSecretsFromAWS = func(secretName string) (string, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String(os.Getenv("AWS_REGION")),
	})
	if err != nil {
		return "", err
	}
	svc := secretsmanager.New(sess)
	input := &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretName),
	}
	result, err := svc.GetSecretValue(input)
	if err != nil {
		return "", err
	}
	return *result.SecretString, nil
}
