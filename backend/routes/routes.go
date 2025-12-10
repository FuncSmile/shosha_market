package routes

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"shosha_mart/backend/config"
	"shosha_mart/backend/controllers"
	syncsvc "shosha_mart/backend/sync"
)

func Register(r *gin.Engine, db *gorm.DB, cfg config.AppConfig, worker *syncsvc.Worker) {
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:5173", "http://127.0.0.1:5173"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type"},
	}))

	r.GET("/api/health", controllers.Health)

	r.GET("/api/products", controllers.ListProducts(db))
	r.POST("/api/products", controllers.CreateProduct(db, cfg))

	r.GET("/api/branches", controllers.ListBranches(db))
	r.POST("/api/branches", controllers.CreateBranch(db, cfg))

	r.POST("/api/sales", controllers.CreateSale(db, cfg))

	r.POST("/api/stock-opname", controllers.CreateStockOpname(db, cfg))

	r.GET("/api/sync/summary", controllers.SyncSummary(db, cfg, worker))
	r.POST("/api/sync/run", controllers.SyncRun(worker))
	r.GET("/api/analytics/sales", controllers.SalesAnalytics(db))

	r.GET("/api/reports/sales", controllers.SalesReport(db, cfg))
}
