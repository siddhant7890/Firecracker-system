package billing

import "time"

type Status string

const (
	StatusPending  Status = "pending"
	StatusApproved Status = "approved"
	StatusRejected Status = "rejected"
)

type PaymentMode string

const (
	PaymentCash PaymentMode = "cash"
	PaymentUPI  PaymentMode = "upi"
)

// Bill is one entry in "Bill History" / "Recent Bills" / "Cash Counter".
type Bill struct {
	ID                int          `json:"id"`
	AdminID           int          `json:"-"`
	SalesStaffID      int          `json:"sales_staff_id"`
	SalesStaffName    string       `json:"sales_staff_name,omitempty"`
	BillNo            string       `json:"bill_no"`
	FinancialYear     string       `json:"financial_year"`
	CustomerName      string       `json:"customer_name"`
	CustomerMobile    string       `json:"customer_mobile,omitempty"`
	TokenNumber       string       `json:"token_number,omitempty"`
	NumberOfCartoon   int          `json:"number_of_cartoon"`
	GSTNumber         string       `json:"gst_number,omitempty"`
	TaxableAmount     float64      `json:"taxable_amount"`
	CGSTAmount        float64      `json:"cgst_amount"`
	SGSTAmount        float64      `json:"sgst_amount"`
	DiscountAmount    float64      `json:"discount_amount"`
	TotalAmount       float64      `json:"total_amount"`
	Status            Status       `json:"status"`
	PaymentMode       *PaymentMode `json:"payment_mode,omitempty"`
	RazorpayOrderID   string       `json:"razorpay_order_id,omitempty"`
	RazorpayPaymentID string       `json:"razorpay_payment_id,omitempty"`
	WhatsappSent      bool         `json:"whatsapp_sent"`
	ItemCount         int          `json:"item_count"`
	Items             []BillItem   `json:"items,omitempty"`
	ApprovedAt        *time.Time   `json:"approved_at,omitempty"`
	CreatedAt         time.Time    `json:"created_at"`
}

type BillItem struct {
	ID            int     `json:"id,omitempty"`
	ProductID     int     `json:"product_id"`
	ProductName   string  `json:"product_name"`
	HSNCode       string  `json:"hsn_code"`
	Qty           int     `json:"qty"`
	Rate          float64 `json:"rate"`
	TaxableAmount float64 `json:"taxable_amount"`
	GSTPercent    float64 `json:"gst_percent"`
	CGSTPercent   float64 `json:"cgst_percent"`
	CGSTAmount    float64 `json:"cgst_amount"`
	SGSTPercent   float64 `json:"sgst_percent"`
	SGSTAmount    float64 `json:"sgst_amount"`
	TotalAmount   float64 `json:"total_amount"`
}

// CreateBillItemRequest is what the "New Bill" cart sends per line. The
// frontend computes the full GST breakup for the line (rate, taxable amount,
// CGST/SGST split, total) and the backend just persists it as-is; only the
// product_id is verified server-side (must belong to this shop and be active).
type CreateBillItemRequest struct {
	ProductID     int     `json:"product_id" binding:"required"`
	Qty           int     `json:"qty" binding:"required,gt=0"`
	Rate          float64 `json:"rate" binding:"required,gt=0"`
	TaxableAmount float64 `json:"taxable_amount" binding:"required,gt=0"`
	GSTPercent    float64 `json:"gst_percent" binding:"gte=0"`
	CGSTPercent   float64 `json:"cgst_percent" binding:"gte=0"`
	CGSTAmount    float64 `json:"cgst_amount" binding:"gte=0"`
	SGSTPercent   float64 `json:"sgst_percent" binding:"gte=0"`
	SGSTAmount    float64 `json:"sgst_amount" binding:"gte=0"`
	TotalAmount   float64 `json:"total_amount" binding:"required,gt=0"`
}

type CreateBillRequest struct {
	CustomerName    string                  `json:"customer_name"`
	CustomerMobile  string                  `json:"customer_mobile"`
	TokenNumber     string                  `json:"token_number"`
	NumberOfCartoon int                     `json:"number_of_cartoon" binding:"gte=0"`
	GSTNumber       string                  `json:"gst_number"`
	DiscountAmount  float64                 `json:"discount_amount" binding:"gte=0"`
	Items           []CreateBillItemRequest `json:"items" binding:"required,min=1,dive"`
}

type ApproveBillRequest struct {
	PaymentMode PaymentMode `json:"payment_mode" binding:"required,oneof=cash upi"`
}

// UpdateBillRequest lets admin or sales staff correct a bill after it's been
// created — the payment mode (e.g. cash entered by mistake instead of UPI)
// and/or its customer/header details (name, mobile, token, carton count, GST
// number, discount). Every field is optional/partial: omit a field to leave
// it unchanged. Line items and their GST breakup are not editable here;
// changing discount_amount recalculates total_amount.
type UpdateBillRequest struct {
	PaymentMode     *PaymentMode `json:"payment_mode,omitempty" binding:"omitempty,oneof=cash upi"`
	CustomerName    *string      `json:"customer_name,omitempty"`
	CustomerMobile  *string      `json:"customer_mobile,omitempty"`
	TokenNumber     *string      `json:"token_number,omitempty"`
	NumberOfCartoon *int         `json:"number_of_cartoon,omitempty" binding:"omitempty,gte=0"`
	GSTNumber       *string      `json:"gst_number,omitempty"`
	DiscountAmount  *float64     `json:"discount_amount,omitempty" binding:"omitempty,gte=0"`
}

// DashboardHome matches the mobile "Home" screen for a single sales agent.
type DashboardHome struct {
	TodaysBills int     `json:"todays_bills"`
	TodaysSales float64 `json:"todays_sales"`
	RecentBills []Bill  `json:"recent_bills"`
}
