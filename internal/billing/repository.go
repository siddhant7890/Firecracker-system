package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"salestrack/pkg/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound       = errors.New("bill not found")
	ErrAlreadyHandled = errors.New("this bill has already been approved or rejected")
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ProductSnapshot confirms the product_id on a bill line belongs to this shop
// and is active, and carries its name/HSN code for the bill_items snapshot.
// Pricing/GST are computed by the frontend and taken from the request as-is.
type ProductSnapshot struct {
	ID      int
	Name    string
	HSNCode string
}

// Create inserts a bill + its line items atomically, generating the next
// sequential bill number for the shop (e.g. "SF/26-27/00483").
func (r *Repository) Create(ctx context.Context, adminID, salesStaffID int, req CreateBillRequest, products map[int]ProductSnapshot) (Bill, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Bill{}, err
	}
	defer tx.Rollback(ctx)

	var prefix string
	if err := tx.QueryRow(ctx, `SELECT bill_prefix FROM admins WHERE id = $1`, adminID).Scan(&prefix); err != nil {
		return Bill{}, fmt.Errorf("shop not found: %w", err)
	}

	var seq int
	err = tx.QueryRow(ctx, `
		INSERT INTO bill_sequences (admin_id, last_number) VALUES ($1, 1)
		ON CONFLICT (admin_id) DO UPDATE SET last_number = bill_sequences.last_number + 1
		RETURNING last_number
	`, adminID).Scan(&seq)
	if err != nil {
		return Bill{}, err
	}

	now := time.Now()
	fy := utils.FinancialYear(now.Year(), int(now.Month()))
	billNo := fmt.Sprintf("%s/%s/%05d", prefix, fy, seq)

	customerName := req.CustomerName
	if customerName == "" {
		customerName = "Walk-in Customer"
	}

	// Pricing/GST for each line is computed by the frontend and trusted as-is;
	// the product lookup here is only to confirm the product_id belongs to
	// this shop and is active, and to snapshot its name/HSN code.
	var taxableTotal, cgstTotal, sgstTotal, grandTotal float64
	items := make([]BillItem, 0, len(req.Items))
	for _, line := range req.Items {
		p, ok := products[line.ProductID]
		if !ok {
			return Bill{}, fmt.Errorf("product %d is not available", line.ProductID)
		}

		items = append(items, BillItem{
			ProductID: p.ID, ProductName: p.Name, HSNCode: p.HSNCode, Qty: line.Qty,
			Rate: line.Rate, TaxableAmount: line.TaxableAmount, GSTPercent: line.GSTPercent,
			CGSTPercent: line.CGSTPercent, CGSTAmount: line.CGSTAmount,
			SGSTPercent: line.SGSTPercent, SGSTAmount: line.SGSTAmount, TotalAmount: line.TotalAmount,
		})
		taxableTotal += line.TaxableAmount
		cgstTotal += line.CGSTAmount
		sgstTotal += line.SGSTAmount
		grandTotal += line.TotalAmount
	}
	taxableTotal, cgstTotal, sgstTotal, grandTotal = utils.Round2(taxableTotal), utils.Round2(cgstTotal), utils.Round2(sgstTotal), utils.Round2(grandTotal)

	var bill Bill
	err = tx.QueryRow(ctx, `
		INSERT INTO bills (admin_id, sales_staff_id, bill_no, financial_year, customer_name, customer_mobile,
			token_number, number_of_cartoon, gst_number,
			taxable_amount, cgst_amount, sgst_amount, total_amount, status)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), NULLIF($7, ''), NULLIF($8, 0), NULLIF($9, ''), $10, $11, $12, $13, 'pending')
		RETURNING id, bill_no, financial_year, customer_name, COALESCE(customer_mobile, ''),
			COALESCE(token_number, ''), COALESCE(number_of_cartoon, 0), COALESCE(gst_number, ''),
			taxable_amount, cgst_amount, sgst_amount, total_amount, status, whatsapp_sent, created_at
	`, adminID, salesStaffID, billNo, fy, customerName, req.CustomerMobile,
		req.TokenNumber, req.NumberOfCartoon, req.GSTNumber, taxableTotal, cgstTotal, sgstTotal, grandTotal).
		Scan(&bill.ID, &bill.BillNo, &bill.FinancialYear, &bill.CustomerName, &bill.CustomerMobile,
			&bill.TokenNumber, &bill.NumberOfCartoon, &bill.GSTNumber,
			&bill.TaxableAmount, &bill.CGSTAmount, &bill.SGSTAmount, &bill.TotalAmount, &bill.Status, &bill.WhatsappSent, &bill.CreatedAt)
	if err != nil {
		return Bill{}, err
	}
	bill.AdminID = adminID
	bill.SalesStaffID = salesStaffID

	batch := &pgx.Batch{}
	for _, it := range items {
		batch.Queue(`
			INSERT INTO bill_items (bill_id, product_id, product_name, hsn_code, qty, rate, taxable_amount,
				gst_percent, cgst_percent, cgst_amount, sgst_percent, sgst_amount, total_amount)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		`, bill.ID, it.ProductID, it.ProductName, it.HSNCode, it.Qty, it.Rate, it.TaxableAmount,
			it.GSTPercent, it.CGSTPercent, it.CGSTAmount, it.SGSTPercent, it.SGSTAmount, it.TotalAmount)
	}
	br := tx.SendBatch(ctx, batch)
	for range items {
		if _, err := br.Exec(); err != nil {
			br.Close()
			return Bill{}, err
		}
	}
	if err := br.Close(); err != nil {
		return Bill{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Bill{}, err
	}

	bill.Items = items
	bill.ItemCount = len(items)
	return bill, nil
}

const billSelectCols = `
	b.id, b.admin_id, b.sales_staff_id, s.name, b.bill_no, b.financial_year, b.customer_name,
	COALESCE(b.customer_mobile, ''), COALESCE(b.token_number, ''), COALESCE(b.number_of_cartoon, 0), COALESCE(b.gst_number, ''),
	b.taxable_amount, b.cgst_amount, b.sgst_amount, b.total_amount,
	b.status, b.payment_mode, COALESCE(b.razorpay_order_id, ''), COALESCE(b.razorpay_payment_id, ''),
	b.whatsapp_sent, b.approved_at, b.created_at,
	(SELECT COUNT(*) FROM bill_items bi WHERE bi.bill_id = b.id)
`

func scanBillRow(row pgx.Row) (Bill, error) {
	var b Bill
	// Scan the two Postgres enum columns (bill_status, payment_mode_t) into
	// plain strings first — safest option without registering their OIDs
	// with pgx's type map — then convert to our named types.
	var status string
	var mode *string
	err := row.Scan(&b.ID, &b.AdminID, &b.SalesStaffID, &b.SalesStaffName, &b.BillNo, &b.FinancialYear,
		&b.CustomerName, &b.CustomerMobile, &b.TokenNumber, &b.NumberOfCartoon, &b.GSTNumber,
		&b.TaxableAmount, &b.CGSTAmount, &b.SGSTAmount, &b.TotalAmount,
		&status, &mode, &b.RazorpayOrderID, &b.RazorpayPaymentID, &b.WhatsappSent, &b.ApprovedAt, &b.CreatedAt, &b.ItemCount)
	b.Status = Status(status)
	if mode != nil {
		pm := PaymentMode(*mode)
		b.PaymentMode = &pm
	}
	return b, err
}

// ListByStaff powers the sales-agent "Bill History" screen (Today / This
// Week / All tabs).
func (r *Repository) ListByStaff(ctx context.Context, adminID, staffID int, since *time.Time) ([]Bill, error) {
	query := `SELECT ` + billSelectCols + ` FROM bills b JOIN sales_staff s ON s.id = b.sales_staff_id
		WHERE b.admin_id = $1 AND b.sales_staff_id = $2`
	args := []any{adminID, staffID}
	if since != nil {
		args = append(args, *since)
		query += ` AND b.created_at >= $3`
	}
	query += ` ORDER BY b.created_at DESC`
	return r.queryBills(ctx, query, args...)
}

// ListByAdmin powers the admin recent-bills tables, with optional status
// and date-range filters (used by Cash Counter + Report Management).
type ListFilter struct {
	Status  *Status
	StaffID *int
	From    *time.Time
	To      *time.Time
}

func (r *Repository) ListByAdmin(ctx context.Context, adminID int, f ListFilter) ([]Bill, error) {
	query := `SELECT ` + billSelectCols + ` FROM bills b JOIN sales_staff s ON s.id = b.sales_staff_id WHERE b.admin_id = $1`
	args := []any{adminID}

	if f.Status != nil {
		args = append(args, *f.Status)
		query += fmt.Sprintf(" AND b.status = $%d", len(args))
	}
	if f.StaffID != nil {
		args = append(args, *f.StaffID)
		query += fmt.Sprintf(" AND b.sales_staff_id = $%d", len(args))
	}
	if f.From != nil {
		args = append(args, *f.From)
		query += fmt.Sprintf(" AND b.created_at >= $%d", len(args))
	}
	if f.To != nil {
		args = append(args, *f.To)
		query += fmt.Sprintf(" AND b.created_at < $%d", len(args))
	}
	query += ` ORDER BY b.created_at DESC`
	return r.queryBills(ctx, query, args...)
}

func (r *Repository) queryBills(ctx context.Context, query string, args ...any) ([]Bill, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Bill
	for rows.Next() {
		b, err := scanBillRow(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.attachItems(ctx, out); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, adminID, id int) (Bill, error) {
	row := r.db.QueryRow(ctx, `SELECT `+billSelectCols+` FROM bills b JOIN sales_staff s ON s.id = b.sales_staff_id
		WHERE b.admin_id = $1 AND b.id = $2`, adminID, id)
	b, err := scanBillRow(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return b, ErrNotFound
	}
	if err != nil {
		return b, err
	}

	bills := []Bill{b}
	if err := r.attachItems(ctx, bills); err != nil {
		return b, err
	}
	return bills[0], nil
}

// attachItems batch-fetches bill_items for every bill in the slice (one
// query, not one per bill) and populates each Bill.Items in place.
func (r *Repository) attachItems(ctx context.Context, bills []Bill) error {
	if len(bills) == 0 {
		return nil
	}
	ids := make([]int, len(bills))
	for i, b := range bills {
		ids[i] = b.ID
	}

	rows, err := r.db.Query(ctx, `
		SELECT bill_id, id, product_id, product_name, hsn_code, qty, rate, taxable_amount, gst_percent,
			cgst_percent, cgst_amount, sgst_percent, sgst_amount, total_amount
		FROM bill_items WHERE bill_id = ANY($1)
		ORDER BY bill_id, id
	`, ids)
	if err != nil {
		return err
	}
	defer rows.Close()

	itemsByBill := make(map[int][]BillItem)
	for rows.Next() {
		var billID int
		var it BillItem
		var productID *int
		if err := rows.Scan(&billID, &it.ID, &productID, &it.ProductName, &it.HSNCode, &it.Qty, &it.Rate, &it.TaxableAmount,
			&it.GSTPercent, &it.CGSTPercent, &it.CGSTAmount, &it.SGSTPercent, &it.SGSTAmount, &it.TotalAmount); err != nil {
			return err
		}
		if productID != nil {
			it.ProductID = *productID
		}
		itemsByBill[billID] = append(itemsByBill[billID], it)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for i := range bills {
		bills[i].Items = itemsByBill[bills[i].ID]
	}
	return nil
}

// Approve marks a pending bill approved/rejected from the Cash Counter, or
// confirms it once a Razorpay UPI payment has actually been captured.
// approvedBy is an admins(id) or a sales_staff(id) depending on approverRole.
func (r *Repository) Approve(ctx context.Context, adminID, id, approvedBy int, approverRole string, mode PaymentMode) (Bill, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE bills SET status = 'approved', payment_mode = $3, approved_by = $4, approved_by_role = $5, approved_at = now()
		WHERE admin_id = $1 AND id = $2 AND status = 'pending'
		RETURNING id
	`, adminID, id, mode, approvedBy, approverRole)
	var returnedID int
	if err := row.Scan(&returnedID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Either it doesn't exist, or it's not pending anymore.
			if _, getErr := r.GetByID(ctx, adminID, id); getErr == nil {
				return Bill{}, ErrAlreadyHandled
			}
			return Bill{}, ErrNotFound
		}
		return Bill{}, err
	}
	return r.GetByID(ctx, adminID, id)
}

// UpdatePaymentMode lets the admin correct the payment_mode recorded on a
// bill (e.g. cash entered by mistake instead of UPI, or vice versa).
func (r *Repository) UpdatePaymentMode(ctx context.Context, adminID, id int, mode PaymentMode) (Bill, error) {
	tag, err := r.db.Exec(ctx, `UPDATE bills SET payment_mode = $3 WHERE admin_id = $1 AND id = $2`, adminID, id, mode)
	if err != nil {
		return Bill{}, err
	}
	if tag.RowsAffected() == 0 {
		return Bill{}, ErrNotFound
	}
	return r.GetByID(ctx, adminID, id)
}

func (r *Repository) Reject(ctx context.Context, adminID, id, approvedBy int, approverRole string) (Bill, error) {
	tag, err := r.db.Exec(ctx, `
		UPDATE bills SET status = 'rejected', approved_by = $3, approved_by_role = $4, approved_at = now()
		WHERE admin_id = $1 AND id = $2 AND status = 'pending'
	`, adminID, id, approvedBy, approverRole)
	if err != nil {
		return Bill{}, err
	}
	if tag.RowsAffected() == 0 {
		return Bill{}, ErrAlreadyHandled
	}
	return r.GetByID(ctx, adminID, id)
}

// SetRazorpayOrder records the order created for a UPI bill, so the webhook
// can later match the payment back to this bill.
func (r *Repository) SetRazorpayOrder(ctx context.Context, adminID, id int, orderID string) error {
	tag, err := r.db.Exec(ctx, `UPDATE bills SET razorpay_order_id = $3 WHERE admin_id = $1 AND id = $2`, adminID, id, orderID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// ApproveByRazorpayOrder is called from the payment webhook once Razorpay
// confirms a UPI payment was captured.
func (r *Repository) ApproveByRazorpayOrder(ctx context.Context, orderID, paymentID string) (Bill, error) {
	var adminID, id int
	err := r.db.QueryRow(ctx, `
		UPDATE bills SET status = 'approved', payment_mode = 'upi', razorpay_payment_id = $2, approved_at = now()
		WHERE razorpay_order_id = $1 AND status = 'pending'
		RETURNING admin_id, id
	`, orderID, paymentID).Scan(&adminID, &id)
	if errors.Is(err, pgx.ErrNoRows) {
		return Bill{}, ErrNotFound
	}
	if err != nil {
		return Bill{}, err
	}
	return r.GetByID(ctx, adminID, id)
}

func (r *Repository) MarkWhatsappSent(ctx context.Context, adminID, id int) error {
	tag, err := r.db.Exec(ctx, `UPDATE bills SET whatsapp_sent = true, whatsapp_sent_at = now() WHERE admin_id = $1 AND id = $2`, adminID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DashboardStats powers both the sales-agent Home screen and the admin
// "Today's overview" tiles.
type DashboardStats struct {
	BillsCount  int
	SalesTotal  float64
	GSTTotal    float64
	PendingCount int
}

func (r *Repository) StatsSince(ctx context.Context, adminID int, since time.Time, staffID *int) (DashboardStats, error) {
	query := `
		SELECT COUNT(*), COALESCE(SUM(total_amount), 0), COALESCE(SUM(cgst_amount + sgst_amount), 0)
		FROM bills WHERE admin_id = $1 AND created_at >= $2 AND status != 'rejected'`
	args := []any{adminID, since}
	if staffID != nil {
		args = append(args, *staffID)
		query += fmt.Sprintf(" AND sales_staff_id = $%d", len(args))
	}
	var stats DashboardStats
	if err := r.db.QueryRow(ctx, query, args...).Scan(&stats.BillsCount, &stats.SalesTotal, &stats.GSTTotal); err != nil {
		return stats, err
	}

	pendingQuery := `SELECT COUNT(*) FROM bills WHERE admin_id = $1 AND status = 'pending'`
	if err := r.db.QueryRow(ctx, pendingQuery, adminID).Scan(&stats.PendingCount); err != nil {
		return stats, err
	}
	return stats, nil
}

// AdminDayStats powers a single day's worth of the admin "Today's overview"
// tiles: bills generated (any status) and realized sales/GST (approved only).
type AdminDayStats struct {
	BillsGenerated int     `json:"bills_generated"`
	SalesTotal     float64 `json:"sales_total"`
	GSTCollected   float64 `json:"gst_collected"`
}

func (r *Repository) AdminDayStats(ctx context.Context, adminID int, dayStart, dayEnd time.Time) (AdminDayStats, error) {
	var stats AdminDayStats
	err := r.db.QueryRow(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE status != 'rejected'),
			COALESCE(SUM(total_amount) FILTER (WHERE status = 'approved'), 0),
			COALESCE(SUM(cgst_amount + sgst_amount) FILTER (WHERE status = 'approved'), 0)
		FROM bills
		WHERE admin_id = $1 AND created_at >= $2 AND created_at < $3
	`, adminID, dayStart, dayEnd).Scan(&stats.BillsGenerated, &stats.SalesTotal, &stats.GSTCollected)
	return stats, err
}

func (r *Repository) PendingApprovalCount(ctx context.Context, adminID int) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM bills WHERE admin_id = $1 AND status = 'pending'`, adminID).Scan(&n)
	return n, err
}

// GSTSlabTotals powers the "GST Snapshot — this month" widget, grouping
// taxable sales by the GST% slab of each line item.
type GSTSlabTotal struct {
	GSTPercent float64 `json:"gst_percent"`
	SalesTotal float64 `json:"sales_total"`
}

func (r *Repository) GSTSlabTotals(ctx context.Context, adminID int, from, to time.Time) ([]GSTSlabTotal, error) {
	rows, err := r.db.Query(ctx, `
		SELECT bi.gst_percent, COALESCE(SUM(bi.total_amount), 0)
		FROM bill_items bi
		JOIN bills b ON b.id = bi.bill_id
		WHERE b.admin_id = $1 AND b.status != 'rejected' AND b.created_at >= $2 AND b.created_at < $3
		GROUP BY bi.gst_percent
		ORDER BY bi.gst_percent DESC
	`, adminID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []GSTSlabTotal
	for rows.Next() {
		var g GSTSlabTotal
		if err := rows.Scan(&g.GSTPercent, &g.SalesTotal); err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

// ProductSalesTotals powers the "Product-wise Sales — this month" dashboard
// array, grouping sales by product.
type ProductSalesTotal struct {
	ProductID   int     `json:"product_id"`
	ProductName string  `json:"product_name"`
	QtySold     int     `json:"qty_sold"`
	SalesTotal  float64 `json:"sales_total"`
}

func (r *Repository) ProductSalesTotals(ctx context.Context, adminID int, from, to time.Time) ([]ProductSalesTotal, error) {
	rows, err := r.db.Query(ctx, `
		SELECT bi.product_id, bi.product_name, COALESCE(SUM(bi.qty), 0), COALESCE(SUM(bi.total_amount), 0)
		FROM bill_items bi
		JOIN bills b ON b.id = bi.bill_id
		WHERE b.admin_id = $1 AND b.status != 'rejected' AND b.created_at >= $2 AND b.created_at < $3
		GROUP BY bi.product_id, bi.product_name
		ORDER BY SUM(bi.total_amount) DESC
	`, adminID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ProductSalesTotal
	for rows.Next() {
		var p ProductSalesTotal
		var productID *int
		if err := rows.Scan(&productID, &p.ProductName, &p.QtySold, &p.SalesTotal); err != nil {
			return nil, err
		}
		if productID != nil {
			p.ProductID = *productID
		}
		out = append(out, p)
	}
	return out, rows.Err()
}
