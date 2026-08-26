package report

import "time"

// Filter matches the Report Management screen: a date range, plus either
// "All entries" or "By sales person", optionally narrowed to one bill_no.
type Filter struct {
	From    time.Time
	To      time.Time
	StaffID *int
	BillNo  *string
}

// BillRow is one row of the "Bill-wise" report tab. ItemName, HSNCode, and
// ItemCount are fixed placeholder values (this report summarizes one row
// per bill, not per line item, and the shop only sells one HSN category of
// goods) rather than being read off the bill's actual line items.
type BillRow struct {
	Date           string  `json:"date"`
	BillNo         string  `json:"bill_no"`
	CustomerName   string  `json:"customer_name"`
	ItemName       string  `json:"item_particulars"`
	HSNCode        string  `json:"hsn_code"`
	ItemCount      int     `json:"item_count"`
	DiscountAmount float64 `json:"discount_amount"`
	TaxableAmount  float64 `json:"taxable_amount"`
	CGSTAmount     float64 `json:"cgst_amount"`
	SGSTAmount     float64 `json:"sgst_amount"`
	TotalAmount    float64 `json:"total_amount"`
	TotalCash      float64 `json:"total_cash"`
	TotalUPI       float64 `json:"total_upi"`
	Status         string  `json:"status"`
	PaymentMode    string  `json:"payment_mode"`
}

const (
	billRowItemName  = "Mixed Fire Works"
	billRowHSNCode   = "36040000"
	billRowItemCount = 1
)

// ProductRow is one row of the "Product-wise" report tab — one row per
// billed item, matching the design's "one row per item" table exactly.
type ProductRow struct {
	Date          string  `json:"date"`
	BillNo        string  `json:"bill_no"`
	CustomerName  string  `json:"customer_name"`
	ItemName      string  `json:"item_particulars"`
	HSNCode       string  `json:"hsn_code"`
	Qty           int     `json:"qty"`
	Rate          float64 `json:"rate"`
	TaxableAmount float64 `json:"taxable_amt"`
	CGSTPercent   float64 `json:"cgst_pct"`
	CGSTAmount    float64 `json:"cgst_amt"`
	SGSTPercent   float64 `json:"sgst_pct"`
	SGSTAmount    float64 `json:"sgst_amt"`
	TotalAmount   float64 `json:"total_amt"`
}
