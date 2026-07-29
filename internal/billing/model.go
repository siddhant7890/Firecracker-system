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
	ID                int         `json:"id"`
	AdminID           int         `json:"-"`
	SalesStaffID      int         `json:"sales_staff_id"`
	SalesStaffName    string      `json:"sales_staff_name,omitempty"`
	BillNo            string      `json:"bill_no"`
	FinancialYear     string      `json:"financial_year"`
	CustomerName      string      `json:"customer_name"`
	CustomerMobile    string      `json:"customer_mobile,omitempty"`
	TaxableAmount     float64     `json:"taxable_amount"`
	CGSTAmount        float64     `json:"cgst_amount"`
	SGSTAmount        float64     `json:"sgst_amount"`
	TotalAmount       float64     `json:"total_amount"`
	Status            Status      `json:"status"`
	PaymentMode       *PaymentMode `json:"payment_mode,omitempty"`
	RazorpayOrderID   string      `json:"razorpay_order_id,omitempty"`
	RazorpayPaymentID string      `json:"razorpay_payment_id,omitempty"`
	WhatsappSent      bool        `json:"whatsapp_sent"`
	ItemCount         int         `json:"item_count"`
	Items             []BillItem  `json:"items,omitempty"`
	ApprovedAt        *time.Time  `json:"approved_at,omitempty"`
	CreatedAt         time.Time   `json:"created_at"`
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
	CustomerName   string                  `json:"customer_name"`
	CustomerMobile string                  `json:"customer_mobile"`
	Items          []CreateBillItemRequest `json:"items" binding:"required,min=1,dive"`
}

type ApproveBillRequest struct {
	PaymentMode PaymentMode `json:"payment_mode" binding:"required,oneof=cash upi"`
}

type UpdateBillRequest struct {
	PaymentMode PaymentMode `json:"payment_mode" binding:"required,oneof=cash upi"`
}

// DashboardHome matches the mobile "Home" screen for a single sales agent.
type DashboardHome struct {
	TodaysBills int     `json:"todays_bills"`
	TodaysSales float64 `json:"todays_sales"`
	RecentBills []Bill  `json:"recent_bills"`
}
