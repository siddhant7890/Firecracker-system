package payment

import (
	"context"
	"errors"
	"fmt"

	"salestrack/config"
	"salestrack/internal/billing"
)

var ErrSignatureMismatch = errors.New("payment signature did not match; cannot trust this payment")

type Service struct {
	billing *billing.Service
}

func NewService(billingSvc *billing.Service) *Service {
	return &Service{billing: billingSvc}
}

// CreateOrderForBill opens a Razorpay order for a pending bill's total, so
// the admin can show the customer a UPI checkout/QR. The bill itself stays
// "pending" until the payment is confirmed (via webhook or VerifyPayment).
func (s *Service) CreateOrderForBill(ctx context.Context, adminID, billID int) (CreateOrderResponse, error) {
	bill, err := s.billing.Get(ctx, adminID, billID)
	if err != nil {
		return CreateOrderResponse{}, err
	}
	if bill.Status != billing.StatusPending {
		return CreateOrderResponse{}, fmt.Errorf("bill is not pending approval")
	}

	order, err := CreateOrder(bill.TotalAmount, bill.BillNo)
	if err != nil {
		return CreateOrderResponse{}, fmt.Errorf("razorpay order creation failed: %w", err)
	}
	orderID, _ := order["id"].(string)
	if orderID == "" {
		return CreateOrderResponse{}, errors.New("razorpay did not return an order id")
	}

	if err := s.billing.SetRazorpayOrder(ctx, adminID, billID, orderID); err != nil {
		return CreateOrderResponse{}, err
	}
	bill.RazorpayOrderID = orderID

	amountPaise := int(bill.TotalAmount*100 + 0.5)
	return CreateOrderResponse{
		OrderID:     orderID,
		AmountPaise: amountPaise,
		Currency:    "INR",
		KeyID:       config.App.RazorpayKeyID,
		Bill:        bill,
	}, nil
}

// VerifyPayment is called by the admin webapp right after Razorpay
// Checkout.js reports success client-side. It's a fast path; the webhook
// below is the authoritative, server-to-server confirmation.
func (s *Service) VerifyPayment(ctx context.Context, req VerifyPaymentRequest) (billing.Bill, error) {
	if !VerifyPaymentSignature(req.RazorpayOrderID, req.RazorpayPaymentID, req.RazorpaySignature) {
		return billing.Bill{}, ErrSignatureMismatch
	}
	return s.billing.ApproveByRazorpayOrder(ctx, req.RazorpayOrderID, req.RazorpayPaymentID)
}

// HandleWebhookPayment processes a "payment.captured" event from Razorpay's
// webhook (server-to-server, doesn't depend on the admin's browser staying
// open).
func (s *Service) HandleWebhookPayment(ctx context.Context, orderID, paymentID string) (billing.Bill, error) {
	return s.billing.ApproveByRazorpayOrder(ctx, orderID, paymentID)
}
