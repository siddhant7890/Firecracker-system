// Package authctx holds the request-scoped auth types shared between the
// middleware (which sets them) and feature handlers (which read them).
// It exists so packages like staff can read the caller's claims without
// importing internal/middleware, which itself depends on internal/auth,
// which depends on internal/staff — pulling Claims out here breaks that cycle.
package authctx

import (
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const (
	RoleAdmin      = "admin"
	RoleSalesStaff = "sales_staff"
)

const ClaimsKey = "claims"

// Claims identifies who is calling the API: an admin (shop owner) or a
// sales_staff member (mobile app agent). AdminID scopes every query to the
// correct shop, whichever role is calling.
type Claims struct {
	UserID  int    `json:"user_id"`
	AdminID int    `json:"admin_id"`
	Role    string `json:"role"`
	Name    string `json:"name"`
	jwt.RegisteredClaims
}

// FromContext pulls the parsed claims set by middleware.RequireAuth.
func FromContext(c *gin.Context) *Claims {
	v, ok := c.Get(ClaimsKey)
	if !ok {
		return nil
	}
	claims, ok := v.(*Claims)
	if !ok {
		return nil
	}
	return claims
}
