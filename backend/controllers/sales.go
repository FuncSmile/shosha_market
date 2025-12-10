package controllers

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"shosha_mart/backend/config"
	"shosha_mart/backend/models"
)

// CreateSale records a checkout and decrements stock offline-first.
func CreateSale(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			BranchID  string `json:"branch_id"`
			ReceiptNo string `json:"receipt_no"`
			Items     []struct {
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
			}
		}

		sale := models.Sale{
			ID:        uuid.NewString(),
			ReceiptNo: payload.ReceiptNo,
			BranchID:  branchID,
			Synced:    false,
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
