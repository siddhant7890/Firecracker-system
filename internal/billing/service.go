package billing

import (
	"context"
	"errors"
	"fmt"
	"time"

	"salestrack/internal/product"
)

var ErrEmptyCart = errors.New("bill must have at least one item")

type Service struct {
	repo       *Repository
	productSvc *product.Service
}

func NewService(repo *Repository, productSvc *product.Service) *Service {
	return &Service{repo: repo, productSvc: productSvc}
}

func (s *Service) CreateBill(ctx context.Context, adminID, salesStaffID int, req CreateBillRequest) (Bill, error) {
	if len(req.Items) == 0 {
		return Bill{}, ErrEmptyCart
	}

	ids := make([]int, 0, len(req.Items))
	seen := map[int]bool{}
	for _, it := range req.Items {
		if !seen[it.ProductID] {
			ids = append(ids, it.ProductID)
			seen[it.ProductID] = true
		}
	}

	products, err := s.productSvc.GetMany(ctx, adminID, ids)
	if err != nil {
		return Bill{}, err
	}
	snapshots := map[int]ProductSnapshot{}
	for _, id := range ids {
		p, ok := products[id]
		if !ok || !p.IsActive {
			return Bill{}, fmt.Errorf("product %d is not available", id)
		}
		snapshots[id] = ProductSnapshot{ID: p.ID, Name: p.Name, HSNCode: p.HSNCode}
	}

	return s.repo.Create(ctx, adminID, salesStaffID, req, snapshots)
}

// ListForStaff serves the "Today / This Week / All" tabs on Bill History.
func (s *Service) ListForStaff(ctx context.Context, adminID, staffID int, rangeParam string) ([]Bill, error) {
	since := rangeStart(rangeParam)
	return s.repo.ListByStaff(ctx, adminID, staffID, since)
}

func rangeStart(rangeParam string) *time.Time {
	now := time.Now()
	switch rangeParam {
	case "week":
		t := now.AddDate(0, 0, -7)
		return &t
	case "all":
		return nil
	default: // "today"
		t := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		return &t
	}
}

func (s *Service) ListForAdmin(ctx context.Context, adminID int, f ListFilter) ([]Bill, error) {
	return s.repo.ListByAdmin(ctx, adminID, f)
}

func (s *Service) Get(ctx context.Context, adminID, id int) (Bill, error) {
	return s.repo.GetByID(ctx, adminID, id)
}

func (s *Service) Approve(ctx context.Context, adminID, id, approvedBy int, approverRole string, mode PaymentMode) (Bill, error) {
	return s.repo.Approve(ctx, adminID, id, approvedBy, approverRole, mode)
}

func (s *Service) UpdateBill(ctx context.Context, adminID, id int, req UpdateBillRequest) (Bill, error) {
	return s.repo.UpdateBill(ctx, adminID, id, req)
}

func (s *Service) Reject(ctx context.Context, adminID, id, approvedBy int, approverRole string) (Bill, error) {
	return s.repo.Reject(ctx, adminID, id, approvedBy, approverRole)
}

func (s *Service) SetRazorpayOrder(ctx context.Context, adminID, id int, orderID string) error {
	return s.repo.SetRazorpayOrder(ctx, adminID, id, orderID)
}

func (s *Service) ApproveByRazorpayOrder(ctx context.Context, orderID, paymentID string) (Bill, error) {
	return s.repo.ApproveByRazorpayOrder(ctx, orderID, paymentID)
}

func (s *Service) MarkWhatsappSent(ctx context.Context, adminID, id int) error {
	return s.repo.MarkWhatsappSent(ctx, adminID, id)
}

// SalesHome powers the mobile "Namaste, <name>" dashboard: today's bill
// count/sales total for this agent, plus their most recent bills.
func (s *Service) SalesHome(ctx context.Context, adminID, staffID int) (DashboardHome, error) {
	todayStart := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 0, 0, 0, 0, time.Now().Location())
	stats, err := s.repo.StatsSince(ctx, adminID, todayStart, &staffID)
	if err != nil {
		return DashboardHome{}, err
	}
	recent, err := s.repo.ListByStaff(ctx, adminID, staffID, nil)
	if err != nil {
		return DashboardHome{}, err
	}
	if len(recent) > 5 {
		recent = recent[:5]
	}
	return DashboardHome{TodaysBills: stats.BillsCount, TodaysSales: stats.SalesTotal, RecentBills: recent}, nil
}

func (s *Service) AdminStatsSince(ctx context.Context, adminID int, since time.Time) (DashboardStats, error) {
	return s.repo.StatsSince(ctx, adminID, since, nil)
}

func (s *Service) AdminDayStats(ctx context.Context, adminID int, dayStart, dayEnd time.Time) (AdminDayStats, error) {
	return s.repo.AdminDayStats(ctx, adminID, dayStart, dayEnd)
}

func (s *Service) PendingApprovalCount(ctx context.Context, adminID int) (int, error) {
	return s.repo.PendingApprovalCount(ctx, adminID)
}

func (s *Service) GSTSlabTotals(ctx context.Context, adminID int, from, to time.Time) ([]GSTSlabTotal, error) {
	return s.repo.GSTSlabTotals(ctx, adminID, from, to)
}

func (s *Service) ProductSalesTotals(ctx context.Context, adminID int, from, to time.Time) ([]ProductSalesTotal, error) {
	return s.repo.ProductSalesTotals(ctx, adminID, from, to)
}
