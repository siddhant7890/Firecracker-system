package auth

import (
	"errors"
	"net/http"
	"time"

	"salestrack/config"
	"salestrack/internal/staff"
	"salestrack/pkg/geo"
	"salestrack/pkg/response"

	"github.com/gin-gonic/gin"
)

// istZone is a fixed +5:30 offset rather than time.LoadLocation("Asia/Kolkata")
// so this doesn't depend on the OS having tzdata installed. IST has no DST,
// so the fixed offset is always correct.
var istZone = time.FixedZone("IST", 5*60*60+30*60)

const (
	salesLoginWindowStartHour = 10 // 10:00 IST
	salesLoginWindowEndHour   = 19 // 19:00 (7pm) IST
)

type Handler struct {
	service         *Service
	factoryLocation config.FactoryLocation
}

func NewHandler(service *Service, factoryLocation config.FactoryLocation) *Handler {
	return &Handler{service: service, factoryLocation: factoryLocation}
}

// RegisterAdminRoutes wires the admin (shop owner) auth endpoints.
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.POST("/register", h.registerAdmin)
	rg.POST("/login", h.loginAdmin)
}

// RegisterSalesRoutes wires the mobile-app sales-agent auth endpoint.
func (h *Handler) RegisterSalesRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.loginSales)
}

// RegisterCashRoutes wires the cash-counter agent auth endpoint. Separate
// URL from the sales-agent login, but otherwise identical behavior/authority
// — see auth.Service.LoginCashAgent.
func (h *Handler) RegisterCashRoutes(rg *gin.RouterGroup) {
	rg.POST("/login", h.loginCash)
}

// registerAdmin bootstraps a new shop + its first admin login. There's no
// screen for this in the design (only staff logins are admin-created), so
// it's meant to be called once, directly, when setting up a new shop.
func (h *Handler) registerAdmin(c *gin.Context) {
	var req RegisterAdminRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	admin, err := h.service.RegisterAdmin(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not register shop: "+err.Error())
		return
	}
	response.OK(c, http.StatusCreated, "shop registered", admin)
}

func (h *Handler) loginAdmin(c *gin.Context) {
	var req AdminLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	admin, token, err := h.service.LoginAdmin(c.Request.Context(), req)
	if err != nil {
		response.Fail(c, http.StatusUnauthorized, err.Error())
		return
	}
	response.OK(c, http.StatusOK, "logged in", LoginResponse{Token: token, Role: RoleAdmin, User: admin})
}

func (h *Handler) loginSales(c *gin.Context) {
	var req SalesLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	// if !withinSalesLoginWindow(time.Now()) {
	// 	response.Fail(c, http.StatusForbidden, "sales login is only allowed between 10 AM and 7 PM IST")
	// 	return
	// }

	if err := geo.CheckWithinRadius(req.Latitude, req.Longitude,
		h.factoryLocation.Latitude, h.factoryLocation.Longitude,
		h.factoryLocation.RadiusMeters); err != nil {
		response.Fail(c, http.StatusForbidden, "you must be at the factory to log in")
		return
	}

	member, token, err := h.service.LoginSalesStaff(c.Request.Context(), req)
	if err != nil {
		status := http.StatusUnauthorized
		msg := "mobile number is incorrect"
		if errors.Is(err, staff.ErrInactive) {
			msg = err.Error()
		}
		response.Fail(c, status, msg)
		return
	}
	response.OK(c, http.StatusOK, "logged in", LoginResponse{Token: token, Role: RoleSalesStaff, User: member})
}

// loginCash is the cash-counter agent's own login endpoint — same mobile +
// 4-digit code + factory-radius check as loginSales, but only accounts
// created with role "cash_agent" are accepted (see auth.Service.LoginCashAgent).
// Resulting token/role/authority are identical to a sales-agent login.
func (h *Handler) loginCash(c *gin.Context) {
	var req SalesLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	if err := geo.CheckWithinRadius(req.Latitude, req.Longitude,
		h.factoryLocation.Latitude, h.factoryLocation.Longitude,
		h.factoryLocation.RadiusMeters); err != nil {
		response.Fail(c, http.StatusForbidden, "you must be at the factory to log in")
		return
	}

	member, token, err := h.service.LoginCashAgent(c.Request.Context(), req)
	if err != nil {
		status := http.StatusUnauthorized
		msg := "mobile number is incorrect"
		if errors.Is(err, staff.ErrInactive) {
			msg = err.Error()
		}
		response.Fail(c, status, msg)
		return
	}
	response.OK(c, http.StatusOK, "logged in", LoginResponse{Token: token, Role: RoleSalesStaff, User: member})
}

// withinSalesLoginWindow reports whether t, in IST, falls within
// [salesLoginWindowStartHour, salesLoginWindowEndHour).
func withinSalesLoginWindow(t time.Time) bool {
	hour := t.In(istZone).Hour()
	return hour >= salesLoginWindowStartHour && hour < salesLoginWindowEndHour
}
