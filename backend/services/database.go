package services

import (
    "fmt"
    "log"

    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"

    "github.com/FuncSmile/shosha_market/backend/config"
    "github.com/FuncSmile/shosha_market/backend/models"
)

// Connect opens SQLite database and performs migrations.
func Connect(cfg config.AppConfig) (*gorm.DB, error) {
    db, err := gorm.Open(sqlite.Open(cfg.DBPath), &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    })
    if err != nil {
        return nil, fmt.Errorf("open database: %w", err)
    }

    // Enable foreign keys.
    if err := db.Exec("PRAGMA foreign_keys = ON").Error; err != nil {
        log.Printf("warn: failed enabling foreign keys: %v", err)
    }

    if err := db.AutoMigrate(
        &models.Product{},
        &models.Branch{},
        &models.Sale{},
        &models.SaleItem{},
        &models.StockOpname{},
        &models.StockOpnameItem{},
        &models.SyncState{},
    ); err != nil {
        return nil, fmt.Errorf("auto migrate: %w", err)
    }

    return db, nil
}
