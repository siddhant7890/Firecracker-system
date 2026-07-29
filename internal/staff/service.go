package staff

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// generateCode returns a random 4-digit numeric login code, e.g. "4471".
// Leading zeros are allowed, matching the design (e.g. "2098").
func generateCode() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(10000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%04d", n.Int64()), nil
}

func hashCode(code string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(code), bcrypt.DefaultCost)
	return string(h), err
}

func (s *Service) Create(ctx context.Context, adminID int, req CreateStaffRequest) (CreateStaffResponse, error) {
	// code, err := generateCode()
	// if err != nil {
	// 	return CreateStaffResponse{}, err
	// }
	hash, err := hashCode(req.LoginCode)
	if err != nil {
		return CreateStaffResponse{}, err
	}
	created, err := s.repo.Create(ctx, adminID, req.Name, req.MobileNumber, hash)
	if err != nil {
		return CreateStaffResponse{}, err
	}
	return CreateStaffResponse{SalesStaff: created, LoginCode: req.LoginCode}, nil
}

func (s *Service) List(ctx context.Context, adminID int) ([]SalesStaff, error) {
	return s.repo.ListByAdmin(ctx, adminID)
}

func (s *Service) Get(ctx context.Context, adminID, id int) (SalesStaff, error) {
	return s.repo.GetByID(ctx, adminID, id)
}

func (s *Service) Update(ctx context.Context, adminID, id int, req UpdateStaffRequest) (SalesStaff, error) {
	return s.repo.Update(ctx, adminID, id, req.Name, req.MobileNumber)
}

func (s *Service) SetActive(ctx context.Context, adminID, id int, active bool) error {
	return s.repo.SetActive(ctx, adminID, id, active)
}

func (s *Service) Delete(ctx context.Context, adminID, id int) error {
	return s.repo.Delete(ctx, adminID, id)
}

// ResetCode issues a brand new login code for a staff member (e.g. "forgot PIN").
func (s *Service) ResetCode(ctx context.Context, adminID, id int) (ResetCodeResponse, error) {
	code, err := generateCode()
	if err != nil {
		return ResetCodeResponse{}, err
	}
	hash, err := hashCode(code)
	if err != nil {
		return ResetCodeResponse{}, err
	}
	if err := s.repo.ResetCode(ctx, adminID, id, hash); err != nil {
		return ResetCodeResponse{}, err
	}
	return ResetCodeResponse{LoginCode: code}, nil
}

// VerifyLogin checks a mobile+code pair at sales-agent login time.
func (s *Service) VerifyLogin(ctx context.Context, mobile, code string) (SalesStaff, error) {
	member, hash, err := s.repo.GetByMobile(ctx, mobile)
	if err != nil {
		return SalesStaff{}, err
	}
	if !member.IsActive {
		return SalesStaff{}, ErrInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)); err != nil {
		return SalesStaff{}, ErrBadCredentials
	}
	return member, nil
}
