package handlers

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
)

// HoldPaymentRequest is the expected input for /holdpayment
// Example: {"amount": 1000, "currency": "usd", "payment_method": "pm_xxx"}
type HoldPaymentRequest struct {
	Amount        int64  `json:"amount"`
	Currency      string `json:"currency"`
	PaymentMethod string `json:"payment_method"`
}

// PaymentIntentNew is a function variable for testability
var PaymentIntentNew = paymentintent.New

// HoldPaymentHandler places a hold on a payment method using Stripe manual capture
func HoldPaymentHandler(w http.ResponseWriter, r *http.Request) {
	var req HoldPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "Invalid request body"})
		return
	}

	stripe.Key = os.Getenv("STRIPE_KEY")
	if stripe.Key == "" {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"error": "Stripe key not set"})
		return
	}

	params := &stripe.PaymentIntentParams{
		Amount:        stripe.Int64(req.Amount),
		Currency:      stripe.String(req.Currency),
		PaymentMethod: stripe.String(req.PaymentMethod),
		CaptureMethod: stripe.String("manual"),
		Confirm:       stripe.Bool(true),
		AutomaticPaymentMethods: &stripe.PaymentIntentAutomaticPaymentMethodsParams{
			Enabled:        stripe.Bool(true),
			AllowRedirects: stripe.String("never"),
		},
	}
	pi, err := PaymentIntentNew(params)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		return
	}

	json.NewEncoder(w).Encode(map[string]interface{}{
		"payment_intent_id": pi.ID,
		"status":            pi.Status,
		"amount":            pi.Amount,
		"currency":          pi.Currency,
	})
}
