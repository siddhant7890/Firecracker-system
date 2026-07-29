package report

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) BillWise(ctx context.Context, adminID int, f Filter) ([]BillRow, error) {
	query := `
		SELECT to_char(b.created_at, 'DD/MM/YYYY'), b.bill_no, b.customer_name, s.name,
			(SELECT COUNT(*) FROM bill_items bi WHERE bi.bill_id = b.id),
			b.taxable_amount, b.cgst_amount, b.sgst_amount, b.total_amount, b.status,
			COALESCE(b.payment_mode::text, '')
		FROM bills b JOIN sales_staff s ON s.id = b.sales_staff_id
		WHERE b.admin_id = $1 AND b.created_at >= $2 AND b.created_at < $3`
	args := []any{adminID, f.From, f.To}
	if f.StaffID != nil {
		args = append(args, *f.StaffID)
		query += fmt.Sprintf(" AND b.sales_staff_id = $%d", len(args))
	}
	query += ` ORDER BY b.created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BillRow
	for rows.Next() {
		var row BillRow
		if err := rows.Scan(&row.Date, &row.BillNo, &row.CustomerName, &row.SalesPerson, &row.ItemCount,
			&row.TaxableAmount, &row.CGSTAmount, &row.SGSTAmount, &row.TotalAmount, &row.Status, &row.PaymentMode); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (r *Repository) ProductWise(ctx context.Context, adminID int, f Filter) ([]ProductRow, error) {
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
	query += ` ORDER BY b.created_at DESC, bi.id`

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
