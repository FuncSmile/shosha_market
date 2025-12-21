package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// UnsyncedCounts returns counts of records with synced = false per table.
func UnsyncedCounts(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var products int64
		var branches int64
		var sales int64
		var saleItems int64
		var opnames int64
		var opItems int64

		// Use Count; if table doesn't exist, treat as 0 (avoid error)
		_ = db.Table("products").Where("synced = ?", false).Count(&products).Error
		_ = db.Table("branches").Where("synced = ?", false).Count(&branches).Error
		_ = db.Table("sales").Where("synced = ?", false).Count(&sales).Error
		_ = db.Table("sale_items").Where("synced = ?", false).Count(&saleItems).Error
		_ = db.Table("stock_opnames").Where("synced = ?", false).Count(&opnames).Error
		_ = db.Table("stock_opname_items").Where("synced = ?", false).Count(&opItems).Error

		c.JSON(http.StatusOK, gin.H{
			"products":           products,
			"branches":           branches,
			"sales":              sales,
			"sale_items":         saleItems,
			"stock_opnames":      opnames,
			"stock_opname_items": opItems,
		})
	}
}
