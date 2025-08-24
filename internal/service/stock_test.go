package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/yourusername/vehicle-stock-service/internal/config"
	"github.com/yourusername/vehicle-stock-service/internal/models"
)

// importConfig sets dummy config values for testing
func importConfig() {
	config.AppConfig.KafkaBrokers = []string{"dummy-broker:9092"}
	config.AppConfig.KafkaTopic = "dummy-topic"
}

// Mock Kafka Producer
type MockProducer struct {
	mock.Mock
	Published []models.StockData
}

func (m *MockProducer) Publish(key string, value []byte) {
	var stock models.StockData
	_ = json.Unmarshal(value, &stock)
	m.Published = append(m.Published, stock)
	m.Called(key, value)
}

// Close is a no-op for the mock producer (required to satisfy interface)
func (m *MockProducer) Close() {
	// No resources to release in mock
}

func TestParseVehicleJSON(t *testing.T) {
	var active bool
	var err error
	// Edge: empty payload
	emptyPayload := `{"payload":{}}`
	active, err = ParseVehicleJSON(emptyPayload)
	assert.NoError(t, err)
	assert.False(t, active)

	// Edge: missing vehicleSubscriptions
	missingSubs := `{"payload":{"guid":"abc"}}`
	active, err = ParseVehicleJSON(missingSubs)
	assert.NoError(t, err)
	assert.False(t, active)

	// Edge: malformed JSON
	malformed := `{"payload":{"vehicleSubscriptions":[]` // missing closing braces
	active, err = ParseVehicleJSON(malformed)
	assert.Error(t, err)
	assert.False(t, active)

	// Edge: multiple actives
	multiActive := `{"payload":{"vehicleSubscriptions":[{"vin":"1","activePaidSubscriptions":true},{"vin":"2","activePaidSubscriptions":true}]}}`
	active, err = ParseVehicleJSON(multiActive)
	assert.NoError(t, err)
	assert.True(t, active)

	// Edge: all inactive
	allInactive := `{"payload":{"vehicleSubscriptions":[{"vin":"1","activePaidSubscriptions":false},{"vin":"2","activePaidSubscriptions":false}]}}`
	active, err = ParseVehicleJSON(allInactive)
	assert.NoError(t, err)
	assert.False(t, active)

	// Edge: nil input
	active, err = ParseVehicleJSON("null")
	assert.NoError(t, err)
	assert.False(t, active)
	validJSON := `{"payload":{"vehicleSubscriptions":[{"vin":"123","activePaidSubscriptions":true}]}}`
	noActiveJSON := `{"payload":{"vehicleSubscriptions":[{"vin":"123","activePaidSubscriptions":false}]}}`
	invalidJSON := `{"payload":{"vehicleSubscriptions":[{"vin":"123"}]}` // missing closing

	// Happy path
	active, err = ParseVehicleJSON(validJSON)
	assert.NoError(t, err)
	assert.True(t, active)

	// No active subscriptions
	active, err = ParseVehicleJSON(noActiveJSON)
	assert.NoError(t, err)
	assert.False(t, active)

	// Invalid JSON
	active, err = ParseVehicleJSON(invalidJSON)
	assert.Error(t, err)
	assert.False(t, active)

	// Empty JSON
	active, err = ParseVehicleJSON("")
	assert.Error(t, err)
	assert.False(t, active)

	// Multiple vehicles, one active
	multiJSON := `{"payload":{"vehicleSubscriptions":[{"vin":"1","activePaidSubscriptions":false},{"vin":"2","activePaidSubscriptions":true}]}}`
	active, err = ParseVehicleJSON(multiJSON)
	assert.NoError(t, err)
	assert.True(t, active)
}

func TestSendStockDataFromVehicles(t *testing.T) {
	var mockProd *MockProducer
	mockProd = &MockProducer{}
	mockProd.On("Publish", mock.Anything, mock.Anything)
	var noPayload, emptySubs, extraFields, dupVIN, allInactive, allActive, mixed string
	// Edge: input with no payload
	mockProd.Published = nil
	noPayload = `{"foo":123}`
	SendStockDataFromVehicles(noPayload, mockProd)
	assert.Len(t, mockProd.Published, 0)

	// Edge: input with empty vehicleSubscriptions
	mockProd.Published = nil
	emptySubs = `{"payload":{"vehicleSubscriptions":[]}}`
	SendStockDataFromVehicles(emptySubs, mockProd)
	assert.Len(t, mockProd.Published, 0)

	// Edge: input with extra fields
	mockProd.Published = nil
	extraFields = `{"payload":{"vehicleSubscriptions":[{"vin":"VINX","activePaidSubscriptions":true,"extra":123}]}}`
	SendStockDataFromVehicles(extraFields, mockProd)
	assert.Len(t, mockProd.Published, 1)
	assert.Equal(t, "VEHICLE-VINX", mockProd.Published[0].Ticker)

	// Edge: input with duplicate VINs
	mockProd.Published = nil
	dupVIN = `{"payload":{"vehicleSubscriptions":[{"vin":"VINY","activePaidSubscriptions":true},{"vin":"VINY","activePaidSubscriptions":true}]}}`
	SendStockDataFromVehicles(dupVIN, mockProd)
	assert.Len(t, mockProd.Published, 2)
	assert.Equal(t, "VEHICLE-VINY", mockProd.Published[0].Ticker)
	assert.Equal(t, "VEHICLE-VINY", mockProd.Published[1].Ticker)

	// Edge: input with all inactive
	mockProd.Published = nil
	allInactive = `{"payload":{"vehicleSubscriptions":[{"vin":"VINZ","activePaidSubscriptions":false}]}}`
	SendStockDataFromVehicles(allInactive, mockProd)
	assert.Len(t, mockProd.Published, 0)

	// Edge: input with all active
	mockProd.Published = nil
	allActive = `{"payload":{"vehicleSubscriptions":[{"vin":"VINA","activePaidSubscriptions":true},{"vin":"VINB","activePaidSubscriptions":true}]}}`
	SendStockDataFromVehicles(allActive, mockProd)
	assert.Len(t, mockProd.Published, 2)
	assert.Equal(t, "VEHICLE-VINA", mockProd.Published[0].Ticker)
	assert.Equal(t, "VEHICLE-VINB", mockProd.Published[1].Ticker)

	// Edge: input with mixed valid/invalid vehicles
	mockProd.Published = nil
	mixed = `{"payload":{"vehicleSubscriptions":[{"vin":"VIN1","activePaidSubscriptions":true},{"vin":"VIN2"}]}}`
	SendStockDataFromVehicles(mixed, mockProd)
	assert.Len(t, mockProd.Published, 1)
	assert.Equal(t, "VEHICLE-VIN1", mockProd.Published[0].Ticker)
	// Setup mock producer implementing KafkaPublisher
	mockProd = &MockProducer{}
	mockProd.On("Publish", mock.Anything, mock.Anything)

	// Valid input with active subscription
	jsonInput := `{"payload":{"vehicleSubscriptions":[{"vin":"VIN1","activePaidSubscriptions":true}]}}`
	SendStockDataFromVehicles(jsonInput, mockProd)
	assert.Len(t, mockProd.Published, 1)
	assert.Equal(t, "VEHICLE-VIN1", mockProd.Published[0].Ticker)

	// Valid input with no active subscription
	mockProd.Published = nil
	jsonInput = `{"payload":{"vehicleSubscriptions":[{"vin":"VIN2","activePaidSubscriptions":false}]}}`
	SendStockDataFromVehicles(jsonInput, mockProd)
	assert.Len(t, mockProd.Published, 0)

	// Multiple vehicles, mixed active
	mockProd.Published = nil
	jsonInput = `{"payload":{"vehicleSubscriptions":[{"vin":"VIN3","activePaidSubscriptions":false},{"vin":"VIN4","activePaidSubscriptions":true}]}}`
	SendStockDataFromVehicles(jsonInput, mockProd)
	assert.Len(t, mockProd.Published, 1)
	assert.Equal(t, "VEHICLE-VIN4", mockProd.Published[0].Ticker)

	// Invalid JSON
	mockProd.Published = nil
	jsonInput = `{"payload":{"vehicleSubscriptions":[{"vin":"VIN5"}]}`
	SendStockDataFromVehicles(jsonInput, mockProd)
	assert.Len(t, mockProd.Published, 0)

	// Empty input
	mockProd.Published = nil
	SendStockDataFromVehicles("", mockProd)
	assert.Len(t, mockProd.Published, 0)
	// Edge: input with no payload
	mockProd.Published = nil
	noPayload = `{"foo":123}`
	SendStockDataFromVehicles(noPayload, mockProd)
	assert.Len(t, mockProd.Published, 0)

	// Edge: input with empty vehicleSubscriptions
	mockProd.Published = nil
	emptySubs = `{"payload":{"vehicleSubscriptions":[]}}`
	SendStockDataFromVehicles(emptySubs, mockProd)
	assert.Len(t, mockProd.Published, 0)

	// Edge: input with extra fields
	mockProd.Published = nil
	extraFields = `{"payload":{"vehicleSubscriptions":[{"vin":"VINX","activePaidSubscriptions":true,"extra":123}]}}`
	SendStockDataFromVehicles(extraFields, mockProd)
	assert.Len(t, mockProd.Published, 1)
	assert.Equal(t, "VEHICLE-VINX", mockProd.Published[0].Ticker)

	// Edge: input with duplicate VINs
	mockProd.Published = nil
	dupVIN = `{"payload":{"vehicleSubscriptions":[{"vin":"VINY","activePaidSubscriptions":true},{"vin":"VINY","activePaidSubscriptions":true}]}}`
	SendStockDataFromVehicles(dupVIN, mockProd)
	assert.Len(t, mockProd.Published, 2)
	assert.Equal(t, "VEHICLE-VINY", mockProd.Published[0].Ticker)
	assert.Equal(t, "VEHICLE-VINY", mockProd.Published[1].Ticker)

	// Edge: input with all inactive
	mockProd.Published = nil
	allInactive = `{"payload":{"vehicleSubscriptions":[{"vin":"VINZ","activePaidSubscriptions":false}]}}`
	SendStockDataFromVehicles(allInactive, mockProd)
	assert.Len(t, mockProd.Published, 0)

	// Edge: input with all active
	mockProd.Published = nil
	allActive = `{"payload":{"vehicleSubscriptions":[{"vin":"VINA","activePaidSubscriptions":true},{"vin":"VINB","activePaidSubscriptions":true}]}}`
	SendStockDataFromVehicles(allActive, mockProd)
	assert.Len(t, mockProd.Published, 2)
	assert.Equal(t, "VEHICLE-VINA", mockProd.Published[0].Ticker)
	assert.Equal(t, "VEHICLE-VINB", mockProd.Published[1].Ticker)

	// Edge: input with mixed valid/invalid vehicles
	mockProd.Published = nil
	mixed = `{"payload":{"vehicleSubscriptions":[{"vin":"VIN1","activePaidSubscriptions":true},{"vin":"VIN2"}]}}`
	SendStockDataFromVehicles(mixed, mockProd)
	assert.Len(t, mockProd.Published, 1)
	assert.Equal(t, "VEHICLE-VIN1", mockProd.Published[0].Ticker)
}

func TestStartStockProducerLoopInit(t *testing.T) {
	// This test only checks initialization, not the actual loop
	// The actual periodic sending is best tested with integration tests or with a timer mock
	// Here, we just ensure no panic and correct producer creation
	// Set dummy config values to prevent index out of range
	importConfig()
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("StartStockProducerLoop panicked: %v", r)
		}
	}()
	go StartStockProducerLoop(`{"payload":{"vehicleSubscriptions":[]}}`, 1*time.Second)
	// Allow goroutine to start
	time.Sleep(100 * time.Millisecond)
	// No assertion, just ensure no panic
}
