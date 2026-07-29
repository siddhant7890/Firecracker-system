package product

import "time"

// Product mirrors the "Product Management" screen: category, HSN, GST%,
// and a final price that's computed server-side from taxable value + GST.
type Product struct {
	ID           int       `json:"id"`
	AdminID      int       `json:"-"`
	ItemCode     string    `json:"item_code"`
	Name         string    `json:"name"`
	Category     string    `json:"category"`
	HSNCode      string    `json:"hsn_code"`
	Unit         string    `json:"unit"`
	TaxableValue float64   `json:"taxable_value"`
	GSTPercent   float64   `json:"gst_percent"`
	FinalPrice   float64   `json:"final_price"`
	ImageURL     string    `json:"image_url,omitempty"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateProductRequest struct {
	ItemCode     string  `json:"item_code"`
	Name         string  `json:"name" binding:"required"`
	Category     string  `json:"category" binding:"required"`
	HSNCode      string  `json:"hsn_code"`
	Unit         string  `json:"unit"`
	TaxableValue float64 `json:"taxable_value" binding:"required,gt=0"`
	GSTPercent   float64 `json:"gst_percent" binding:"required,gte=0"`
	ImageURL     string  `json:"image_url"`
}

type UpdateProductRequest struct {
	Name         *string  `json:"name"`
	Category     *string  `json:"category"`
	HSNCode      *string  `json:"hsn_code"`
	Unit         *string  `json:"unit"`
	TaxableValue *float64 `json:"taxable_value"`
	GSTPercent   *float64 `json:"gst_percent"`
	ImageURL     *string  `json:"image_url"`
}

// BulkUploadRow is one validated, ready-to-persist row from a bulk upload
// spreadsheet. TaxableValue is back-computed from the sheet's Price (which
// is GST-inclusive) and GSTPercent, since the sheet doesn't carry it directly.
type BulkUploadRow struct {
	ItemCode     string
	Name         string
	Category     string
	HSNCode      string
	Unit         string
	GSTPercent   float64
	TaxableValue float64
}

// BulkUploadRowError explains why a spreadsheet row was skipped. Row is
// 1-indexed against the spreadsheet, including the header row (so the first
// data row is Row 2), matching what a user sees when they open the file.
type BulkUploadRowError struct {
	Row    int    `json:"row"`
	Reason string `json:"reason"`
}

// BulkUploadResult summarizes a bulk upload: existing item codes were
// updated, new ones were inserted, invalid rows were skipped (not failed —
// the rest of the file still processes).
type BulkUploadResult struct {
	Inserted int                  `json:"inserted"`
	Updated  int                  `json:"updated"`
	Skipped  []BulkUploadRowError `json:"skipped,omitempty"`
}
