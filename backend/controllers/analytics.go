package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"shosha_mart/backend/models"
)

type AnalyticsResponse struct {
	Start        string           `json:"start"`
	End          string           `json:"end"`
	TotalRevenue float64          `json:"totalRevenue"`
	TotalOrders  int64            `json:"totalOrders"`
	TotalItems   int64            `json:"totalItems"`
	PerDay       []AnalyticsDaily `json:"perDay"`
}

type AnalyticsDaily struct {
	Day     string  `json:"day"`
	Orders  int64   `json:"orders"`
	Items   int64   `json:"items"`
	Revenue float64 `json:"revenue"`
}

// SalesAnalytics returns quick numbers for dashboard.
func SalesAnalytics(db *gorm.DB) gin.HandlerFunc {
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
		// inclusive end
		end = end.Add(24*time.Hour - time.Nanosecond)

		var totalRevenue float64
		var totalOrders int64
		if err := db.Model(&models.Sale{}).
			Where("created_at BETWEEN ? AND ?", start, end).
			Count(&totalOrders).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Get total revenue with COALESCE to handle NULL
		if err := db.Model(&models.Sale{}).
			Where("created_at BETWEEN ? AND ?", start, end).
			Select("COALESCE(SUM(total), 0)").
			Scan(&totalRevenue).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var totalItems int64
		if err := db.Model(&models.SaleItem{}).
			Joins("JOIN sales ON sales.id = sale_items.sale_id").
			Where("sales.created_at BETWEEN ? AND ?", start, end).
			Select("COALESCE(SUM(qty), 0)").
			Scan(&totalItems).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var perDay []AnalyticsDaily
		if err := db.Table("sales").
			Select("strftime('%Y-%m-%d', created_at) as day, COUNT(*) as orders, COALESCE(SUM(total), 0) as revenue").
			Where("created_at BETWEEN ? AND ?", start, end).
			Group("day").
			Order("day").
			Scan(&perDay).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// fill item count per day
		for i := range perDay {
			var items int64
			db.Table("sale_items").
				Joins("JOIN sales ON sales.id = sale_items.sale_id").
				Where("strftime('%Y-%m-%d', sales.created_at) = ?", perDay[i].Day).
				Select("COALESCE(SUM(qty), 0)").Scan(&items)
			perDay[i].Items = items
		}

		c.JSON(http.StatusOK, AnalyticsResponse{
			Start:        start.Format("2006-01-02"),
			End:          endStr,
			TotalRevenue: totalRevenue,
			TotalOrders:  totalOrders,
			TotalItems:   totalItems,
			PerDay:       perDay,
		})
	}
}
