package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"shosha_mart_backend/config"
	"shosha_mart_backend/exports"
	"shosha_mart_backend/models"
)

// CreateSale records a checkout and decrements stock offline-first.
func CreateSale(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			BranchID      string `json:"branch_id"`
			ReceiptNo     string `json:"receipt_no"`
			PaymentMethod string `json:"payment_method"` // "cash" or "hutang"
			Notes         string `json:"notes"`
			Items         []struct {
				ProductID string  `json:"product_id"`
				Qty       int     `json:"qty"`
				Price     float64 `json:"price"`
			} `json:"items"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil || len(payload.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		type validatedItem struct {
			ProductID string
			Qty       int
			Price     float64
		}
		validated := make([]validatedItem, 0, len(payload.Items))
		for _, item := range payload.Items {
			if item.ProductID == "" || item.Qty <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "productId required and qty must be > 0"})
				return
			}
			var product models.Product
			if err := db.First(&product, "id = ?", item.ProductID).Error; err != nil {
				status := http.StatusInternalServerError
				if errors.Is(err, gorm.ErrRecordNotFound) {
					status = http.StatusBadRequest
				}
				c.JSON(status, gin.H{"error": fmt.Sprintf("product not found: %s", item.ProductID)})
				return
			}
			price := item.Price
			if price <= 0 {
				price = product.Price
			}
			validated = append(validated, validatedItem{
				ProductID: item.ProductID,
				Qty:       item.Qty,
				Price:     price,
			})
		}

		branchID := chooseBranch(payload.BranchID, cfg.BranchID)
		branchName := ""
		// Pastikan branch ada agar FK tidak gagal.
		if branchID != "" {
			var branch models.Branch
			if err := db.First(&branch, "id = ?", branchID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if branch.ID == "" && branchID != "" {
				// buat stub branch jika belum ada agar tidak gagal FK
				if err := db.Create(&models.Branch{
					ID:        branchID,
					Name:      branchID,
					Synced:    false,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}).Error; err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
				branchName = branchID
			} else {
				branchName = branch.Name
			}
		}

		paymentMethod := payload.PaymentMethod
		if paymentMethod != "cash" && paymentMethod != "hutang" {
			paymentMethod = "cash" // default
		}

		sale := models.Sale{
			ID:            uuid.NewString(),
			ReceiptNo:     payload.ReceiptNo,
			BranchID:      branchID,
			BranchName:    branchName,
			PaymentMethod: paymentMethod,
			Notes:         payload.Notes,
			Synced:        false,
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			// simpan sale lebih dulu agar FK sale_items -> sales tidak gagal
			if sale.ReceiptNo == "" {
				sale.ReceiptNo = generateReceiptNo()
			}
			if err := tx.Create(&sale).Error; err != nil {
				return err
			}

			var total float64
			for _, item := range validated {
				saleItem := models.SaleItem{
					ID:        uuid.NewString(),
					SaleID:    sale.ID,
					ProductID: item.ProductID,
					Qty:       item.Qty,
					Price:     item.Price,
					Synced:    false,
				}

				if err := tx.Create(&saleItem).Error; err != nil {
					return err
				}

				if err := tx.Model(&models.Product{}).
					Where("id = ?", item.ProductID).
					UpdateColumn("stock", gorm.Expr("stock - ?", item.Qty)).Error; err != nil {
					return err
				}

				total += float64(item.Qty) * item.Price
			}
			sale.Total = total
			return tx.Model(&sale).Updates(map[string]interface{}{
				"total":      total,
				"updated_at": time.Now(),
			}).Error
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, sale)
	}
}

func chooseBranch(payload, fallback string) string {
	if payload != "" {
		return payload
	}
	return fallback
}

func generateReceiptNo() string {
	return time.Now().Format("060102150405")
}

// ListSales returns recent sales with basic fields (including BranchName and Total)
func ListSales(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var sales []models.Sale
		if err := db.Order("created_at DESC").Find(&sales).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, sales)
	}
}

// GetSale returns a sale by id including its items with computed subtotals
func GetSale(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var sale models.Sale
		if err := db.First(&sale, "id = ?", id).Error; err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": "sale not found"})
			return
		}
		var items []models.SaleItem
		if err := db.Where("sale_id = ?", sale.ID).Find(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		sale.Items = items
		c.JSON(http.StatusOK, sale)
	}
}

// ExportSalesReport generates an Excel file with sales grouped by branch (one sheet per branch)
func ExportSalesReport(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		f, err := exports.ExportSalesByBranch(db)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		filename := fmt.Sprintf("sales-export-%s.xlsx", time.Now().Format("20060102-150405"))
		c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		if err := f.Write(c.Writer); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}
