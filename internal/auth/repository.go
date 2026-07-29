package auth

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("admin not found")

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAdmin(ctx context.Context, req RegisterAdminRequest, passwordHash string) (Admin, error) {
	prefix := req.BillPrefix
	if prefix == "" {
		prefix = "SF"
	}

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return Admin{}, err
	}
	defer tx.Rollback(ctx)

	var a Admin
	err = tx.QueryRow(ctx, `
		INSERT INTO admins (name, mobile_number, email, password_hash, shop_name, bill_prefix)
		VALUES ($1, $2, NULLIF($3, ''), $4, $5, $6)
		RETURNING id, name, mobile_number, COALESCE(email, ''), shop_name, bill_prefix, created_at
	`, req.Name, req.MobileNumber, req.Email, passwordHash, req.ShopName, prefix).
		Scan(&a.ID, &a.Name, &a.MobileNumber, &a.Email, &a.ShopName, &a.BillPrefix, &a.CreatedAt)
	if err != nil {
		return Admin{}, err
	}

	if _, err := tx.Exec(ctx, `INSERT INTO bill_sequences (admin_id, last_number) VALUES ($1, 0)`, a.ID); err != nil {
		return Admin{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return Admin{}, err
	}
	return a, nil
}

func (r *Repository) GetByMobile(ctx context.Context, mobile string) (Admin, string, error) {
	var a Admin
	var hash string
	err := r.db.QueryRow(ctx, `
		SELECT id, name, mobile_number, COALESCE(email, ''), password_hash, shop_name, bill_prefix, created_at
		FROM admins WHERE mobile_number = $1
	`, mobile).Scan(&a.ID, &a.Name, &a.MobileNumber, &a.Email, &hash, &a.ShopName, &a.BillPrefix, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, "", ErrNotFound
	}
	return a, hash, err
}

func (r *Repository) GetByID(ctx context.Context, id int) (Admin, error) {
	var a Admin
	err := r.db.QueryRow(ctx, `
		SELECT id, name, mobile_number, COALESCE(email, ''), shop_name, bill_prefix, created_at
		FROM admins WHERE id = $1
	`, id).Scan(&a.ID, &a.Name, &a.MobileNumber, &a.Email, &a.ShopName, &a.BillPrefix, &a.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return a, ErrNotFound
	}
	return a, err
}
