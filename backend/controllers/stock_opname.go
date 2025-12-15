package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"shosha_mart_backend/config"
	"shosha_mart_backend/models"
)

func CreateStockOpname(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		var payload struct {
			PerformedBy string `json:"performedBy"`
			Note        string `json:"note"`
			Items       []struct {
				ProductID   string `json:"productId"`
				SystemQty   int    `json:"systemQty"`
				PhysicalQty int    `json:"physicalQty"`
			} `json:"items"`
		}
		if err := c.ShouldBindJSON(&payload); err != nil || len(payload.Items) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid payload"})
			return
		}

		opname := models.StockOpname{
			ID:          uuid.NewString(),
			BranchID:    cfg.BranchID,
			PerformedBy: payload.PerformedBy,
			Note:        payload.Note,
			Synced:      false,
		}

		err := db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&opname).Error; err != nil {
				return err
			}
			for _, item := range payload.Items {
				detail := models.StockOpnameItem{
					ID:            uuid.NewString(),
					StockOpnameID: opname.ID,
					ProductID:     item.ProductID,
					SystemQty:     item.SystemQty,
					PhysicalQty:   item.PhysicalQty,
					Synced:        false,
				}
				if err := tx.Create(&detail).Error; err != nil {
					return err
				}
				// Bring stock in line with physical count.
				if err := tx.Model(&models.Product{}).
					Where("id = ?", item.ProductID).
					UpdateColumn("stock", item.PhysicalQty).Error; err != nil {
					return err
				}
			}
			return nil
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, opname)
	}
}
