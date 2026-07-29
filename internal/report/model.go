package report

import "time"

// Filter matches the Report Management screen: a date range, plus either
// "All entries" or "By sales person".
type Filter struct {
	From    time.Time
	To      time.Time
	StaffID *int
}

// BillRow is one row of the "Bill-wise" report tab.
type BillRow struct {
	Date          string  `json:"date"`
	BillNo        string  `json:"bill_no"`
	CustomerName  string  `json:"customer_name"`
	SalesPerson   string  `json:"sales_person"`
	ItemCount     int     `json:"item_count"`
	TaxableAmount float64 `json:"taxable_amount"`
	CGSTAmount    float64 `json:"cgst_amount"`
	SGSTAmount    float64 `json:"sgst_amount"`
	TotalAmount   float64 `json:"total_amount"`
	Status        string  `json:"status"`
	PaymentMode   string  `json:"payment_mode"`
}

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
