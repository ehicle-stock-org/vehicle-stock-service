package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/yourusername/vehicle-stock-service/internal/config"
	"github.com/yourusername/vehicle-stock-service/internal/handlers"
	"github.com/yourusername/vehicle-stock-service/internal/mongo"
	"github.com/yourusername/vehicle-stock-service/internal/service"
)

func main() {
	// Load configuration
	config.LoadConfig()

	// Connect to MongoDB
	if _, err := mongo.ConnectMongo(config.AppConfig.MongoURI); err != nil {
		log.Fatal("MongoDB connection failed:", err)
	}

	// Start stock producer loop in background
	jsonInput := `{
			"status": {
				"messages": [
					{
						"description": "Request Processed Successfully",
						"responseCode": "SUB-0000",
						"detailedDescription": "Request Processed Successfully"
					}
				]
			},
			"payload": {
				"guid": "200d617c92c9a889cdda4c31559472f",
				"vehicleSubscriptions": [
					{
						"vehicleStatus": "SUBSCRIBED",
						"region": "US",
						"vin": "AA450000007141513",
						"activePaidSubscriptions": true
					}
				]
			}
		}`
	service.StartStockProducerLoop(jsonInput, 30*time.Second)

	// Initialize router
	r := mux.NewRouter()

	// Robust CORS middleware with logging
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, startDate, endDate")
			w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
			log.Printf("CORS middleware executed for %s %s", req.Method, req.URL.Path)
			if req.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, req)
		})
	})

	// Register /getstock endpoint
	r.HandleFunc("/getstock", handlers.GetStockHandler).Methods("GET")

	// Register /holdpayment endpoint for Stripe payment hold
	r.HandleFunc("/holdpayment", handlers.HoldPaymentHandler).Methods("POST")

	// Start HTTP server
	log.Println("REST API running on http://localhost:8080/getstock")
	log.Fatal(http.ListenAndServe(":8080", r))
}
