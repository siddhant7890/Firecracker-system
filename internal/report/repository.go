package report

import (
	"context"
	"fmt"
	"strings"

	"salestrack/internal/billing"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// BillWise returns bill-wise report rows. limit > 0 pages the result set;
// start is the offset applied when limit is set.
func (r *Repository) BillWise(ctx context.Context, adminID int, f Filter, start, limit int) ([]BillRow, error) {
	query := `
		SELECT to_char(b.created_at, 'DD/MM/YYYY'), b.bill_no, b.customer_name,
			b.discount_amount, b.taxable_amount, b.cgst_amount, b.sgst_amount, b.total_amount, b.status,
			COALESCE(b.payment_mode::text, ''), b.total_cash, b.total_upi
		FROM bills b
		WHERE b.admin_id = $1 AND b.created_at >= $2 AND b.created_at < $3`
	args := []any{adminID, f.From, f.To}

	if f.StaffID != nil {
		args = append(args, *f.StaffID)
		query += fmt.Sprintf(" AND b.sales_staff_id = $%d", len(args))
	}
	if f.BillNo != nil {
		args = append(args, likePrefix(*f.BillNo))
		query += fmt.Sprintf(" AND b.bill_no LIKE $%d ESCAPE '\\'", len(args))
	}
	if f.PaymentMode != nil {
		args = append(args, *f.PaymentMode)
		query += fmt.Sprintf(" AND b.payment_mode::text = $%d", len(args))
	}
	query += ` ORDER BY b.created_at DESC`
	if limit > 0 {
		args = append(args, limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if start > 0 {
		args = append(args, start)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BillRow
	for rows.Next() {
		var row BillRow
		var storedCash, storedUPI *float64
		if err := rows.Scan(&row.Date, &row.BillNo, &row.CustomerName,
			&row.DiscountAmount, &row.TaxableAmount, &row.CGSTAmount, &row.SGSTAmount, &row.TotalAmount, &row.Status, &row.PaymentMode,
			&storedCash, &storedUPI); err != nil {
			return nil, err
		}
		row.ItemName = billRowItemName
		row.HSNCode = billRowHSNCode
		row.ItemCount = billRowItemCount
		row.TotalCash, row.TotalUPI = billRowCashUPISplit(row.PaymentMode, row.TotalAmount, storedCash, storedUPI)
		out = append(out, row)
	}
	return out, rows.Err()
}

// billRowCashUPISplit derives the report's Total Cash / Total UPI columns
// from the bill's payment mode: cash and credit sales put the full amount
// under cash (credit is still settled/tracked as cash once collected), upi
// sales put it all under UPI, and cash_upi sales use the actual split saved
// on the bill (nil when never set, e.g. an unapproved bill — treated as 0).
func billRowCashUPISplit(mode string, total float64, storedCash, storedUPI *float64) (cash, upi float64) {
	switch billing.PaymentMode(mode) {
	case billing.PaymentCash, billing.PaymentCredit:
		return total, 0
	case billing.PaymentUPI:
		return 0, total
	case billing.PaymentCashUPI:
		if storedCash != nil {
			cash = *storedCash
		}
		if storedUPI != nil {
			upi = *storedUPI
		}
		return cash, upi
	default:
		return 0, 0
	}
}

// ProductWise returns product-wise report rows. limit > 0 pages the result
// set; start is the offset applied when limit is set.
func (r *Repository) ProductWise(ctx context.Context, adminID int, f Filter, start, limit int) ([]ProductRow, error) {
	query := `
		SELECT to_char(b.created_at, 'DD/MM/YYYY'), b.bill_no, b.customer_name, bi.product_name, bi.hsn_code,
			bi.qty, bi.rate, bi.taxable_amount, bi.cgst_percent, bi.cgst_amount, bi.sgst_percent, bi.sgst_amount, bi.total_amount
		FROM bill_items bi
		JOIN bills b ON b.id = bi.bill_id
		WHERE b.admin_id = $1 AND b.created_at >= $2 AND b.created_at < $3`
	args := []any{adminID, f.From, f.To}
	if f.StaffID != nil {
		args = append(args, *f.StaffID)
		query += fmt.Sprintf(" AND b.sales_staff_id = $%d", len(args))
	}
	if f.BillNo != nil {
		args = append(args, likePrefix(*f.BillNo))
		query += fmt.Sprintf(" AND b.bill_no LIKE $%d ESCAPE '\\'", len(args))
	}
	if f.PaymentMode != nil {
		args = append(args, *f.PaymentMode)
		query += fmt.Sprintf(" AND b.payment_mode::text = $%d", len(args))
	}
	query += ` ORDER BY b.created_at DESC, bi.id`
	if limit > 0 {
		args = append(args, limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}
	if start > 0 {
		args = append(args, start)
		query += fmt.Sprintf(" OFFSET $%d", len(args))
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ProductRow
	for rows.Next() {
		var row ProductRow
		if err := rows.Scan(&row.Date, &row.BillNo, &row.CustomerName, &row.ItemName, &row.HSNCode,
			&row.Qty, &row.Rate, &row.TaxableAmount, &row.CGSTPercent, &row.CGSTAmount, &row.SGSTPercent, &row.SGSTAmount, &row.TotalAmount); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

// likePrefix turns a raw bill_no filter (e.g. "SFA" or a full bill number)
// into a "starts with" LIKE pattern, escaping any literal % or _ in the
// input so it can't be used to inject wildcard matches.
func likePrefix(s string) string {
	escaper := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)
	return escaper.Replace(s) + "%"
}
