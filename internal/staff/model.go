package staff

import "time"

// SalesStaff is a mobile-app sales agent, created by the shop admin from the
// "User Management" screen. Login is by mobile number + the 4-digit code the
// admin generated for them. The code is stored and returned in plaintext
// (not hashed) since the admin UI needs to display and edit it at any time,
// not just once at creation.
type SalesStaff struct {
	ID           int       `json:"id"`
	AdminID      int       `json:"-"`
	Name         string    `json:"name"`
	MobileNumber string    `json:"mobile_number"`
	ShopNumber   string    `json:"shop_number"`
	LoginCode    string    `json:"login_code"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
}

// ShopNumber identifies which physical shop a staff member sells for. It
// drives the bill number prefix (see billing.BillPrefixForShop) — each shop
// runs its own bill sequence.
type CreateStaffRequest struct {
	Name         string `json:"name" binding:"required"`
	MobileNumber string `json:"mobile_number" binding:"required,len=10"`
	ShopNumber   string `json:"shop_number" binding:"required,oneof=SHOP-AKR SHOP-14-15"`
	LoginCode    string `json:"login_code" binding:"required,len=4"`
}

type UpdateStaffRequest struct {
	Name         *string `json:"name"`
	MobileNumber *string `json:"mobile_number"`
	ShopNumber   *string `json:"shop_number" binding:"omitempty,oneof=SHOP-AKR SHOP-14-15"`
	LoginCode    *string `json:"login_code" binding:"omitempty,len=4"`
}

type ResetCodeResponse struct {
	LoginCode string `json:"login_code"`
}
