package product

import (
	"context"
	"io"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, adminID int, req CreateProductRequest) (Product, error) {
	return s.repo.Create(ctx, adminID, req)
}

// List with activeOnly=false is for the admin Product Management table
// (shows disabled products too, with a toggle). activeOnly=true is for the
// sales-agent "New Bill" product picker (also excludes soft-deleted rows).
func (s *Service) List(ctx context.Context, adminID int, activeOnly bool, category, search string, start, limit int) ([]Product, error) {
	return s.repo.ListByAdmin(ctx, adminID, activeOnly, category, search, start, limit)
}

// BulkUpload parses a bulk-upload spreadsheet and upserts its rows by
// item_code: an existing item_code updates that product, a new one inserts.
// Invalid rows are skipped and reported rather than failing the whole file.
func (s *Service) BulkUpload(ctx context.Context, adminID int, file io.Reader) (BulkUploadResult, error) {
	rows, skipped, err := parseBulkUploadRows(file)
	if err != nil {
		return BulkUploadResult{}, err
	}

	inserted, updated, err := s.repo.BulkUpsert(ctx, adminID, rows)
	if err != nil {
		return BulkUploadResult{}, err
	}

	return BulkUploadResult{Inserted: inserted, Updated: updated, Skipped: skipped}, nil
}

func (s *Service) Get(ctx context.Context, adminID, id int) (Product, error) {
	return s.repo.GetByID(ctx, adminID, id)
}

// GetMany fetches several products in one round trip (used when pricing a
// whole bill's worth of line items).
func (s *Service) GetMany(ctx context.Context, adminID int, ids []int) (map[int]Product, error) {
	return s.repo.GetByIDs(ctx, adminID, ids)
}

func (s *Service) Update(ctx context.Context, adminID, id int, req UpdateProductRequest) (Product, error) {
	return s.repo.Update(ctx, adminID, id, req)
}

func (s *Service) SetActive(ctx context.Context, adminID, id int, active bool) error {
	return s.repo.SetActive(ctx, adminID, id, active)
}

func (s *Service) Delete(ctx context.Context, adminID, id int) error {
	return s.repo.Delete(ctx, adminID, id)
}
