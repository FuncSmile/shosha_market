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
			for _, b := range payload.Branches {
				if b.IsDeleted {
					if err := db.Delete(&models.Branch{}, "id = ?", b.ID).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					continue
				}
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"code", "name", "address", "phone", "synced", "is_deleted", "updated_at", "created_at"}),
			}).Create(&b).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
		if len(payload.Products) > 0 {
			for _, p := range payload.Products {
				if p.IsDeleted {
					if err := db.Delete(&models.Product{}, "id = ?", p.ID).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					continue
				}
				if err := db.Clauses(clause.OnConflict{
					Columns:   []clause.Column{{Name: "id"}},
					DoUpdates: clause.AssignmentColumns([]string{"name", "unit", "stock", "price", "price_investor", "price_shosha", "branch_id", "synced", "is_deleted", "updated_at", "created_at"}),
				}).Create(&p).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
		if len(payload.Sales) > 0 {
			for _, s := range payload.Sales {
				if s.IsDeleted {
					// delete items first then sale
					if err := db.Where("sale_id = ?", s.ID).Delete(&models.SaleItem{}).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					if err := db.Delete(&models.Sale{}, "id = ?", s.ID).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					continue
				}
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"receipt_no", "branch_id", "branch_name", "payment_method", "notes", "total", "synced", "is_deleted", "updated_at", "created_at"}),
			}).Create(&s).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
		if len(payload.SaleItems) > 0 {
			for _, si := range payload.SaleItems {
				if si.IsDeleted {
					if err := db.Delete(&models.SaleItem{}, "id = ?", si.ID).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					continue
				}
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"sale_id", "product_id", "qty", "price", "synced", "is_deleted", "updated_at", "created_at"}),
			}).Create(&si).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
		if len(payload.StockOpnames) > 0 {
			for _, so := range payload.StockOpnames {
				if so.IsDeleted {
					if err := db.Where("id = ?", so.ID).Delete(&models.StockOpname{}).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					continue
				}
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"branch_id", "performed_by", "note", "synced", "is_deleted", "updated_at", "created_at"}),
			}).Create(&so).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
		if len(payload.StockOpnameItems) > 0 {
			for _, soi := range payload.StockOpnameItems {
				if soi.IsDeleted {
					if err := db.Delete(&models.StockOpnameItem{}, "id = ?", soi.ID).Error; err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}
					continue
				}
			if err := db.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"stock_opname_id", "product_id", "system_qty", "physical_qty", "synced", "is_deleted", "updated_at", "created_at"}),
			}).Create(&soi).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
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
		branchID := c.Query("branch_id")
		var (
			products []models.Product
			branches []models.Branch
			sales    []models.Sale
			items    []models.SaleItem
			opnames  []models.StockOpname
			opItems  []models.StockOpnameItem
		)
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&products)
		db.Find(&branches)
		log.Printf("[SYNC] Branches found: %d, error: %v", len(branches), db.Error)
		if branchID != "" {
			db.Where("(updated_at >= ? OR created_at >= ?) AND branch_id = ?", since, since, branchID).Preload("Items").Find(&sales)
		} else {
			db.Where("updated_at >= ? OR created_at >= ?", since, since).Preload("Items").Find(&sales)
		}
		log.Printf("[SYNC] Sales found: %d, error: %v", len(sales), db.Error)
		log.Printf("[SYNC] Products found: %d", len(products))
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&items)
		log.Printf("[SYNC] SaleItems found: %d", len(items))
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&opnames)
		db.Where("updated_at >= ? OR created_at >= ?", since, since).Find(&opItems)

		now := time.Now().UTC()
		log.Printf("[SYNC] Sending response: Products=%d, Branches=%d, Sales=%d, SaleItems=%d, Opnames=%d, OpItems=%d",
			len(products), len(branches), len(sales), len(items), len(opnames), len(opItems))
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
