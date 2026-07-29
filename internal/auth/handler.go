package auth

import (
	"errors"
	"net/http"

	"salestrack/internal/staff"
	"salestrack/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
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
	member, token, err := h.service.LoginSalesStaff(c.Request.Context(), req)
	if err != nil {
		status := http.StatusUnauthorized
		msg := "mobile number isincorrect"
		if errors.Is(err, staff.ErrInactive) {
			msg = err.Error()
		}
		response.Fail(c, status, msg)
		return
	}
	response.OK(c, http.StatusOK, "logged in", LoginResponse{Token: token, Role: RoleSalesStaff, User: member})
}
