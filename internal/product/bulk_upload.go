package product

import (
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
)

// headerAliases maps a normalized (lowercased, spaces/underscores stripped)
// column header to the canonical field it fills, so the template tolerates
// "Item Code", "item_code", "Code", "GST %", etc.
var headerAliases = map[string]string{
	"itemcode":    "item_code",
	"code":        "item_code",
	"name":        "name",
	"productname": "name",
	"category":    "category",
	"hsncode":     "hsn_code",
	"hsn":         "hsn_code",
	"unit":        "unit",
	"gstpercent":  "gst_percent",
	"gst":         "gst_percent",
	"gst%":        "gst_percent",
	"price":       "price",
	"finalprice":  "price",
	"mrp":         "price",
}

var requiredBulkUploadColumns = []string{"item_code", "name", "category", "gst_percent", "price"}

func normalizeHeader(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "_", "")
	return s
}

// parseBulkUploadRows reads a product bulk-upload spreadsheet: item_code,
// name, category, hsn_code (optional), unit (optional), gst_percent, price
// (GST-inclusive). Rows that fail validation are collected as skips rather
// than aborting the whole upload — one bad row shouldn't block the rest.
func parseBulkUploadRows(r io.Reader) ([]BulkUploadRow, []BulkUploadRowError, error) {
	xf, err := excelize.OpenReader(r)
	if err != nil {
		return nil, nil, fmt.Errorf("invalid excel file: %w", err)
	}
	defer xf.Close()

	sheet := xf.GetSheetName(0)
	allRows, err := xf.GetRows(sheet)
	fmt.Println(allRows)
	if err != nil {
		return nil, nil, err
	}
	if len(allRows) < 2 {
		return nil, nil, errors.New("file has no data rows")
	}

	col := map[string]int{}
	for i, h := range allRows[0] {
		if canon, ok := headerAliases[normalizeHeader(h)]; ok {
			col[canon] = i
		}
	}
	for _, key := range requiredBulkUploadColumns {
		if _, ok := col[key]; !ok {
			return nil, nil, fmt.Errorf("missing required column: %s", key)
		}
	}

	cell := func(row []string, key string) string {
		i, ok := col[key]
		if !ok || i >= len(row) {
			return ""
		}
		return strings.TrimSpace(row[i])
	}

	var rows []BulkUploadRow
	var skipped []BulkUploadRowError
	for i, raw := range allRows[1:] {
		rowNum := i + 2 // +1 for 0-index, +1 for the header row
		if len(raw) == 0 || strings.TrimSpace(strings.Join(raw, "")) == "" {
			continue // blank row
		}

		itemCode := cell(raw, "item_code")
		name := cell(raw, "name")
		category := cell(raw, "category")
		hsn := cell(raw, "hsn_code")
		unit := cell(raw, "unit")

		if itemCode == "" {
			skipped = append(skipped, BulkUploadRowError{Row: rowNum, Reason: "item_code is required"})
			continue
		}
		if name == "" {
			skipped = append(skipped, BulkUploadRowError{Row: rowNum, Reason: "name is required"})
			continue
		}
		if category == "" {
			skipped = append(skipped, BulkUploadRowError{Row: rowNum, Reason: "category is required"})
			continue
		}

		gst, err := strconv.ParseFloat(cell(raw, "gst_percent"), 64)
		if err != nil || gst < 0 {
			skipped = append(skipped, BulkUploadRowError{Row: rowNum, Reason: "gst_percent must be a non-negative number"})
			continue
		}
		price, err := strconv.ParseFloat(cell(raw, "price"), 64)
		if err != nil || price <= 0 {
			skipped = append(skipped, BulkUploadRowError{Row: rowNum, Reason: "price must be a positive number"})
			continue
		}

		if hsn == "" {
			hsn = "3604"
		}
		if unit == "" {
			unit = "Piece"
		}

		// price is GST-inclusive; back out the taxable value the same way
		// the DB derives final_price = taxable_value * (1 + gst/100).
		taxableValue := math.Round(price/(1+gst/100)*100) / 100

		rows = append(rows, BulkUploadRow{
			ItemCode:     itemCode,
			Name:         name,
			Category:     category,
			HSNCode:      hsn,
			Unit:         unit,
			GSTPercent:   gst,
			TaxableValue: taxableValue,
		})
	}

	return rows, skipped, nil
}
