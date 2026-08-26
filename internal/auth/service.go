package auth

import (
	"context"
	"errors"

	"salestrack/internal/staff"

	"golang.org/x/crypto/bcrypt"
)

var ErrBadCredentials = errors.New("mobile number or password is incorrect")

type Service struct {
	repo     *Repository
	staffSvc *staff.Service
}

func NewService(repo *Repository, staffSvc *staff.Service) *Service {
	return &Service{repo: repo, staffSvc: staffSvc}
}

func (s *Service) RegisterAdmin(ctx context.Context, req RegisterAdminRequest) (Admin, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return Admin{}, err
	}
	return s.repo.CreateAdmin(ctx, req, string(hash))
}

func (s *Service) LoginAdmin(ctx context.Context, req AdminLoginRequest) (Admin, string, error) {
	a, hash, err := s.repo.GetByMobile(ctx, req.MobileNumber)
	if err != nil {
		return Admin{}, "", ErrBadCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return Admin{}, "", ErrBadCredentials
	}
	token, err := GenerateToken(a.ID, a.ID, RoleAdmin, a.Name)
	return a, token, err
}

// LoginSalesStaff verifies mobile + admin-issued 4-digit code (the "OTP" on
// the mobile login screen is really this static code, mocked / non-SMS by
// design for now — swap in a real SMS OTP provider later without touching
// callers of this method). Only accounts created with role "sale_agent" can
// log in here — a cash-counter account must use LoginCashAgent instead.
func (s *Service) LoginSalesStaff(ctx context.Context, req SalesLoginRequest) (staff.SalesStaff, string, error) {
	member, err := s.staffSvc.VerifyLogin(ctx, req.MobileNumber, req.LoginCode, staff.RoleSaleAgent)
	if err != nil {
		return staff.SalesStaff{}, "", err
	}
	token, err := GenerateToken(member.ID, member.AdminID, RoleSalesStaff, member.Name)
	return member, token, err
}

// LoginCashAgent is the cash-counter staff login: same mobile + 4-digit
// code flow as LoginSalesStaff, restricted to accounts created with role
// "cash_agent". It issues the same RoleSalesStaff JWT — cash agents and
// sale agents share identical staff-level authority/access; this is purely
// a separate login entry point, not a distinct permission level.
func (s *Service) LoginCashAgent(ctx context.Context, req SalesLoginRequest) (staff.SalesStaff, string, error) {
	member, err := s.staffSvc.VerifyLogin(ctx, req.MobileNumber, req.LoginCode, staff.RoleCashAgent)
	if err != nil {
		return staff.SalesStaff{}, "", err
	}
	token, err := GenerateToken(member.ID, member.AdminID, RoleSalesStaff, member.Name)
	return member, token, err
}
