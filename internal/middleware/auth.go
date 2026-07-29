package middleware

import (
	"net/http"
	"strings"

	"salestrack/internal/auth"
	"salestrack/internal/authctx"
	"salestrack/pkg/response"

	"github.com/gin-gonic/gin"
)

const ClaimsKey = authctx.ClaimsKey

// RequireAuth validates the Bearer JWT and stores the parsed claims on the
// context. Pass one or more allowed roles (auth.RoleAdmin / auth.RoleSalesStaff);
// omit to just require any valid token.
func RequireAuth(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			response.Fail(c, http.StatusUnauthorized, "missing bearer token")
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := auth.ParseToken(tokenStr)
		if err != nil {
			response.Fail(c, http.StatusUnauthorized, "invalid or expired token")
			c.Abort()
			return
		}

		if len(allowedRoles) > 0 {
			allowed := false
			for _, r := range allowedRoles {
				if claims.Role == r {
					allowed = true
					break
				}
			}
			if !allowed {
				response.Fail(c, http.StatusForbidden, "you do not have access to this resource")
				c.Abort()
				return
			}
		}

		c.Set(ClaimsKey, claims)
		c.Next()
	}
}

// FromContext pulls the parsed claims set by RequireAuth.
func FromContext(c *gin.Context) *auth.Claims {
	return authctx.FromContext(c)
}
