package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"shosha_mart_backend/models"
)

type UploadPayload struct {
	BranchID         string                   `json:"branch_id"`
	Products         []models.Product         `json:"products"`
	Branches         []models.Branch          `json:"branches"`
	Sales            []models.Sale            `json:"sales"`
	SaleItems        []models.SaleItem        `json:"sale_items"`
	StockOpnames     []models.StockOpname     `json:"stock_opnames"`
	StockOpnameItems []models.StockOpnameItem `json:"stock_opname_items"`
}

type ChangesResponse struct {
	Products         []models.Product         `json:"products"`
	Branches         []models.Branch          `json:"branches"`
	Sales            []models.Sale            `json:"sales"`
	SaleItems        []models.SaleItem        `json:"sale_items"`
	StockOpnames     []models.StockOpname     `json:"stock_opnames"`
	StockOpnameItems []models.StockOpnameItem `json:"stock_opname_items"`
	LastSyncAt       *time.Time               `json:"last_sync_at"`
}

func main() {
	_ = godotenv.Load()

	dsn := os.Getenv("PG_DSN")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		// fallback PG envs
		dsn = "host=localhost user=postgres password=postgres dbname=pos_sync port=5432 sslmode=disable"
	}
	bind := os.Getenv("UPSTREAM_BIND")
	if bind == "" {
		bind = ":9000"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	if err := db.AutoMigrate(&models.Product{}, &models.Branch{}, &models.Sale{}, &models.SaleItem{}, &models.StockOpname{}, &models.StockOpnameItem{}); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	r := gin.Default()
	r.POST("/api/sync/upload", func(c *gin.Context) {
		var payload UploadPayload
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if len(payload.Branches) > 0 {
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "address", "phone", "synced", "updated_at", "created_at"}),
			}).Create(&payload.Branches).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if len(payload.Products) > 0 {
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "stock", "price", "branch_id", "synced", "updated_at", "created_at"}),
			}).Create(&payload.Products).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if len(payload.Sales) > 0 {
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"receipt_no", "branch_id", "total", "synced", "updated_at", "created_at"}),
			}).Create(&payload.Sales).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if len(payload.SaleItems) > 0 {
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"sale_id", "product_id", "qty", "price", "synced", "updated_at", "created_at"}),
			}).Create(&payload.SaleItems).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if len(payload.StockOpnames) > 0 {
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"branch_id", "performed_by", "note", "synced", "updated_at", "created_at"}),
			}).Create(&payload.StockOpnames).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		if len(payload.StockOpnameItems) > 0 {
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"stock_opname_id", "product_id", "system_qty", "physical_qty", "synced", "updated_at", "created_at"}),
			}).Create(&payload.StockOpnameItems).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/api/sync/changes", func(c *gin.Context) {
		sinceStr := c.Query("since")
		var since time.Time
		if sinceStr != "" {
			t, err := time.Parse(time.RFC3339, sinceStr)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid since"})
				return
			}
			since = t
		} else {
			since = time.Time{} // epoch -> all data
		}
		var (
			products []models.Product
			branches []models.Branch
			sales    []models.Sale
			items    []models.SaleItem
			opnames  []models.StockOpname
			opItems  []models.StockOpnameItem
		)
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&products)
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&branches)
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&sales)
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&items)
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&opnames)
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&opItems)

		now := time.Now().UTC()
		c.JSON(http.StatusOK, ChangesResponse{
			Products:         products,
			Branches:         branches,
			Sales:            sales,
			SaleItems:        items,
			StockOpnames:     opnames,
			StockOpnameItems: opItems,
			LastSyncAt:       &now,
		})
	})

	log.Printf("Upstream sync API listening on %s (Postgres DSN: %s)", bind, dsn)
	if err := r.Run(bind); err != nil {
		log.Fatal(err)
	}
}
