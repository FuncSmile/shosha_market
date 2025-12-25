package models

import "time"

// Product represents an inventory item stored locally first.
type Product struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	Name          string     `json:"name"`
	Unit          string     `json:"unit"` // Satuan (kg, pcs, liter, dll)
	Stock         int        `json:"stock"`
	Price         float64    `json:"price"`          // Legacy/default price used by existing sales logic
	PriceInvestor float64    `json:"price_investor"` // Harga untuk Investor
	PriceShosha   float64    `json:"price_shosha"`   // Harga untuk SHOSHA
	Synced        bool       `json:"synced"`
	BranchID      string     `json:"branch_id"`
	IsDeleted     bool       `json:"is_deleted"`
	DeletedAt     *time.Time `json:"deleted_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// Branch represents a store branch entry.
type Branch struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	Code      string     `json:"code"`
	Name      string     `json:"name"`
	Address   string     `json:"address"`
	Phone     string     `json:"phone"`
	Synced    bool       `json:"synced"`
	IsDeleted bool       `json:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// Sale captures a checkout transaction.
type Sale struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	ReceiptNo     string     `json:"receipt_no"`
	BranchID      string     `json:"branch_id"`
	BranchName    string     `json:"branch_name"`
	PaymentMethod string     `json:"payment_method"` // "cash" or "hutang"
	Notes         string     `json:"notes"`
	Total         float64    `json:"total"`
	Synced        bool       `json:"synced"`
	IsDeleted     bool       `json:"is_deleted"`
	DeletedAt     *time.Time `json:"deleted_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Items         []SaleItem `json:"items"`
}

// SaleItem links to Sale.
type SaleItem struct {
	ID        string     `json:"id" gorm:"primaryKey"`
	SaleID    string     `json:"sale_id"`
	ProductID string     `json:"product_id"`
	Qty       int        `json:"qty"`
	Price     float64    `json:"price"`
	Synced    bool       `json:"synced"`
	IsDeleted bool       `json:"is_deleted"`
	DeletedAt *time.Time `json:"deleted_at"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

// StockOpname represents a stock take session.
type StockOpname struct {
	ID          string            `json:"id" gorm:"primaryKey"`
	BranchID    string            `json:"branch_id"`
	PerformedBy string            `json:"performed_by"`
	Note        string            `json:"note"`
	Synced      bool              `json:"synced"`
	IsDeleted   bool              `json:"is_deleted"`
	DeletedAt   *time.Time        `json:"deleted_at"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	Items       []StockOpnameItem `json:"items"`
}

// StockOpnameItem holds per-product opname figures.
type StockOpnameItem struct {
	ID            string     `json:"id" gorm:"primaryKey"`
	StockOpnameID string     `json:"stock_opname_id"`
	ProductID     string     `json:"product_id"`
	SystemQty     int        `json:"system_qty"`
	PhysicalQty   int        `json:"physical_qty"`
	Synced        bool       `json:"synced"`
	IsDeleted     bool       `json:"is_deleted"`
	DeletedAt     *time.Time `json:"deleted_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// SyncState stores last sync metadata.
type SyncState struct {
	ID         string     `json:"id" gorm:"primaryKey"`
	LastSyncAt *time.Time `json:"last_sync_at"`
}
