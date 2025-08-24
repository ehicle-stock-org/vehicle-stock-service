package config

import (
	"log"
	"os"
)

// Config holds all app-wide configuration
type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	MongoURI     string
	MongoDB      string
	MongoColl    string
	StripeKey    string
}

// AppConfig is the exported global configuration
var AppConfig Config

// LoadConfig loads configuration from environment variables
func LoadConfig() {
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:9092" // default local broker
	}

	topic := os.Getenv("KAFKA_TOPIC")
	if topic == "" {
		topic = "vehicle-stock" // default topic
	}

	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	mongoDB := os.Getenv("MONGO_DB")
	if mongoDB == "" {
		mongoDB = "vehicle_stock_db"
	}

	mongoColl := os.Getenv("MONGO_COLLECTION")
	if mongoColl == "" {
		mongoColl = "stock_data"
	}

	stripeKey := os.Getenv("STRIPE_KEY")
	if stripeKey == "" {
		log.Println("Warning: STRIPE_KEY is not set. Stripe operations will fail.")
	}

	AppConfig = Config{
		KafkaBrokers: []string{brokers},
		KafkaTopic:   topic,
		MongoURI:     mongoURI,
		MongoDB:      mongoDB,
		MongoColl:    mongoColl,
		StripeKey:    stripeKey,
	}

	log.Println("Configuration loaded successfully")
}
