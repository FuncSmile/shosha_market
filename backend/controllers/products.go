package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"shosha_mart_backend/config"
	"shosha_mart_backend/models"
)

// ListProducts returns all products ordered by update time.
func ListProducts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products []models.Product
		if err := db.Where("is_deleted = ?", false).Order("updated_at desc").Find(&products).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, products)
	}
}

// CreateProduct inserts a new local product record flagged as unsynced.
func CreateProduct(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			Name          string  `json:"name" binding:"required"`
			Unit          string  `json:"unit" binding:"required"`
			Stock         int     `json:"stock"`
			Price         float64 `json:"price"` // Optional, will be set from investor/shosha
			PriceInvestor float64 `json:"price_investor"`
			PriceShosha   float64 `json:"price_shosha"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validasi: minimal salah satu harga harus > 0
		if payload.PriceInvestor <= 0 && payload.PriceShosha <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "at least one price (investor or shosha) must be greater than 0"})
			return
		}

		// Set price dari investor atau shosha jika tidak diisi
		price := payload.Price
		if price <= 0 {
			if payload.PriceInvestor > 0 {
				price = payload.PriceInvestor
			} else {
				price = payload.PriceShosha
			}
		}

		product := models.Product{
			ID:            uuid.NewString(),
			Name:          payload.Name,
			Unit:          payload.Unit,
			Stock:         payload.Stock,
			Price:         price,
			PriceInvestor: payload.PriceInvestor,
			PriceShosha:   payload.PriceShosha,
			Synced:        false,
			BranchID:      cfg.BranchID,
		}
		// Default values jika salah satu kosong
		if product.PriceInvestor <= 0 {
			product.PriceInvestor = product.Price
		}
		if product.PriceShosha <= 0 {
			product.PriceShosha = product.Price
		}
		if err := db.Create(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, product)
	}
}

// UpdateProduct updates an existing product.
func UpdateProduct(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var payload struct {
			Name          string  `json:"name"`
			Unit          string  `json:"unit"`
			Stock         *int    `json:"stock"` // pointer untuk detect null
			Price         float64 `json:"price"`
			PriceInvestor float64 `json:"price_investor"`
			PriceShosha   float64 `json:"price_shosha"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		// Check if product exists
		var product models.Product
		if err := db.First(&product, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}

		// Build updates map hanya untuk field yang dikirim
		updates := map[string]interface{}{
			"synced": false, // Always mark as unsynced
		}

		if payload.Name != "" {
			updates["name"] = payload.Name
		}
		if payload.Unit != "" {
			updates["unit"] = payload.Unit
		}
		if payload.Stock != nil {
			updates["stock"] = *payload.Stock
		}
		if payload.Price > 0 {
			updates["price"] = payload.Price
		}
		if payload.PriceInvestor > 0 {
			updates["price_investor"] = payload.PriceInvestor
		}
		if payload.PriceShosha > 0 {
			updates["price_shosha"] = payload.PriceShosha
		}

		if err := db.Model(&product).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// Reload to return fresh values
		if err := db.First(&product, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, product)
	}
}

// DeleteProduct removes a product from database.
func DeleteProduct(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Check if product exists
		var product models.Product
		if err := db.First(&product, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
			return
		}

		// Soft delete product (tombstone)
		updates := map[string]interface{}{
			"is_deleted": true,
			"deleted_at": gorm.Expr("CURRENT_TIMESTAMP"),
			"synced":     false,
		}
		if err := db.Model(&product).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
	}
}

// BulkCreateProducts inserts multiple products in one request.
func BulkCreateProducts(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	type Row struct {
		Name          string  `json:"name" binding:"required"`
		Unit          string  `json:"unit" binding:"required"`
		Stock         int     `json:"stock"`
		Price         float64 `json:"price"`
		PriceInvestor float64 `json:"price_investor"`
		PriceShosha   float64 `json:"price_shosha"`
	}
	return func(c *gin.Context) {
		var rows []Row
		if err := c.ShouldBindJSON(&rows); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		created := make([]models.Product, 0, len(rows))
		for _, r := range rows {
			// Validasi: minimal salah satu harga > 0
			if r.PriceInvestor <= 0 && r.PriceShosha <= 0 {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "at least one price (investor or shosha) must be greater than 0 for product: " + r.Name,
				})
				return
			}

			// Set price dari investor atau shosha jika tidak diisi
			price := r.Price
			if price <= 0 {
				if r.PriceInvestor > 0 {
					price = r.PriceInvestor
				} else {
					price = r.PriceShosha
				}
			}

			p := models.Product{
				ID:            uuid.NewString(),
				Name:          r.Name,
				Unit:          r.Unit,
				Stock:         r.Stock,
				Price:         price,
				PriceInvestor: r.PriceInvestor,
				PriceShosha:   r.PriceShosha,
				Synced:        false,
				BranchID:      cfg.BranchID,
			}
			// Default values jika salah satu kosong
			if p.PriceInvestor <= 0 {
				p.PriceInvestor = p.Price
			}
			if p.PriceShosha <= 0 {
				p.PriceShosha = p.Price
			}
			if err := db.Create(&p).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			created = append(created, p)
		}
		c.JSON(http.StatusCreated, gin.H{"count": len(created), "items": created})
	}
}
