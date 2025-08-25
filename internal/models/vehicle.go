package models

// VehicleSubscription represents a single vehicle subscription in the JSON response
type VehicleSubscription struct {
	VehicleStatus               string `json:"vehicleStatus"`
	Generation                  string `json:"generation,omitempty"`
	Region                      string `json:"region,omitempty"` // optional
	Vin                         string `json:"vin"`
	IsSafetyActive              bool   `json:"isSafetyActive,omitempty"`
	IsServiceConnectActive      bool   `json:"isServiceConnectActive,omitempty"`
	IsRemoteActive              bool   `json:"isRemoteActive,omitempty"`
	IsDigitalKeyRemoteActive    bool   `json:"isDigitalKeyRemoteActive,omitempty"`
	IsDestinationAssistActive   bool   `json:"isDestinationAssistActive,omitempty"`
	IsNavigationActive          bool   `json:"isNavigationActive,omitempty"`
	IsVirtualAssistantActive    bool   `json:"isVirtualAssistantActive,omitempty"`
	IsIntegratedStreamingActive bool   `json:"isIntegratedStreamingActive,omitempty"`
	IsWifiActive                bool   `json:"isWifiActive,omitempty"`
	Brand                       string `json:"brand,omitempty"`
	ActivePaidSubscriptions     bool   `json:"activePaidSubscriptions"`
}

// Payload represents the payload containing vehicle subscriptions
type Payload struct {
	Guid                 string                `json:"guid"`
	VehicleSubscriptions []VehicleSubscription `json:"vehicleSubscriptions"`
}

// VehicleResponse represents the entire JSON response from the REST API
type VehicleResponse struct {
	Status struct {
		Messages []struct {
			Description         string `json:"description"`
			ResponseCode        string `json:"responseCode"`
			DetailedDescription string `json:"detailedDescription,omitempty"`
		} `json:"messages"`
	} `json:"status"`
	Payload Payload `json:"payload"`
}

// StockData represents the bid/ask stock data sent to Kafka
type StockData struct {
	Ticker string  `json:"ticker"`
	Bid    float64 `json:"bid"`
	Ask    float64 `json:"ask"`
	Time   string  `json:"time"`
}

// Helper method: Validate VIN format (simple example)
func (v VehicleSubscription) IsValidVIN() bool {
	return len(v.Vin) == 17
}
