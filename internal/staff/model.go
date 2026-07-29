package staff

import "time"

// SalesStaff is a mobile-app sales agent, created by the shop admin from the
// "User Management" screen. Login is by mobile number + the 4-digit code the
// admin generated for them (shown once in the admin UI, entered as a PIN on
// the mobile login screen).
type SalesStaff struct {
	ID            int       `json:"id"`
	AdminID       int       `json:"-"`
	Name          string    `json:"name"`
	MobileNumber  string    `json:"mobile_number"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
}

type CreateStaffRequest struct {
	Name         string `json:"name" binding:"required"`
	MobileNumber string `json:"mobile_number" binding:"required,len=10"`
	LoginCode    string `json:"login_code" binding:"required,len=4"`
}

// CreateStaffResponse returns the plaintext login code exactly once, at
// creation time, so the admin can share it with the agent. It is never
// retrievable again (only its hash is stored).
type CreateStaffResponse struct {
	SalesStaff
	LoginCode string `json:"login_code"`
}

type UpdateStaffRequest struct {
	Name         *string `json:"name"`
	MobileNumber *string `json:"mobile_number"`
}

type ResetCodeResponse struct {
	LoginCode string `json:"login_code"`
}
