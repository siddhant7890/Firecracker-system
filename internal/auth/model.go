package auth

import "time"

// Admin is a shop owner using the web admin portal (e.g. Sunil Milwani /
// Shama Fireworks). One Admin row == one shop/tenant; every other table is
// scoped by admin_id.
type Admin struct {
	ID           int       `json:"id"`
	Name         string    `json:"name"`
	MobileNumber string    `json:"mobile_number"`
	Email        string    `json:"email,omitempty"`
	ShopName     string    `json:"shop_name"`
	BillPrefix   string    `json:"bill_prefix"`
	CreatedAt    time.Time `json:"created_at"`
}

type RegisterAdminRequest struct {
	Name         string `json:"name" binding:"required"`
	MobileNumber string `json:"mobile_number" binding:"required,len=10"`
	Email        string `json:"email"`
	Password     string `json:"password" binding:"required,min=6"`
	ShopName     string `json:"shop_name" binding:"required"`
	BillPrefix   string `json:"bill_prefix"`
}

type AdminLoginRequest struct {
	MobileNumber string `json:"mobile_number" binding:"required"`
	Password     string `json:"password" binding:"required"`
}

// SalesLoginRequest matches the mobile login screen: mobile number + the
// 4-digit code the admin generated for this agent.
type SalesLoginRequest struct {
	MobileNumber string  `json:"mobile_number" binding:"required"`
	LoginCode    string  `json:"login_code" binding:"required,len=4"`
	Latitude     float64 `json:"latitude" binding:"required"`
	Longitude    float64 `json:"longitude" binding:"required"`
}

type LoginResponse struct {
	Token string `json:"token"`
	Role  string `json:"role"`
	User  any    `json:"user"`
}
