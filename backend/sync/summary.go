package sync

import (
    "time"

    "gorm.io/gorm"

    "shosha_mart/backend/models"
)

// Summary aggregates unsynced records and last sync marker.
type Summary struct {
    QueuedChanges int        `json:"queuedChanges"`
    LastSyncAt    *time.Time `json:"lastSyncAt"`
    DbPath        string     `json:"dbPath"`
    Status        string     `json:"status"`
    LastError     string     `json:"lastError,omitempty"`
}

func Build(db *gorm.DB, dbPath, status, lastErr string) (Summary, error) {
    var (
        unsyncedProducts int64
        unsyncedBranches int64
        unsyncedSales    int64
        unsyncedItems    int64
        unsyncedOpname   int64
        unsyncedOpItems  int64
        syncState        models.SyncState
    )

    _ = db.First(&syncState, "id = ?", "singleton").Error

    db.Model(&models.Product{}).Where("synced = ?", false).Count(&unsyncedProducts)
    db.Model(&models.Branch{}).Where("synced = ?", false).Count(&unsyncedBranches)
    db.Model(&models.Sale{}).Where("synced = ?", false).Count(&unsyncedSales)
    db.Model(&models.SaleItem{}).Where("synced = ?", false).Count(&unsyncedItems)
    db.Model(&models.StockOpname{}).Where("synced = ?", false).Count(&unsyncedOpname)
    db.Model(&models.StockOpnameItem{}).Where("synced = ?", false).Count(&unsyncedOpItems)

    total := int(unsyncedProducts + unsyncedBranches + unsyncedSales + unsyncedItems + unsyncedOpname + unsyncedOpItems)

    return Summary{
        QueuedChanges: total,
        LastSyncAt:    syncState.LastSyncAt,
        DbPath:        dbPath,
        Status:        status,
        LastError:     lastErr,
    }, nil
}

// MarkSync updates the last sync timestamp (helper for future sync flows).
func MarkSync(db *gorm.DB, at time.Time) error {
    state := models.SyncState{ID: "singleton"}
    return db.Where(models.SyncState{ID: state.ID}).Assign(models.SyncState{LastSyncAt: &at}).FirstOrCreate(&state).Error
}
