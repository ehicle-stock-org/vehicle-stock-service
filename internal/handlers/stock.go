package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/yourusername/vehicle-stock-service/internal/mongo"
)

// vehiclePayloadSource returns the JSON payload for vehicles (can be mocked in tests)
var vehiclePayloadSource = func() string {
	return `{
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
				   "generation": "24MM",
				   "region": "US",
				   "vin": "AA450000007141513",
				   "isSafetyActive": true,
				   "isServiceConnectActive": false,
				   "isRemoteActive": false,
				   "isDigitalKeyRemoteActive": false,
				   "isDestinationAssistActive": false,
				   "isNavigationActive": false,
				   "isVirtualAssistantActive": false,
				   "isIntegratedStreamingActive": false,
				   "isWifiActive": false,
				   "brand": "L",
				   "activePaidSubscriptions": true
			   },
			   {
				   "vehicleStatus": "SUBSCRIBED",
				   "generation": "24MM",
				   "region": "CA",
				   "vin": "AA450000007141573",
				   "isSafetyActive": true,
				   "isServiceConnectActive": false,
				   "isRemoteActive": false,
				   "isDigitalKeyRemoteActive": false,
				   "isDestinationAssistActive": false,
				   "isNavigationActive": false,
				   "isVirtualAssistantActive": false,
				   "isIntegratedStreamingActive": false,
				   "isWifiActive": false,
				   "brand": "T",
				   "activePaidSubscriptions": false
			   },
			   {
				   "vehicleStatus": "SUBSCRIBED",
				   "generation": "24MM",
				   "region": "CA",
				   "vin": "AA450000007141603",
				   "isSafetyActive": true
			   }
		   ]
	   }
	}`
}

// GetStockHandler handles /getstock requests
func GetStockHandler(w http.ResponseWriter, r *http.Request) {
	startDate := r.Header.Get("startDate")
	endDate := r.Header.Get("endDate")
	if startDate == "" || endDate == "" {
		http.Error(w, "startDate and endDate headers are required", http.StatusBadRequest)
		return
	}

	log.Printf("Fetching stock data from %s to %s\n", startDate, endDate)

	// Use the payload source (can be mocked in tests)
	jsonInput := vehiclePayloadSource()

	// Parse vehicle payload
	var vehicleResp struct {
		Status  interface{} `json:"status"`
		Payload struct {
			Guid                 string                   `json:"guid"`
			VehicleSubscriptions []map[string]interface{} `json:"vehicleSubscriptions"`
		} `json:"payload"`
	}
	if err := json.Unmarshal([]byte(jsonInput), &vehicleResp); err != nil {
		http.Error(w, "Failed to parse vehicle payload", http.StatusInternalServerError)
		return
	}

	// For each vehicle, fetch stock data for startDate and endDate
	type VehicleStock struct {
		VIN        string `json:"vin"`
		Region     string `json:"region,omitempty"`
		StartPrice *struct {
			Bid float64 `json:"bid"`
			Ask float64 `json:"ask"`
		} `json:"startPrice,omitempty"`
		EndPrice *struct {
			Bid float64 `json:"bid"`
			Ask float64 `json:"ask"`
		} `json:"endPrice,omitempty"`
		Difference *struct {
			Bid float64 `json:"bid"`
			Ask float64 `json:"ask"`
		} `json:"difference,omitempty"`
	}
	var vehicleStocks []VehicleStock
	for _, v := range vehicleResp.Payload.VehicleSubscriptions {
		vin, _ := v["vin"].(string)
		region, _ := v["region"].(string)
		ticker := "VEHICLE-" + vin

		// Fetch start and end price from MongoDB
		startStock, _ := mongo.FindStockByTickerAndDate("vehicle_stock_db", "stock_data", ticker, startDate)
		endStock, _ := mongo.FindStockByTickerAndDate("vehicle_stock_db", "stock_data", ticker, endDate)

		var startPrice, endPrice, diff *struct {
			Bid float64 `json:"bid"`
			Ask float64 `json:"ask"`
		}
		if startStock != nil {
			startPrice = &struct {
				Bid float64 `json:"bid"`
				Ask float64 `json:"ask"`
			}{Bid: startStock.Bid, Ask: startStock.Ask}
		}
		if endStock != nil {
			endPrice = &struct {
				Bid float64 `json:"bid"`
				Ask float64 `json:"ask"`
			}{Bid: endStock.Bid, Ask: endStock.Ask}
		}
		if startStock != nil && endStock != nil {
			diff = &struct {
				Bid float64 `json:"bid"`
				Ask float64 `json:"ask"`
			}{
				Bid: endStock.Bid - startStock.Bid,
				Ask: endStock.Ask - startStock.Ask,
			}
		}
		vehicleStocks = append(vehicleStocks, VehicleStock{
			VIN:        vin,
			Region:     region,
			StartPrice: startPrice,
			EndPrice:   endPrice,
			Difference: diff,
		})
	}

	// Check for any active paid subscriptions
	hasActive := false
	for _, v := range vehicleResp.Payload.VehicleSubscriptions {
		if active, ok := v["activePaidSubscriptions"].(bool); ok && active {
			hasActive = true
			break
		}
	}

	resp := map[string]interface{}{
		"startDate":               startDate,
		"endDate":                 endDate,
		"activePaidSubscriptions": hasActive,
		"vehiclePayload":          vehicleResp.Payload,
		"vehicleStocks":           vehicleStocks,
		"message":                 "Handler is working. Kafka & MongoDB integration running",
		"timestamp":               time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
