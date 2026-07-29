package auth

import (
	"errors"
	"time"

	"salestrack/config"
	"salestrack/internal/authctx"

	"github.com/golang-jwt/jwt/v5"
)

const (
	RoleAdmin      = authctx.RoleAdmin
	RoleSalesStaff = authctx.RoleSalesStaff
)

// Claims is an alias for authctx.Claims so existing callers can keep
// referring to auth.Claims.
type Claims = authctx.Claims

func GenerateToken(userID, adminID int, role, name string) (string, error) {
	claims := Claims{
		UserID:  userID,
		AdminID: adminID,
		Role:    role,
		Name:    name,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(30 * 24 * time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.App.JWTSecret))
}

func ParseToken(tokenStr string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(config.App.JWTSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}
	return claims, nil
}
