package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"shosha_mart_backend/config"
	"shosha_mart_backend/models"
	syncsvc "shosha_mart_backend/sync"
)

func SyncSummary(db *gorm.DB, cfg config.AppConfig, worker *syncsvc.Worker) gin.HandlerFunc {
	return func(c *gin.Context) {
		status, lastErr, _ := worker.Status()
		summary, err := syncsvc.Build(db, cfg.DBPath, status, lastErr)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, summary)
	}
}

func SyncRun(worker *syncsvc.Worker) gin.HandlerFunc {
	return func(c *gin.Context) {
		if worker == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "sync worker not initialized"})
			return
		}
		if err := worker.RunOnce(context.Background()); err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}

// PruneDeleted hard-deletes rows that have been tombstoned and already synced
func PruneDeleted(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Order matters for FKs: items before parents
		if err := db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.SaleItem{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.Sale{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.StockOpnameItem{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.StockOpname{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.Product{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.Branch{}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "pruned"})
	}
}
