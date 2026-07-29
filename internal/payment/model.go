package payment

import "salestrack/internal/billing"

// CreateOrderResponse is everything the admin webapp needs to open Razorpay
// Checkout for the customer (amount is in paise, as Razorpay expects).
type CreateOrderResponse struct {
	OrderID  string  `json:"order_id"`
	AmountPaise int  `json:"amount_paise"`
	Currency string  `json:"currency"`
	KeyID    string  `json:"key_id"`
	Bill     billing.Bill `json:"bill"`
}

// VerifyPaymentRequest is the payload Razorpay Checkout.js hands back to the
// frontend on a successful payment (handler.success callback).
type VerifyPaymentRequest struct {
	RazorpayOrderID   string `json:"razorpay_order_id" binding:"required"`
	RazorpayPaymentID string `json:"razorpay_payment_id" binding:"required"`
	RazorpaySignature string `json:"razorpay_signature" binding:"required"`
}
