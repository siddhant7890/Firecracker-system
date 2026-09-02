package product

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("product not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

const selectCols = `id, admin_id, item_code, name, category, hsn_code, unit, taxable_value, gst_percent, final_price, COALESCE(image_url, ''), is_active, created_at, updated_at`

func scanProduct(row pgx.Row) (Product, error) {
	var p Product
	err := row.Scan(&p.ID, &p.AdminID, &p.ItemCode, &p.Name, &p.Category, &p.HSNCode, &p.Unit, &p.TaxableValue, &p.GSTPercent, &p.FinalPrice, &p.ImageURL, &p.IsActive, &p.CreatedAt, &p.UpdatedAt)
	return p, err
}

func (r *Repository) Create(ctx context.Context, adminID int, req CreateProductRequest) (Product, error) {
	hsn := req.HSNCode
	if hsn == "" {
		hsn = "3604"
	}
	unit := req.Unit
	if unit == "" {
		unit = "Piece"
	}
	row := r.db.QueryRow(ctx, `
		INSERT INTO products (admin_id, item_code, name, category, hsn_code, unit, taxable_value, gst_percent, image_url)
		VALUES ($1, NULLIF($2, ''), $3, $4, $5, $6, $7, $8, NULLIF($9, ''))
		RETURNING `+selectCols,
		adminID, req.ItemCode, req.Name, req.Category, hsn, unit, req.TaxableValue, req.GSTPercent, req.ImageURL)
	return scanProduct(row)
}

// BulkUpsert inserts or updates products keyed on (admin_id, item_code): an
// existing item_code gets its fields overwritten, a new one is inserted.
// is_active/is_deleted are left untouched on update. Requires the partial
// unique index on (admin_id, item_code) added in migration 004.
func (r *Repository) BulkUpsert(ctx context.Context, adminID int, rows []BulkUploadRow) (inserted, updated int, err error) {
	if len(rows) == 0 {
		return 0, 0, nil
	}

	batch := &pgx.Batch{}
	for _, row := range rows {
		batch.Queue(`
			INSERT INTO products (admin_id, item_code, name, category, hsn_code, unit, taxable_value, gst_percent)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
			ON CONFLICT (admin_id, item_code) WHERE item_code IS NOT NULL DO UPDATE SET
				name = EXCLUDED.name,
				category = EXCLUDED.category,
				hsn_code = EXCLUDED.hsn_code,
				unit = EXCLUDED.unit,
				taxable_value = EXCLUDED.taxable_value,
				gst_percent = EXCLUDED.gst_percent,
				updated_at = now()
			RETURNING (xmax = 0) AS inserted`,
			adminID, row.ItemCode, row.Name, row.Category, row.HSNCode, row.Unit, row.TaxableValue, row.GSTPercent)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for range rows {
		var wasInsert bool
		if scanErr := br.QueryRow().Scan(&wasInsert); scanErr != nil {
			return inserted, updated, scanErr
		}
		if wasInsert {
			inserted++
		} else {
			updated++
		}
	}
	return inserted, updated, nil
}

// ListByAdmin powers both the admin "Product Management" table (activeOnly
// false) and the staff "New Bill" product picker (activeOnly true, which
// also excludes soft-deleted products). search matches against item_code,
// name, and category. start/limit page the result set.
func (r *Repository) ListByAdmin(ctx context.Context, adminID int, activeOnly bool, category, search string, start, limit int) ([]Product, error) {
	query := `SELECT ` + selectCols + ` FROM products WHERE admin_id = $1`
	args := []any{adminID}

	if true {
		query += ` AND is_deleted = false`
	}
	if category != "" && category != "All" {
		args = append(args, category)
		query += ` AND category = $` + itoa(len(args))
	}
	if search != "" {
		args = append(args, "%"+search+"%")
		n := itoa(len(args))
		query += ` AND (item_code ILIKE $` + n + ` OR name ILIKE $` + n + ` OR category ILIKE $` + n + `)`
	}
	query += ` ORDER BY category, name`
	if limit > 0 {
		args = append(args, limit)
		query += ` LIMIT $` + itoa(len(args))
	}
	if start > 0 {
		args = append(args, start)
		query += ` OFFSET $` + itoa(len(args))
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, adminID, id int) (Product, error) {
	row := r.db.QueryRow(ctx, `SELECT `+selectCols+` FROM products WHERE admin_id = $1 AND id = $2`, adminID, id)
	p, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

// GetByIDs is used when building a bill, to fetch the current authoritative
// price/GST for every product line in one round trip.
func (r *Repository) GetByIDs(ctx context.Context, adminID int, ids []int) (map[int]Product, error) {
	ids32 := make([]int32, len(ids))
	for i, id := range ids {
		ids32[i] = int32(id)
	}
	rows, err := r.db.Query(ctx, `SELECT `+selectCols+` FROM products WHERE admin_id = $1 AND id = ANY($2)`, adminID, ids32)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := map[int]Product{}
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out[p.ID] = p
	}
	return out, rows.Err()
}

func (r *Repository) Update(ctx context.Context, adminID, id int, req UpdateProductRequest) (Product, error) {
	row := r.db.QueryRow(ctx, `
		UPDATE products SET
			name = COALESCE($3, name),
			category = COALESCE($4, category),
			hsn_code = COALESCE($5, hsn_code),
			unit = COALESCE($6, unit),
			taxable_value = COALESCE($7, taxable_value),
			gst_percent = COALESCE($8, gst_percent),
			image_url = COALESCE($9, image_url),
			updated_at = now()
		WHERE admin_id = $1 AND id = $2
		RETURNING `+selectCols,
		adminID, id, req.Name, req.Category, req.HSNCode, req.Unit, req.TaxableValue, req.GSTPercent, req.ImageURL)
	p, err := scanProduct(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return p, ErrNotFound
	}
	return p, err
}

func (r *Repository) SetActive(ctx context.Context, adminID, id int, active bool) error {
	tag, err := r.db.Exec(ctx, `UPDATE products SET is_active = $3, updated_at = now() WHERE admin_id = $1 AND id = $2`, adminID, id, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, adminID, id int) error {
	tag, err := r.db.Exec(ctx, `UPDATE products
     SET is_deleted = TRUE
     WHERE admin_id = $1 AND id = $2`, adminID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteAll soft-deletes every not-yet-deleted product for this admin, e.g.
// for a "start fresh" reset of the product catalog. Returns how many rows
// were affected.
func (r *Repository) DeleteAll(ctx context.Context, adminID int) (int, error) {
	tag, err := r.db.Exec(ctx, `UPDATE products SET is_deleted = TRUE, updated_at = now() WHERE admin_id = $1 AND is_deleted = false`, adminID)
	if err != nil {
		return 0, err
	}
	return int(tag.RowsAffected()), nil
}

func itoa(n int) string {
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if len(digits) == 0 {
		return "0"
	}
	return string(digits)
}
