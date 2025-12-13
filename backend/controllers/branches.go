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

// UpdateBranch updates an existing branch.
func UpdateBranch(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		var payload struct {
			Name    string `json:"name"`
			Address string `json:"address"`
			Phone   string `json:"phone"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		// Check if branch exists
		var branch models.Branch
		if err := db.First(&branch, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "branch not found"})
			return
		}

		// Update branch
		if err := db.Model(&branch).Updates(models.Branch{
			Name:    payload.Name,
			Address: payload.Address,
			Phone:   payload.Phone,
			Synced:  false, // Mark as unsynced when updated
		}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, branch)
	}
}

// DeleteBranch removes a branch from database.
func DeleteBranch(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		// Check if branch exists
		var branch models.Branch
		if err := db.First(&branch, "id = ?", id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "branch not found"})
			return
		}

		// Delete branch
		if err := db.Delete(&branch).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "branch deleted"})
	}
}
