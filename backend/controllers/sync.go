package controllers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"shosha_mart/backend/config"
	syncsvc "shosha_mart/backend/sync"
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
