package sales

import (
	"errors"
	"net/http"
	"strconv"

	"salestrack/internal/billing"
	"salestrack/internal/middleware"
	"salestrack/internal/whatsapp"
	"salestrack/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler wires the mobile-app "sales agent" endpoints: New Bill, Bill
// History, Home, Send to WhatsApp. Product listing lives in the product
// package (RegisterSalesRoutes), mounted separately in main.go.
type Handler struct {
	billing *billing.Service
}

func NewHandler(billingSvc *billing.Service) *Handler {
	return &Handler{billing: billingSvc}
}

func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.GET("/home", h.home)
	rg.POST("/bills", h.createBill)
	rg.GET("/bills", h.listBills)
	rg.GET("/bills/:id", h.getBill)
	rg.PATCH("/bills/:id", h.updateBill)
	rg.POST("/bills/:id/whatsapp", h.sendWhatsapp)
}

// home matches the "Namaste, <name> — Today's Bills / Today's Sales /
// Recent Bills" screen.
func (h *Handler) home(c *gin.Context) {
	claims := middleware.FromContext(c)
	out, err := h.billing.SalesHome(c.Request.Context(), claims.AdminID, claims.UserID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load dashboard")
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

// createBill is the "New Bill" -> "View Cart" -> submit flow. The response
// doubles as the "Confirmation" screen payload.
func (h *Handler) createBill(c *gin.Context) {
	claims := middleware.FromContext(c)
	var req billing.CreateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	bill, err := h.billing.CreateBill(c.Request.Context(), claims.AdminID, claims.UserID, req)
	if err != nil {
		if errors.Is(err, billing.ErrEmptyCart) {
			response.Fail(c, http.StatusBadRequest, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, "could not create bill: "+err.Error())
		return
	}
	response.OK(c, http.StatusCreated, "bill submitted for admin approval", bill)
}

// listBills matches Bill History's Today / This Week / All tabs, via
// ?range=today|week|all (defaults to today).
func (h *Handler) listBills(c *gin.Context) {
	claims := middleware.FromContext(c)
	out, err := h.billing.ListForStaff(c.Request.Context(), claims.AdminID, claims.UserID, c.Query("range"))
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load bill history")
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

func (h *Handler) getBill(c *gin.Context) {
	claims := middleware.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	bill, err := h.billing.Get(c.Request.Context(), claims.AdminID, id)
	if err != nil || bill.SalesStaffID != claims.UserID {
		response.Fail(c, http.StatusNotFound, "bill not found")
		return
	}
	response.OK(c, http.StatusOK, "", bill)
}

// updateBill lets a sales agent correct their own bill's customer/header
// details (name, mobile, token, carton count, GST no., discount) and/or
// payment mode. Every field is optional — only what's sent gets changed.
func (h *Handler) updateBill(c *gin.Context) {
	claims := middleware.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	existing, err := h.billing.Get(c.Request.Context(), claims.AdminID, id)
	if err != nil || existing.SalesStaffID != claims.UserID {
		response.Fail(c, http.StatusNotFound, "bill not found")
		return
	}
	var req billing.UpdateBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	bill, err := h.billing.UpdateBill(c.Request.Context(), claims.AdminID, id, req)
	if err != nil {
		respondBillErr(c, err)
		return
	}
	response.OK(c, http.StatusOK, "bill updated", bill)
}

// sendWhatsapp builds a wa.me deep link with the bill summary and marks the
// bill as sent, matching the "Sent on WhatsApp" tag in Bill History.
func (h *Handler) sendWhatsapp(c *gin.Context) {
	claims := middleware.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	bill, err := h.billing.Get(c.Request.Context(), claims.AdminID, id)
	if err != nil || bill.SalesStaffID != claims.UserID {
		response.Fail(c, http.StatusNotFound, "bill not found")
		return
	}
	if err := h.billing.MarkWhatsappSent(c.Request.Context(), claims.AdminID, id); err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not update bill")
		return
	}
	response.OK(c, http.StatusOK, "ready to send on WhatsApp", WhatsAppLinkResponse{Link: whatsapp.BuildLink(bill)})
}

func respondBillErr(c *gin.Context, err error) {
	switch {
	case errors.Is(err, billing.ErrNotFound):
		response.Fail(c, http.StatusNotFound, "bill not found")
	case errors.Is(err, billing.ErrAlreadyHandled):
		response.Fail(c, http.StatusConflict, err.Error())
	default:
		response.Fail(c, http.StatusInternalServerError, err.Error())
	}
}
