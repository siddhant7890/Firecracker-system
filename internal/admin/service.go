package admin

import (
	"context"
	"time"

	"salestrack/internal/billing"
)

type Service struct {
	billing *billing.Service
}

func NewService(billingSvc *billing.Service) *Service {
	return &Service{billing: billingSvc}
}

func dayBounds(t time.Time) (time.Time, time.Time) {
	start := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return start, start.AddDate(0, 0, 1)
}

func (s *Service) Dashboard(ctx context.Context, adminID int) (DashboardResponse, error) {
	now := time.Now()
	todayStart, todayEnd := dayBounds(now)
	yesterdayStart, yesterdayEnd := dayBounds(now.AddDate(0, 0, -1))

	today, err := s.billing.AdminDayStats(ctx, adminID, todayStart, todayEnd)
	if err != nil {
		return DashboardResponse{}, err
	}
	yesterday, err := s.billing.AdminDayStats(ctx, adminID, yesterdayStart, yesterdayEnd)
	if err != nil {
		return DashboardResponse{}, err
	}

	pending, err := s.billing.PendingApprovalCount(ctx, adminID)
	if err != nil {
		return DashboardResponse{}, err
	}

	recent, err := s.billing.ListForAdmin(ctx, adminID, billing.ListFilter{})
	if err != nil {
		return DashboardResponse{}, err
	}
	if len(recent) > 5 {
		recent = recent[:5]
	}

	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	monthEnd := monthStart.AddDate(0, 1, 0)
	gstSnapshot, err := s.billing.GSTSlabTotals(ctx, adminID, monthStart, monthEnd)
	if err != nil {
		return DashboardResponse{}, err
	}

	resp := DashboardResponse{
		TodaysSales:          today.SalesTotal,
		BillsGenerated:       today.BillsGenerated,
		PendingApprovals:     pending,
		GSTCollectedToday:    today.GSTCollected,
		RecentBills:          recent,
		GSTSnapshotThisMonth: gstSnapshot,
	}
	if yesterday.SalesTotal > 0 {
		pct := ((today.SalesTotal - yesterday.SalesTotal) / yesterday.SalesTotal) * 100
		resp.TodaysSalesChangePct = &pct
	}
	billsChange := today.BillsGenerated - yesterday.BillsGenerated
	resp.BillsGeneratedChange = &billsChange

	return resp, nil
}
