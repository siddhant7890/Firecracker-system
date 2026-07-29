package report

import (
	"net/http"
	"strconv"
	"time"

	"salestrack/internal/middleware"
	"salestrack/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// Register wires "Report Management": Bill-wise / Product-wise tabs, date
// range + sales-person filters, and Excel download.
func (h *Handler) Register(rg *gin.RouterGroup) {
	rg.GET("/bills", h.billWise)
	rg.GET("/products", h.productWise)
	rg.GET("/download", h.download)
}

func (h *Handler) billWise(c *gin.Context) {
	claims := middleware.FromContext(c)
	f := parseFilter(c)
	out, err := h.service.BillWise(c.Request.Context(), claims.AdminID, f)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load report")
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

func (h *Handler) productWise(c *gin.Context) {
	claims := middleware.FromContext(c)
	f := parseFilter(c)
	out, err := h.service.ProductWise(c.Request.Context(), claims.AdminID, f)
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not load report")
		return
	}
	response.OK(c, http.StatusOK, "", out)
}

func (h *Handler) download(c *gin.Context) {
	claims := middleware.FromContext(c)
	f := parseFilter(c)
	reportType := c.DefaultQuery("type", "bill")

	var xlsx *excelize.File
	var err error
	var filename string

	if reportType == "product" {
		xlsx, err = h.service.BuildProductWiseExcel(c.Request.Context(), claims.AdminID, f)
		filename = FileName("product-wise", f)
	} else {
		xlsx, err = h.service.BuildBillWiseExcel(c.Request.Context(), claims.AdminID, f)
		filename = FileName("bill-wise", f)
	}
	if err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not build report")
		return
	}
	defer xlsx.Close()

	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	if err := xlsx.Write(c.Writer); err != nil {
		response.Fail(c, http.StatusInternalServerError, "could not stream report")
	}
}

func parseFilter(c *gin.Context) Filter {
	loc := time.Now().Location()
	from := parseDate(c.Query("from"), loc)
	to := parseDate(c.Query("to"), loc)

	now := time.Now()
	if from.IsZero() {
		from = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}
	if to.IsZero() {
		to = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}
	// "to" is inclusive of the whole day in the UI, so push it to the start
	// of the next day for a half-open range in the query.
	to = to.AddDate(0, 0, 1)

	f := Filter{From: from, To: to}
	if s := c.Query("staff_id"); s != "" {
		if id, err := strconv.Atoi(s); err == nil {
			f.StaffID = &id
		}
	}
	return f
}

func parseDate(s string, loc *time.Location) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.ParseInLocation("2006-01-02", s, loc)
	if err != nil {
		return time.Time{}
	}
	return t
}
