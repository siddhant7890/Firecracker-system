package product

import (
	"errors"
	"net/http"
	"strconv"

	"salestrack/internal/middleware"
	"salestrack/pkg/response"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterAdminRoutes wires the full CRUD used by the admin "Product
// Management" screen.
func (h *Handler) RegisterAdminRoutes(rg *gin.RouterGroup) {
	rg.POST("", h.create)
	rg.POST("/bulk-upload", h.bulkUpload)
	rg.GET("", h.listAll)
	rg.GET("/:id", h.get)
	rg.PUT("/:id", h.update)
	rg.PATCH("/:id/status", h.setActive)
	rg.DELETE("/:id", h.delete)
	rg.DELETE("", h.deleteAll)
}

// RegisterSalesRoutes wires the read-only product picker used by the "New
// Bill" screen (search + category filter), plus the full listing (including
// inactive products) for sales agents that also need admin-parity visibility.
func (h *Handler) RegisterSalesRoutes(rg *gin.RouterGroup) {
	rg.GET("", h.listActive)
	rg.GET("/all", h.listAll)
}

func (h *Handler) create(c *gin.Context) {
	claims := middleware.FromContext(c)
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	p, err := h.service.Create(c.Request.Context(), claims.AdminID, req)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not create product: "+err.Error())
		return
	}
	response.OK(c, http.StatusCreated, "product added", p)
}


// bulkUpload handles the "Bulk Upload" spreadsheet: item_code, name,
// category, hsn_code, unit, gst_percent, price. Existing item codes are
// updated, new ones are inserted; taxable_value is derived server-side from
// price and gst_percent since the sheet only carries the GST-inclusive price.
func (h *Handler) bulkUpload(c *gin.Context) {
	claims := middleware.FromContext(c)
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "file is required")
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "could not read file")
		return
	}
	defer f.Close()

	result, err := h.service.BulkUpload(c.Request.Context(), claims.AdminID, f)
	if err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	response.OK(c, http.StatusOK, "bulk upload complete", result)
}

func (h *Handler) listAll(c *gin.Context) {
	claims := middleware.FromContext(c)
	start, limit := parsePaging(c)
	out, err := h.service.List(c.Request.Context(), claims.AdminID, false, c.Query("category"), c.Query("search"), start, limit)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load products")
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

func (h *Handler) listActive(c *gin.Context) {
	claims := middleware.FromContext(c)
	start, limit := parsePaging(c)
	out, err := h.service.List(c.Request.Context(), claims.AdminID, true, c.Query("category"), c.Query("search"), start, limit)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load products")
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

// parsePaging reads "start" and "limit" query params. start defaults to 0;
// limit defaults to 20 and is only applied when it parses to a positive int.
func parsePaging(c *gin.Context) (start, limit int) {
	start, _ = strconv.Atoi(c.Query("start"))
	if start < 0 {
		start = 0
	}
	limit = 20
	if l, err := strconv.Atoi(c.Query("limit")); err == nil && l > 0 {
		limit = l
	}
	return start, limit
}

func (h *Handler) get(c *gin.Context) {
	claims := middleware.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	p, err := h.service.Get(c.Request.Context(), claims.AdminID, id)
	if err != nil {
		respondNotFound(c, err)
		return
	}
	response.OK(c, http.StatusOK, "", p)
}

func (h *Handler) update(c *gin.Context) {
	claims := middleware.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, http.StatusBadRequest, err.Error())
		return
	}
	p, err := h.service.Update(c.Request.Context(), claims.AdminID, id, req)
	if err != nil {
		respondNotFound(c, err)
		return
	}
	response.OK(c, http.StatusOK, "product updated", p)
}

func (h *Handler) setActive(c *gin.Context) {
	claims := middleware.FromContext(c)
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

func (h *Handler) delete(c *gin.Context) {
	claims := middleware.FromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		response.Fail(c, http.StatusBadRequest, "invalid id")
		return
	}
	if err := h.service.Delete(c.Request.Context(), claims.AdminID, id); err != nil {
		respondNotFound(c, err)
		return
	}
	response.OK(c, http.StatusOK, "product removed", nil)
}

// deleteAll soft-deletes every product for this admin — a "start fresh"
// reset of the product catalog.
func (h *Handler) deleteAll(c *gin.Context) {
	claims := middleware.FromContext(c)
	count, err := h.service.DeleteAll(c.Request.Context(), claims.AdminID)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not delete products")
		return
	}
	response.OK(c, http.StatusOK, "products removed", gin.H{"deleted_count": count})
}

func respondNotFound(c *gin.Context, err error) {
	if errors.Is(err, ErrNotFound) {
		response.Fail(c, http.StatusNotFound, "product not found")
		return
	}
	response.Fail(c, http.StatusInternalServerError, err.Error())
}
