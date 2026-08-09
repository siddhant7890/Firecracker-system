package staff

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("sales staff not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, adminID int, name, mobile, codeHash string) (SalesStaff, error) {
	var s SalesStaff
	err := r.db.QueryRow(ctx, `
		INSERT INTO sales_staff (admin_id, name, mobile_number, login_code_hash)
		VALUES ($1, $2, $3, $4)
		RETURNING id, admin_id, name, mobile_number, is_active, created_at
	`, adminID, name, mobile, codeHash).Scan(&s.ID, &s.AdminID, &s.Name, &s.MobileNumber, &s.IsActive, &s.CreatedAt)
	return s, err
}

func (r *Repository) ListByAdmin(ctx context.Context, adminID int) ([]SalesStaff, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, admin_id, name, mobile_number, is_active, created_at, login_code
		FROM sales_staff WHERE admin_id = $1 ORDER BY created_at DESC
	`, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []SalesStaff
	for rows.Next() {
		var s SalesStaff
		if err := rows.Scan(&s.ID, &s.AdminID, &s.Name, &s.MobileNumber, &s.IsActive, &s.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, adminID, id int) (SalesStaff, error) {
	var s SalesStaff
	err := r.db.QueryRow(ctx, `
		SELECT id, admin_id, name, mobile_number, is_active, created_at
		FROM sales_staff WHERE admin_id = $1 AND id = $2
	`, adminID, id).Scan(&s.ID, &s.AdminID, &s.Name, &s.MobileNumber, &s.IsActive, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrNotFound
	}
	return s, err
}

// GetByMobile looks a staff member up by mobile number alone (used at login,
// before we know which shop/admin they belong to).
func (r *Repository) GetByMobile(ctx context.Context, mobile string) (SalesStaff, string, error) {
	var s SalesStaff
	var hash string
	err := r.db.QueryRow(ctx, `
		SELECT id, admin_id, name, mobile_number, login_code_hash, is_active, created_at
		FROM sales_staff WHERE mobile_number = $1
	`, mobile).Scan(&s.ID, &s.AdminID, &s.Name, &s.MobileNumber, &hash, &s.IsActive, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, "", ErrNotFound
	}
	return s, hash, err
}

func (r *Repository) Update(ctx context.Context, adminID, id int, name, mobile *string) (SalesStaff, error) {
	var s SalesStaff
	err := r.db.QueryRow(ctx, `
		UPDATE sales_staff SET
			name = COALESCE($3, name),
			mobile_number = COALESCE($4, mobile_number)
		WHERE admin_id = $1 AND id = $2
		RETURNING id, admin_id, name, mobile_number, is_active, created_at
	`, adminID, id, name, mobile).Scan(&s.ID, &s.AdminID, &s.Name, &s.MobileNumber, &s.IsActive, &s.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return s, ErrNotFound
	}
	return s, err
}

func (r *Repository) SetActive(ctx context.Context, adminID, id int, active bool) error {
	tag, err := r.db.Exec(ctx, `UPDATE sales_staff SET is_active = $3 WHERE admin_id = $1 AND id = $2`, adminID, id, active)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) ResetCode(ctx context.Context, adminID, id int, codeHash string) error {
	tag, err := r.db.Exec(ctx, `UPDATE sales_staff SET login_code_hash = $3 WHERE admin_id = $1 AND id = $2`, adminID, id, codeHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, adminID, id int) error {
	tag, err := r.db.Exec(ctx, `UPDATE sales_staff SET is_deleted = true WHERE admin_id = $1 AND id = $2`, adminID, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
