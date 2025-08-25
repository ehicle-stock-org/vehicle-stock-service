package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stripe/stripe-go/v78"
)

func mockPaymentIntentNew(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
	return &stripe.PaymentIntent{
		ID:       "pi_test_123",
		Status:   "requires_capture",
		Amount:   1000,
		Currency: "usd",
	}, nil
}

func TestHoldPaymentHandlerHappyPath(t *testing.T) {
	os.Setenv("STRIPE_KEY", "sk_test_123")
	orig := PaymentIntentNew
	PaymentIntentNew = mockPaymentIntentNew
	defer func() { PaymentIntentNew = orig }()

	body := HoldPaymentRequest{
		Amount:        1000,
		Currency:      "usd",
		PaymentMethod: "pm_test_123",
	}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/holdpayment", bytes.NewReader(b))
	rw := httptest.NewRecorder()
	HoldPaymentHandler(rw, req)
	resp := rw.Result()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var respBody map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&respBody)
	assert.NoError(t, err)
	assert.Equal(t, "pi_test_123", respBody["payment_intent_id"])
	assert.Equal(t, "requires_capture", respBody["status"])
	assert.Equal(t, float64(1000), respBody["amount"])
	assert.Equal(t, "usd", respBody["currency"])
}

func TestHoldPaymentHandlerInvalidBody(t *testing.T) {
	os.Setenv("STRIPE_KEY", "sk_test_123")
	req := httptest.NewRequest("POST", "/holdpayment", bytes.NewReader([]byte("invalid-json")))
	rw := httptest.NewRecorder()
	HoldPaymentHandler(rw, req)
	resp := rw.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHoldPaymentHandlerMissingStripeKey(t *testing.T) {
	os.Unsetenv("STRIPE_KEY")
	body := HoldPaymentRequest{Amount: 1000, Currency: "usd", PaymentMethod: "pm_test_123"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/holdpayment", bytes.NewReader(b))
	rw := httptest.NewRecorder()
	HoldPaymentHandler(rw, req)
	resp := rw.Result()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestHoldPaymentHandlerStripeError(t *testing.T) {
	os.Setenv("STRIPE_KEY", "sk_test_123")
	orig := PaymentIntentNew
	PaymentIntentNew = func(params *stripe.PaymentIntentParams) (*stripe.PaymentIntent, error) {
		return nil, assert.AnError
	}
	defer func() { PaymentIntentNew = orig }()
	body := HoldPaymentRequest{Amount: 1000, Currency: "usd", PaymentMethod: "pm_test_123"}
	b, _ := json.Marshal(body)
	req := httptest.NewRequest("POST", "/holdpayment", bytes.NewReader(b))
	rw := httptest.NewRecorder()
	HoldPaymentHandler(rw, req)
	resp := rw.Result()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
