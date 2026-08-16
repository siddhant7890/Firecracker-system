package admin

import "salestrack/internal/billing"

// DashboardResponse matches the "Today's overview" admin screen: four
// stat tiles, the recent-bills table, and the GST snapshot for the month.
type DashboardResponse struct {
	TodaysSales           float64                     `json:"todays_sales"`
	TodaysSalesChangePct  *float64                    `json:"todays_sales_change_pct,omitempty"`
	BillsGenerated        int                         `json:"bills_generated"`
	BillsGeneratedChange  *int                        `json:"bills_generated_change,omitempty"`
	PendingApprovals      int                         `json:"pending_approvals"`
	GSTCollectedToday     float64                     `json:"gst_collected_today"`
	RecentBills           []billing.Bill              `json:"recent_bills"`
	GSTSnapshotThisMonth  []billing.GSTSlabTotal      `json:"gst_snapshot_this_month"`
	ProductSalesThisMonth []billing.ProductSalesTotal `json:"product_sales_this_month"`
}
