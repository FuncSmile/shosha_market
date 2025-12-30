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
			CreatedAt     string `json:"created_at"`
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

		// If caller provided created_at, try to parse it and set CreatedAt accordingly.
		if payload.CreatedAt != "" {
			// try RFC3339 first, fallback to date-only YYYY-MM-DD
			if t, err := time.Parse(time.RFC3339, payload.CreatedAt); err == nil {
				sale.CreatedAt = t
			} else if t2, err2 := time.Parse("2006-01-02", payload.CreatedAt); err2 == nil {
				sale.CreatedAt = t2
			}
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
		if err := db.Where("is_deleted = ?", false).Order("created_at DESC").Find(&sales).Error; err != nil {
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
		if err := db.First(&sale, "id = ? AND is_deleted = ?", id, false).Error; err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": "sale not found"})
			return
		}
		var items []models.SaleItem
		if err := db.Where("sale_id = ? AND is_deleted = ?", sale.ID, false).Find(&items).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		sale.Items = items
		c.JSON(http.StatusOK, sale)
	}
}

// UpdateSale updates sale metadata (date, branch, payment method, notes) - cannot edit items
func UpdateSale(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var payload struct {
			CreatedAt     string `json:"created_at"`
			BranchID      string `json:"branch_id"`
			PaymentMethod string `json:"payment_method"`
			Notes         string `json:"notes"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		// Check if sale exists
		var sale models.Sale
		if err := db.First(&sale, "id = ? AND is_deleted = ?", id, false).Error; err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": "sale not found"})
			return
		}

		// Build updates map
		updates := map[string]interface{}{
			"synced": false, // Mark as unsynced
		}

		if payload.CreatedAt != "" {
			// Parse and validate date
			parsedDate, err := time.Parse("2006-01-02", payload.CreatedAt)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format, use YYYY-MM-DD"})
				return
			}
			updates["created_at"] = parsedDate
		}

		if payload.BranchID != "" {
			// Validate branch exists
			var branch models.Branch
			if err := db.First(&branch, "id = ?", payload.BranchID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					c.JSON(http.StatusBadRequest, gin.H{"error": "branch not found"})
					return
				}
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			updates["branch_id"] = payload.BranchID
			updates["branch_name"] = branch.Name
		}

		if payload.PaymentMethod != "" {
			if payload.PaymentMethod != "cash" && payload.PaymentMethod != "hutang" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "payment_method must be 'cash' or 'hutang'"})
				return
			}
			updates["payment_method"] = payload.PaymentMethod
		}

		if payload.Notes != "" {
			updates["notes"] = payload.Notes
		}

		// Update sale
		if err := db.Model(&sale).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Reload to return fresh values
		if err := db.First(&sale, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, sale)
	}
}

// UpdateSaleItem updates quantity and price of an item in a sale
func UpdateSaleItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		saleID := c.Param("id")
		itemID := c.Param("itemId")

		var payload struct {
			Qty   int     `json:"qty"`
			Price float64 `json:"price"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		if payload.Qty <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "qty must be > 0"})
			return
		}

		// Check if sale exists
		var sale models.Sale
		if err := db.First(&sale, "id = ? AND is_deleted = ?", saleID, false).Error; err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": "sale not found"})
			return
		}

		// Get current item to know old qty
		var item models.SaleItem
		if err := db.First(&item, "id = ? AND sale_id = ? AND is_deleted = ?", itemID, saleID, false).Error; err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": "item not found"})
			return
		}

		// Calculate qty difference to adjust stock
		qtyDiff := payload.Qty - item.Qty
		oldTotal := item.Qty * int(item.Price)
		newTotal := payload.Qty * int(payload.Price)
		totalDiff := float64(newTotal - oldTotal)

		err := db.Transaction(func(tx *gorm.DB) error {
			// Update stock
			if err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock - ?", qtyDiff)).Error; err != nil {
				return err
			}

			// Update item
			if err := tx.Model(&item).Updates(map[string]interface{}{
				"qty":    payload.Qty,
				"price":  payload.Price,
				"synced": false,
			}).Error; err != nil {
				return err
			}

			// Update sale total
			if err := tx.Model(&sale).Updates(map[string]interface{}{
				"total":  gorm.Expr("total + ?", totalDiff),
				"synced": false,
			}).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, item)
	}
}

// AddSaleItem adds a new product to an existing sale
func AddSaleItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		saleID := c.Param("id")

		var payload struct {
			ProductID string  `json:"product_id"`
			Qty       int     `json:"qty"`
			Price     float64 `json:"price"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		if payload.ProductID == "" || payload.Qty <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "product_id and qty required"})
			return
		}

		// Check if sale exists
		var sale models.Sale
		if err := db.First(&sale, "id = ? AND is_deleted = ?", saleID, false).Error; err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": "sale not found"})
			return
		}

		// Validate product exists
		var product models.Product
		if err := db.First(&product, "id = ?", payload.ProductID).Error; err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status = http.StatusBadRequest
			}
			c.JSON(status, gin.H{"error": "product not found"})
			return
		}

		price := payload.Price
		if price <= 0 {
			price = product.Price
		}

		newItem := models.SaleItem{
			ID:        uuid.NewString(),
			SaleID:    saleID,
			ProductID: payload.ProductID,
			Qty:       payload.Qty,
			Price:     price,
			Synced:    false,
		}

		itemTotal := float64(payload.Qty) * price

		err := db.Transaction(func(tx *gorm.DB) error {
			// Create item
			if err := tx.Create(&newItem).Error; err != nil {
				return err
			}

			// Update stock
			if err := tx.Model(&models.Product{}).
				Where("id = ?", payload.ProductID).
				UpdateColumn("stock", gorm.Expr("stock - ?", payload.Qty)).Error; err != nil {
				return err
			}

			// Update sale total
			if err := tx.Model(&sale).Updates(map[string]interface{}{
				"total":  gorm.Expr("total + ?", itemTotal),
				"synced": false,
			}).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, newItem)
	}
}

// DeleteSaleItem removes a product from a sale and restores its stock
func DeleteSaleItem(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		saleID := c.Param("id")
		itemID := c.Param("itemId")

		// Check if sale exists
		var sale models.Sale
		if err := db.First(&sale, "id = ? AND is_deleted = ?", saleID, false).Error; err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": "sale not found"})
			return
		}

		// Get item
		var item models.SaleItem
		if err := db.First(&item, "id = ? AND sale_id = ? AND is_deleted = ?", itemID, saleID, false).Error; err != nil {
			status := http.StatusInternalServerError
			if errors.Is(err, gorm.ErrRecordNotFound) {
				status = http.StatusNotFound
			}
			c.JSON(status, gin.H{"error": "item not found"})
			return
		}

		itemSubtotal := float64(item.Qty) * item.Price

		err := db.Transaction(func(tx *gorm.DB) error {
			// Restore stock
			if err := tx.Model(&models.Product{}).
				Where("id = ?", item.ProductID).
				UpdateColumn("stock", gorm.Expr("stock + ?", item.Qty)).Error; err != nil {
				return err
			}

			// Soft-delete item
			if err := tx.Model(&item).Updates(map[string]interface{}{
				"is_deleted": true,
				"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
				"synced":     false,
			}).Error; err != nil {
				return err
			}

			// Update sale total
			if err := tx.Model(&sale).Updates(map[string]interface{}{
				"total":  gorm.Expr("total - ?", itemSubtotal),
				"synced": false,
			}).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "item deleted"})
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

// DeleteSale removes a sale and its items, and restores product stock
func DeleteSale(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		err := db.Transaction(func(tx *gorm.DB) error {
			// First, get all items to restore stock
			var items []models.SaleItem
			if err := tx.Where("sale_id = ? AND is_deleted = ?", id, false).Find(&items).Error; err != nil {
				return err
			}

			// Restore stock for each item
			for _, item := range items {
				if err := tx.Model(&models.Product{}).
					Where("id = ?", item.ProductID).
					UpdateColumn("stock", gorm.Expr("stock + ?", item.Qty)).Error; err != nil {
					return err
				}
			}

			// Soft-delete sale and its items
			if err := tx.Model(&models.SaleItem{}).
				Where("sale_id = ?", id).
				Updates(map[string]interface{}{
					"is_deleted": true,
					"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
					"synced":     false,
				}).Error; err != nil {
				return err
			}

			if err := tx.Model(&models.Sale{}).
				Where("id = ?", id).
				Updates(map[string]interface{}{
					"is_deleted": true,
					"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
					"synced":     false,
				}).Error; err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "sale deleted"})
	}
}
