package staff

import (
	"errors"
	"net/http"
	"strconv"

	"salestrack/internal/authctx"
	"salestrack/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register wires the admin-only "User Management" routes.
func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.POST("", h.create)
	rg.GET("", h.list)
	rg.GET("/:id", h.get)
	rg.PUT("/:id", h.update)
	rg.PATCH("/:id/status", h.setActive)
	rg.POST("/:id/reset-code", h.resetCode)
	rg.DELETE("/:id", h.delete)
}

func (h *Handler) create(c *gin.Context) {
	claims := authctx.FromContext(c)
	var req CreateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Create(c.Request.Context(), claims.AdminID, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not create sales staff: "+err.Error())
		return
	}
	response.OK(c, http.StatusCreated, "sales staff created", out)
}

func (h *Handler) list(c *gin.Context) {
	claims := authctx.FromContext(c)
	out, err := h.service.List(c.Request.Context(), claims.AdminID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load sales staff")
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

func (h *Handler) get(c *gin.Context) {
	claims := authctx.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	out, err := h.service.Get(c.Request.Context(), claims.AdminID, id)
	if err != nil {
		respondNotFound(c, err)
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

func (h *Handler) update(c *gin.Context) {
	claims := authctx.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req UpdateStaffRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	out, err := h.service.Update(c.Request.Context(), claims.AdminID, id, req)
	if err != nil {
		respondNotFound(c, err)
		return
	}
	response.OK(c, http.StatusOK, "sales staff updated", out)
}

func (h *Handler) setActive(c *gin.Context) {
	claims := authctx.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var body struct {
		IsActive bool `json:"is_active"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	if err := h.service.SetActive(c.Request.Context(), claims.AdminID, id, body.IsActive); err != nil {
		respondNotFound(c, err)
		return
	}
	response.OK(c, http.StatusOK, "status updated", nil)
}

func (h *Handler) resetCode(c *gin.Context) {
	claims := authctx.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	out, err := h.service.ResetCode(c.Request.Context(), claims.AdminID, id)
	if err != nil {
		respondNotFound(c, err)
		return
	}
	response.OK(c, http.StatusOK, "new login code generated", out)
}

func (h *Handler) delete(c *gin.Context) {
	claims := authctx.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), claims.AdminID, id); err != nil {
		respondNotFound(c, err)
		return
	}
	response.OK(c, http.StatusOK, "sales staff removed", nil)
}

func respondNotFound(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		response.Fail(c, http.StatusNotFound, "sales staff not found")
		return
	}
	response.Fail(c, http.StatusInternalServerError, err.Error())
}
