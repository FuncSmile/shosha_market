package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "shosha_mart/backend/config"
    "shosha_mart/backend/models"
)

// ListProducts returns all products ordered by update time.
func ListProducts(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var products []models.Product
        if err := db.Order("updated_at desc").Find(&products).Error; err != nil {
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
            Name     string  `json:"name" binding:"required"`
            SKU      string  `json:"sku"`
            Unit     string  `json:"unit" binding:"required"`
            Quantity int     `json:"quantity" binding:"required,gt=0"`
            Stock    int     `json:"stock"`
            Price    float64 `json:"price" binding:"required,gt=0"`
        }
        if err := c.ShouldBindJSON(&payload); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }
        product := models.Product{
            ID:       uuid.NewString(),
            Name:     payload.Name,
            SKU:      payload.SKU,
            Unit:     payload.Unit,
            Quantity: payload.Quantity,
            Stock:    payload.Stock,
            Price:    payload.Price,
            Synced:   false,
            BranchID: cfg.BranchID,
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
			Name     string  `json:"name"`
			SKU      string  `json:"sku"`
			Unit     string  `json:"unit"`
			Quantity int     `json:"quantity"`
			Stock    int     `json:"stock"`
			Price    float64 `json:"price"`
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

		// Update product
		if err := db.Model(&product).Updates(models.Product{
			Name:     payload.Name,
			SKU:      payload.SKU,
			Unit:     payload.Unit,
			Quantity: payload.Quantity,
			Stock:    payload.Stock,
			Price:    payload.Price,
			Synced:   false, // Mark as unsynced when updated
		}).Error; err != nil {
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

		// Delete product
		if err := db.Delete(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "product deleted"})
	}
}
