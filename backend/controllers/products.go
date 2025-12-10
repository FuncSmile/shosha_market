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
            Name  string  `json:"name"`
            Stock int     `json:"stock"`
            Price float64 `json:"price"`
        }
        if err := c.ShouldBindJSON(&payload); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
            return
        }
        product := models.Product{
            ID:       uuid.NewString(),
            Name:     payload.Name,
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
