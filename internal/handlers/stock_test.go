package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/vehicle-stock-service/internal/models"
	"github.com/yourusername/vehicle-stock-service/internal/mongo"
)

func mockFindStockByTickerAndDate(database, collection, ticker, date string) (*models.StockData, error) {
	return &models.StockData{Ticker: ticker, Bid: 100.0, Ask: 101.0, Time: date}, nil
}

func TestGetStockHandlerHappyPath(t *testing.T) {
	// Patch mongo.FindStockByTickerAndDate
	orig := mongo.FindStockByTickerAndDate
	mongo.FindStockByTickerAndDate = mockFindStockByTickerAndDate
	defer func() { mongo.FindStockByTickerAndDate = orig }()

	req := httptest.NewRequest("GET", "/getstock", nil)
	req.Header.Set("startDate", "2025-08-01")
	req.Header.Set("endDate", "2025-08-24")
	rw := httptest.NewRecorder()

	GetStockHandler(rw, req)
	resp := rw.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&body)
	assert.NoError(t, err)

	assert.Equal(t, "2025-08-01", body["startDate"])
	assert.Equal(t, "2025-08-24", body["endDate"])
	assert.Contains(t, body, "activePaidSubscriptions")
	assert.Contains(t, body, "vehicleStocks")
	assert.Contains(t, body, "message")
	assert.Contains(t, body, "timestamp")

	// Check vehicleStocks structure
	vs, ok := body["vehicleStocks"].([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, vs)
}

func TestGetStockHandlerMissingHeaders(t *testing.T) {
	req := httptest.NewRequest("GET", "/getstock", nil)
	rw := httptest.NewRecorder()
	GetStockHandler(rw, req)
	resp := rw.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
