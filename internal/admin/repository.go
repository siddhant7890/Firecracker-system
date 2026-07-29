package admin

// No direct DB access needed here: all admin-portal data (dashboard stats,
// bill approvals, GST snapshot) is read through billing.Service, and
// products/staff CRUD live in their own packages (product, staff). See
// service.go and handler.go.
