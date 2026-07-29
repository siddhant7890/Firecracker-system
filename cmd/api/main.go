package main

import (
	"log"

	"salestrack/config"
	"salestrack/internal/admin"
	"salestrack/internal/auth"
	"salestrack/internal/billing"
	"salestrack/internal/middleware"
	"salestrack/internal/payment"
	"salestrack/internal/product"
	"salestrack/internal/report"
	"salestrack/internal/sales"
	"salestrack/internal/staff"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db := config.ConnectDB()
	defer config.CloseDB()

	// --- wire repositories -> services -> handlers, bottom-up ---
	staffRepo := staff.NewRepository(db)
	staffSvc := staff.NewService(staffRepo)
	staffHandler := staff.NewHandler(staffSvc)

	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, staffSvc)
	authHandler := auth.NewHandler(authSvc)

	productRepo := product.NewRepository(db)
	productSvc := product.NewService(productRepo)
	productHandler := product.NewHandler(productSvc)

	billingRepo := billing.NewRepository(db)
	billingSvc := billing.NewService(billingRepo, productSvc)
	salesHandler := sales.NewHandler(billingSvc)

	paymentSvc := payment.NewService(billingSvc)
	paymentHandler := payment.NewHandler(paymentSvc)

	adminSvc := admin.NewService(billingSvc)
	adminHandler := admin.NewHandler(billingSvc, paymentSvc)

	reportRepo := report.NewRepository(db)
	reportSvc := report.NewService(reportRepo)
	reportHandler := report.NewHandler(reportSvc)

	// --- router ---
	router := gin.Default()
	router.Use(middleware.CORS())

	router.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

	api := router.Group("/api")

	// Public auth endpoints.
	adminAuth := api.Group("/admin/auth")
	authHandler.RegisterAdminRoutes(adminAuth)

	salesAuth := api.Group("/sales/auth")
	authHandler.RegisterSalesRoutes(salesAuth)

	// Public payment webhook (authenticated by Razorpay signature, not JWT).
	payments := api.Group("/payments")
	paymentHandler.RegisterPublicRoutes(payments)

	// --- Admin webapp (JWT role=admin) ---
	adminGroup := api.Group("/admin", middleware.RequireAuth(auth.RoleAdmin))
	adminHandler.RegisterDashboardRoutes(adminGroup, adminSvc)

	adminProducts := adminGroup.Group("/products")
	productHandler.RegisterAdminRoutes(adminProducts)

	adminStaff := adminGroup.Group("/staff")
	staffHandler.Register(adminStaff)

	adminBills := adminGroup.Group("/bills")
	adminHandler.RegisterBillRoutes(adminBills)

	adminReports := adminGroup.Group("/reports")
	reportHandler.Register(adminReports)

	// --- Mobile app (JWT role=sales_staff) ---
	salesGroup := api.Group("/sales", middleware.RequireAuth(auth.RoleSalesStaff))
	salesHandler.Register(salesGroup)

	salesProducts := salesGroup.Group("/products")
	productHandler.RegisterSalesRoutes(salesProducts)

	log.Printf("SalesTrack API listening on :%s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
