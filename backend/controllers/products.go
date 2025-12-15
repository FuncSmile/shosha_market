package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/FuncSmile/shosha_market/backend/config"
	"github.com/FuncSmile/shosha_market/backend/models"
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
			Name  string  `json:"name" binding:"required"`
			Unit  string  `json:"unit" binding:"required"`
			Stock int     `json:"stock"`
			Price float64 `json:"price" binding:"required,gt=0"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		product := models.Product{
			ID:       uuid.NewString(),
			Name:     payload.Name,
			Unit:     payload.Unit,
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
			Name  string  `json:"name"`
			Unit  string  `json:"unit"`
			Stock int     `json:"stock"`
			Price float64 `json:"price"`
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
			Name:   payload.Name,
			Unit:   payload.Unit,
			Stock:  payload.Stock,
			Price:  payload.Price,
			Synced: false, // Mark as unsynced when updated
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

// BulkCreateProducts inserts multiple products in one request.
func BulkCreateProducts(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	type Row struct {
		Name  string  `json:"name" binding:"required"`
		Unit  string  `json:"unit" binding:"required"`
		Stock int     `json:"stock"`
		Price float64 `json:"price" binding:"required,gt=0"`
	}
	return func(c *gin.Context) {
		var rows []Row
		if err := c.ShouldBindJSON(&rows); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		created := make([]models.Product, 0, len(rows))
		for _, r := range rows {
			p := models.Product{
				ID:       uuid.NewString(),
				Name:     r.Name,
				Unit:     r.Unit,
				Stock:    r.Stock,
				Price:    r.Price,
				Synced:   false,
				BranchID: cfg.BranchID,
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
