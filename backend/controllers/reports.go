package controllers

import (
    "net/http"
    "time"

    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    "shosha_mart/backend/config"
    "shosha_mart/backend/reports"
)

// SalesReport generates and returns the filepath for the requested period.
func SalesReport(db *gorm.DB, cfg config.AppConfig) gin.HandlerFunc {
    return func(c *gin.Context) {
        startStr := c.DefaultQuery("start", time.Now().AddDate(0, 0, -7).Format("2006-01-02"))
        endStr := c.DefaultQuery("end", time.Now().Format("2006-01-02"))

        start, err := time.Parse("2006-01-02", startStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date"})
            return
        }
        end, err := time.Parse("2006-01-02", endStr)
        if err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date"})
            return
        }
        end = end.Add(24*time.Hour - time.Nanosecond)

        path, err := reports.GenerateSalesReport(db, cfg, start, end)
        if err != nil {
            c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
            return
        }

        c.JSON(http.StatusOK, gin.H{"file": path})
    }
}
