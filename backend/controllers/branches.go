package controllers

import (
    "net/http"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
    "gorm.io/gorm"

    "shosha_mart/backend/config"
    "shosha_mart/backend/models"
)

func ListBranches(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        var branches []models.Branch
        if err := db.Order("updated_at desc").Find(&branches).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusOK, branches)
    }
}

func CreateBranch(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        var payload struct {
            Name    string `json:"name"`
            Address string `json:"address"`
            Phone   string `json:"phone"`
        }
        if err := c.ShouldBindJSON(&payload); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
            return
        }
        branch := models.Branch{
            ID:      uuid.NewString(),
            Name:    payload.Name,
            Address: payload.Address,
            Phone:   payload.Phone,
            Synced:  false,
        }
        if err := db.Create(&branch).Error; err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }
        c.JSON(http.StatusCreated, branch)
    }
}
