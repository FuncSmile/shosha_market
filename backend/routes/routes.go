package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"shosha_mart_backend/config"
	"shosha_mart_backend/controllers"
	syncsvc "shosha_mart_backend/sync"
)

func Register(r *gin.Engine, db *gorm.DB, cfg config.AppConfig, worker *syncsvc.Worker) {
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://127.0.0.1:5173", "http://localhost:8080", "http://127.0.0.1:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			// Allow Electron file:// protocol dan localhost
			return origin == "" ||
				origin == "file://" ||
				origin == "http://localhost:5173" ||
				origin == "http://127.0.0.1:5173" ||
				origin == "http://localhost:8080" ||
				origin == "http://127.0.0.1:8080"
		},
	}))

	r.GET("/api/health", controllers.Health)

	r.GET("/api/products", controllers.ListProducts(db))
	r.POST("/api/products", controllers.CreateProduct(db, cfg))
	r.POST("/api/products/bulk", controllers.BulkCreateProducts(db, cfg))
	r.PUT("/api/products/:id", controllers.UpdateProduct(db, cfg))
	r.DELETE("/api/products/:id", controllers.DeleteProduct(db, cfg))

	r.GET("/api/branches", controllers.ListBranches(db))
	r.POST("/api/branches", controllers.CreateBranch(db, cfg))
	r.PUT("/api/branches/:id", controllers.UpdateBranch(db, cfg))
	r.DELETE("/api/branches/:id", controllers.DeleteBranch(db, cfg))

	r.POST("/api/sales", controllers.CreateSale(db, cfg))
	r.GET("/api/sales", controllers.ListSales(db))
	r.GET("/api/sales/:id", controllers.GetSale(db))
	r.DELETE("/api/sales/:id", controllers.DeleteSale(db))
	r.GET("/api/sales/export", controllers.ExportSalesReport(db))

	r.POST("/api/stock-opname", controllers.CreateStockOpname(db, cfg))

	r.GET("/api/sync/summary", controllers.SyncSummary(db, cfg, worker))
	r.POST("/api/sync/run", controllers.SyncRun(worker))
	r.POST("/api/sync/prune-deleted", controllers.PruneDeleted(db))
	r.GET("/api/analytics/sales", controllers.SalesAnalytics(db))

	// Debug endpoints
	r.GET("/api/debug/unsynced", controllers.UnsyncedCounts(db))

	r.GET("/api/reports/sales", controllers.SalesReport(db, cfg))
}
