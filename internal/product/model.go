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
