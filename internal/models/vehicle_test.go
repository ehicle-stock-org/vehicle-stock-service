package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVehicleSubscriptionAllFields(t *testing.T) {
	sub := VehicleSubscription{
		VehicleStatus:               "ACTIVE",
		Generation:                  "2025",
		Region:                      "EU",
		Vin:                         "VIN123",
		IsSafetyActive:              true,
		IsServiceConnectActive:      true,
		IsRemoteActive:              true,
		IsDigitalKeyRemoteActive:    true,
		IsDestinationAssistActive:   true,
		IsNavigationActive:          true,
		IsVirtualAssistantActive:    true,
		IsIntegratedStreamingActive: true,
		IsWifiActive:                true,
		Brand:                       "X",
		ActivePaidSubscriptions:     true,
	}
	data, err := json.Marshal(sub)
	assert.NoError(t, err)
	var out VehicleSubscription
	err = json.Unmarshal(data, &out)
	assert.NoError(t, err)
	assert.Equal(t, sub, out)
}

func TestVehicleResponseMultipleMessages(t *testing.T) {
	resp := VehicleResponse{}
	resp.Status.Messages = []struct {
		Description         string `json:"description"`
		ResponseCode        string `json:"responseCode"`
		DetailedDescription string `json:"detailedDescription,omitempty"`
	}{
		{Description: "OK", ResponseCode: "200", DetailedDescription: "Success"},
		{Description: "FAIL", ResponseCode: "500", DetailedDescription: "Error"},
	}
	resp.Payload = Payload{Guid: "guid456"}
	assert.Equal(t, 2, len(resp.Status.Messages))
	assert.Equal(t, "guid456", resp.Payload.Guid)
}

func TestStockDataEdgeValues(t *testing.T) {
	stock := StockData{Ticker: "TSLA", Bid: -1.0, Ask: 0.0, Time: ""}
	assert.Equal(t, "TSLA", stock.Ticker)
	assert.Equal(t, -1.0, stock.Bid)
	assert.Equal(t, 0.0, stock.Ask)
	assert.Equal(t, "", stock.Time)
}

func TestPayloadEmptySubscriptions(t *testing.T) {
	payload := Payload{Guid: "empty", VehicleSubscriptions: []VehicleSubscription{}}
	assert.Equal(t, "empty", payload.Guid)
	assert.Empty(t, payload.VehicleSubscriptions)
}

func TestVehicleSubscriptionJSONError(t *testing.T) {
	badJSON := []byte(`{"vin":123}`) // vin should be string
	var out VehicleSubscription
	err := json.Unmarshal(badJSON, &out)
	assert.Error(t, err)
}

// ...existing code...

func TestPayloadModel(t *testing.T) {
	subs := []VehicleSubscription{{VehicleStatus: "SUBSCRIBED", Vin: "VIN1"}, {VehicleStatus: "UNSUBSCRIBED", Vin: "VIN2"}}
	payload := Payload{Guid: "guid123", VehicleSubscriptions: subs}
	assert.Equal(t, "guid123", payload.Guid)
	assert.Len(t, payload.VehicleSubscriptions, 2)
}

func TestVehicleResponseModel(t *testing.T) {
	resp := VehicleResponse{}
	resp.Status.Messages = []struct {
		Description         string `json:"description"`
		ResponseCode        string `json:"responseCode"`
		DetailedDescription string `json:"detailedDescription,omitempty"`
	}{{Description: "OK", ResponseCode: "200", DetailedDescription: "Success"}}
	resp.Payload = Payload{Guid: "guid123"}
	assert.Equal(t, "OK", resp.Status.Messages[0].Description)
	assert.Equal(t, "guid123", resp.Payload.Guid)
}

func TestStockDataJSON(t *testing.T) {
	stock := StockData{Ticker: "AAPL", Bid: 150.0, Ask: 151.0, Time: "2025-08-24"}
	data, err := json.Marshal(stock)
	assert.NoError(t, err)
	var out StockData
	err = json.Unmarshal(data, &out)
	assert.NoError(t, err)
	assert.Equal(t, stock, out)
}

func TestVehicleSubscriptionEdgeCases(t *testing.T) {
	sub := VehicleSubscription{}
	data, err := json.Marshal(sub)
	assert.NoError(t, err)
	var out VehicleSubscription
	err = json.Unmarshal(data, &out)
	assert.NoError(t, err)
	assert.Equal(t, sub, out)
}
