package report

import (
	"context"
	"fmt"

	"github.com/xuri/excelize/v2"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) BillWise(ctx context.Context, adminID int, f Filter, start, limit int) ([]BillRow, error) {
	return s.repo.BillWise(ctx, adminID, f, start, limit)
}

func (s *Service) ProductWise(ctx context.Context, adminID int, f Filter, start, limit int) ([]ProductRow, error) {
	return s.repo.ProductWise(ctx, adminID, f, start, limit)
}

// BuildBillWiseExcel powers the "Download Excel" button on the Bill-wise tab.
func (s *Service) BuildBillWiseExcel(ctx context.Context, adminID int, f Filter) (*excelize.File, error) {
	rows, err := s.BillWise(ctx, adminID, f, 0, 0)
	if err != nil {
		return nil, err
	}

	f2 := excelize.NewFile()
	sheet := "Bill-wise"
	f2.SetSheetName("Sheet1", sheet)
	headers := []string{"Date", "Bill No", "Customer Name", "Item / Particulars", "HSN Code", "Items",
		"Discount Amt", "Taxable Amt", "CGST Amt", "SGST Amt", "Total Amt", "Total Cash", "Total UPI", "Status", "Payment Mode"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f2.SetCellValue(sheet, cell, h)
	}
	for r, row := range rows {
		rowNum := r + 2
		values := []any{row.Date, row.BillNo, row.CustomerName, row.ItemName, row.HSNCode, row.ItemCount,
			row.DiscountAmount, row.TaxableAmount, row.CGSTAmount, row.SGSTAmount, row.TotalAmount,
			row.TotalCash, row.TotalUPI, row.Status, row.PaymentMode}
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowNum)
			f2.SetCellValue(sheet, cell, v)
		}
	}
	return f2, nil
}

// BuildProductWiseExcel powers "Download Excel" on the Product-wise tab —
// one row per billed item, matching the on-screen table exactly.
func (s *Service) BuildProductWiseExcel(ctx context.Context, adminID int, f Filter) (*excelize.File, error) {
	rows, err := s.ProductWise(ctx, adminID, f, 0, 0)
	if err != nil {
		return nil, err
	}

	f2 := excelize.NewFile()
	sheet := "Product-wise"
	f2.SetSheetName("Sheet1", sheet)
	headers := []string{"Date", "Bill No", "Customer Name", "Item / Particulars", "HSN Code", "Qty", "Rate",
		"Taxable Amt", "CGST %", "CGST Amt", "SGST %", "SGST Amt", "Total Amt"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f2.SetCellValue(sheet, cell, h)
	}
	for r, row := range rows {
		rowNum := r + 2
		values := []any{row.Date, row.BillNo, row.CustomerName, row.ItemName, row.HSNCode, row.Qty, row.Rate,
			row.TaxableAmount, row.CGSTPercent, row.CGSTAmount, row.SGSTPercent, row.SGSTAmount, row.TotalAmount}
		for i, v := range values {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowNum)
			f2.SetCellValue(sheet, cell, v)
		}
	}
	return f2, nil
}

func FileName(reportType string, f Filter) string {
	return fmt.Sprintf("%s-report_%s_to_%s.xlsx", reportType, f.From.Format("2006-01-02"), f.To.Format("2006-01-02"))
}
