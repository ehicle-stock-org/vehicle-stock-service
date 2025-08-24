package service

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/yourusername/vehicle-stock-service/internal/config"
	"github.com/yourusername/vehicle-stock-service/internal/kafka"
	"github.com/yourusername/vehicle-stock-service/internal/models"
	"github.com/yourusername/vehicle-stock-service/internal/mongo"
)

// KafkaPublisher abstracts Publish method for Kafka producer
type KafkaPublisher interface {
	Publish(key string, value []byte)
	Close()
}

// SendStockDataFromVehicles generates stock data for all active subscriptions
func SendStockDataFromVehicles(jsonInput string, prod KafkaPublisher) {
	var data models.VehicleResponse
	if err := json.Unmarshal([]byte(jsonInput), &data); err != nil {
		log.Println("Error parsing vehicle JSON:", err)
		return
	}

	for _, v := range data.Payload.VehicleSubscriptions {
		if v.ActivePaidSubscriptions {
			stock := models.StockData{
				Ticker: fmt.Sprintf("VEHICLE-%s", v.Vin),
				Bid:    100.0 + float64(time.Now().Second())*0.1,
				Ask:    101.0 + float64(time.Now().Second())*0.1,
				Time:   time.Now().Format(time.RFC3339),
			}

			value, _ := json.Marshal(stock)
			prod.Publish(stock.Ticker, value)
			log.Println("Stock sent to Kafka:", stock)

			if mongo.Client != nil {
				if err := mongo.InsertData(config.AppConfig.MongoDB, config.AppConfig.MongoColl, stock); err != nil {
					log.Println("MongoDB insert failed:", err)
				}
			}
		}
	}
}

// ParseVehicleJSON checks if any active subscriptions exist
func ParseVehicleJSON(jsonInput string) (bool, error) {
	var data models.VehicleResponse
	if err := json.Unmarshal([]byte(jsonInput), &data); err != nil {
		return false, err
	}

	active := false
	for _, v := range data.Payload.VehicleSubscriptions {
		if v.ActivePaidSubscriptions {
			active = true
			break
		}
	}
	return active, nil
}

// StartStockProducerLoop starts sending stock data to Kafka periodically
func StartStockProducerLoop(jsonInput string, interval time.Duration) {
	prod, err := kafka.NewProducer(config.AppConfig.KafkaBrokers[0], config.AppConfig.KafkaTopic)
	if err != nil {
		log.Println("Kafka producer initialization failed:", err)
		return
	}

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		defer prod.Close()

		for range ticker.C {
			SendStockDataFromVehicles(jsonInput, prod)
		}
	}()
}
