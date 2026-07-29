package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"salestrack/config"

	razorpay "github.com/razorpay/razorpay-go"
)

// client lazily builds a Razorpay SDK client from the configured keys.
func client() *razorpay.Client {
	return razorpay.NewClient(config.App.RazorpayKeyID, config.App.RazorpayKeySecret)
}

// CreateOrder opens a Razorpay order for a bill's total amount (rupees ->
// paise) so the admin portal can render Razorpay Checkout for the customer
// to pay via UPI.
func CreateOrder(amountRupees float64, receipt string) (map[string]interface{}, error) {
	amountPaise := int(amountRupees*100 + 0.5)
	data := map[string]interface{}{
		"amount":          amountPaise,
		"currency":        "INR",
		"receipt":         receipt,
		"payment_capture": 1,
	}
	return client().Order.Create(data, nil)
}

// VerifyPaymentSignature checks the signature Razorpay Checkout.js returns
// to the frontend after a successful payment: HMAC-SHA256(order_id+"|"+payment_id)
// keyed with the Razorpay key secret. See Razorpay's "Verify Payment
// Signature" docs.
func VerifyPaymentSignature(orderID, paymentID, signature string) bool {
	mac := hmac.New(sha256.New, []byte(config.App.RazorpayKeySecret))
	mac.Write([]byte(orderID + "|" + paymentID))
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}

// VerifyWebhookSignature checks the X-Razorpay-Signature header on incoming
// webhook calls: HMAC-SHA256(raw request body) keyed with the webhook
// secret configured in the Razorpay dashboard.
func VerifyWebhookSignature(rawBody []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(config.App.RazorpayWebhookKey))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}
