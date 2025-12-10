package models

import "time"

// Product represents an inventory item stored locally first.
type Product struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name"`
    Stock     int       `json:"stock"`
    Price     float64   `json:"price"`
    Synced    bool      `json:"synced"`
    BranchID  string    `json:"branchId"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// Branch represents a store branch entry.
type Branch struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    Name      string    `json:"name"`
    Address   string    `json:"address"`
    Phone     string    `json:"phone"`
    Synced    bool      `json:"synced"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// Sale captures a checkout transaction.
type Sale struct {
    ID         string     `json:"id" gorm:"primaryKey"`
    ReceiptNo  string     `json:"receiptNo"`
    BranchID   string     `json:"branchId"`
    Total      float64    `json:"total"`
    Synced     bool       `json:"synced"`
    CreatedAt  time.Time  `json:"createdAt"`
    UpdatedAt  time.Time  `json:"updatedAt"`
    Items      []SaleItem `json:"items"`
}

// SaleItem links to Sale.
type SaleItem struct {
    ID        string    `json:"id" gorm:"primaryKey"`
    SaleID    string    `json:"saleId"`
    ProductID string    `json:"productId"`
    Qty       int       `json:"qty"`
    Price     float64   `json:"price"`
    Synced    bool      `json:"synced"`
    CreatedAt time.Time `json:"createdAt"`
    UpdatedAt time.Time `json:"updatedAt"`
}

// StockOpname represents a stock take session.
type StockOpname struct {
    ID         string             `json:"id" gorm:"primaryKey"`
    BranchID   string             `json:"branchId"`
    PerformedBy string            `json:"performedBy"`
    Note       string             `json:"note"`
    Synced     bool               `json:"synced"`
    CreatedAt  time.Time          `json:"createdAt"`
    UpdatedAt  time.Time          `json:"updatedAt"`
    Items      []StockOpnameItem  `json:"items"`
}

// StockOpnameItem holds per-product opname figures.
type StockOpnameItem struct {
    ID            string    `json:"id" gorm:"primaryKey"`
    StockOpnameID string    `json:"stockOpnameId"`
    ProductID     string    `json:"productId"`
    SystemQty     int       `json:"systemQty"`
    PhysicalQty   int       `json:"physicalQty"`
    Synced        bool      `json:"synced"`
    CreatedAt     time.Time `json:"createdAt"`
    UpdatedAt     time.Time `json:"updatedAt"`
}

// SyncState stores last sync metadata.
type SyncState struct {
    ID         string     `json:"id" gorm:"primaryKey"`
    LastSyncAt *time.Time `json:"lastSyncAt"`
}
