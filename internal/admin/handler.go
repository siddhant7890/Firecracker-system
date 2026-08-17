package admin

import (
	"errors"
	"net/http"
	"strconv"

	"salestrack/internal/billing"
	"salestrack/internal/middleware"
	"salestrack/internal/payment"
	"salestrack/pkg/response"

	"github.com/gin-gonic/gin"
)

// Handler wires the admin webapp's Dashboard and Cash Counter screens.
// Product Management and User Management live in their own packages
// (product.Handler, staff.Handler), mounted separately in main.go.
type Handler struct {
	billing *billing.Service
	payment *payment.Service
}

func NewHandler(billingSvc *billing.Service, paymentSvc *payment.Service) *Handler {
	return &Handler{billing: billingSvc, payment: paymentSvc}
}

func (h *Handler) RegisterDashboardRoutes(rg *gin.RouterGroup, service *Service) {
	rg.GET("/dashboard", func(c *gin.Context) {
		claims := middleware.FromContext(c)
		out, err := service.Dashboard(c.Request.Context(), claims.AdminID)
		if err != nil {
			response.Fail(c, http.StatusInternalServerError, "could not load dashboard")
			return
		}
		response.OK(c, http.StatusOK, "", out)
	})
}

// RegisterBillRoutes wires "Cash Counter — bill approvals".
func (h *Handler) RegisterBillRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.listBills)
	rg.GET("/pending", h.pending)
	rg.GET("/approved", h.approvedToday)
	rg.GET("/:id", h.getBill)
	rg.PATCH("/:id", h.updateBill)

	rg.POST("/:id/approve", h.approve)
	rg.POST("/:id/reject", h.reject)
	rg.POST("/:id/verify-payment", h.verifyPayment)
}

func (h *Handler) listBills(c *gin.Context) {
	claims := middleware.FromContext(c)
	f := parseListFilter(c)
	out, err := h.billing.ListForAdmin(c.Request.Context(), claims.AdminID, f)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load bills")
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

func (h *Handler) pending(c *gin.Context) {
	claims := middleware.FromContext(c)
	status := billing.StatusPending
	out, err := h.billing.ListForAdmin(c.Request.Context(), claims.AdminID, billing.ListFilter{Status: &status})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load pending approvals")
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

func (h *Handler) approvedToday(c *gin.Context) {
	claims := middleware.FromContext(c)
	status := billing.StatusApproved
	out, err := h.billing.ListForAdmin(c.Request.Context(), claims.AdminID, billing.ListFilter{Status: &status})
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load approved bills")
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
	if err != nil {
		respondBillErr(c, err)
		return
	}
	response.OK(c, http.StatusOK, "", bill)
}

// updateBill lets the admin correct a bill's payment_mode after the fact
// (cash vs UPI), independent of the approve/reject flow.
func (h *Handler) updateBill(c *gin.Context) {
	claims := middleware.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
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

// approve is the "Cash" / "UPI" tap on Cash Counter. It approves the bill
// immediately, recording whichever payment_mode was sent in the request —
// no Razorpay order/checkout involved.
func (h *Handler) approve(c *gin.Context) {
	claims := middleware.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req billing.ApproveBillRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	bill, err := h.billing.Approve(c.Request.Context(), claims.AdminID, id, claims.UserID, claims.Role, req.PaymentMode)
	if err != nil {
		respondBillErr(c, err)
		return
	}
	response.OK(c, http.StatusOK, "bill approved ("+string(req.PaymentMode)+")", bill)
}

func (h *Handler) reject(c *gin.Context) {
	claims := middleware.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	bill, err := h.billing.Reject(c.Request.Context(), claims.AdminID, id, claims.UserID, claims.Role)
	if err != nil {
		respondBillErr(c, err)
		return
	}
	response.OK(c, http.StatusOK, "bill rejected", bill)
}

// verifyPayment is called by the admin webapp right after Razorpay
// Checkout.js's success handler fires (fast path; the webhook is the
// authoritative fallback if the browser closes before this call completes).
func (h *Handler) verifyPayment(c *gin.Context) {
	var req payment.VerifyPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	bill, err := h.payment.VerifyPayment(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, payment.ErrSignatureMismatch) {
			response.Fail(c, http.StatusUnauthorized, err.Error())
			return
		}
		response.Fail(c, http.StatusInternalServerError, "could not confirm payment: "+err.Error())
		return
	}
	response.OK(c, http.StatusOK, "payment confirmed, bill approved", bill)
}

func parseListFilter(c *gin.Context) billing.ListFilter {
	var f billing.ListFilter
	if s := c.Query("status"); s != "" {
		st := billing.Status(s)
		f.Status = &st
	}
	if s := c.Query("staff_id"); s != "" {
		if id, err := strconv.Atoi(s); err == nil {
			f.StaffID = &id
		}
	}
	return f
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
