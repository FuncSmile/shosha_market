package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"shosha_mart_backend/config"
	"shosha_mart_backend/models"
)

type Worker struct {
	db       *gorm.DB
	cfg      config.AppConfig
	client   *http.Client
	status   string
	lastErr  string
	lastRun  *time.Time
	mu       sync.Mutex
	stopCh   chan struct{}
	interval time.Duration
	inFlight bool
}

func NewWorker(db *gorm.DB, cfg config.AppConfig) *Worker {
	return &Worker{
		db:       db,
		cfg:      cfg,
		client:   &http.Client{Timeout: 30 * time.Second},
		status:   "offline",
		stopCh:   make(chan struct{}),
		interval: 5 * time.Minute,
	}
}

func (w *Worker) Status() (status string, lastErr string, lastRun *time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.status, w.lastErr, w.lastRun
}

// StartBackground starts periodic sync if upstream is configured.
func (w *Worker) StartBackground() {
	if w.cfg.Upstream == "" {
		return
	}
	go func() {
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = w.RunOnce(context.Background())
			case <-w.stopCh:
				return
			}
		}
	}()
}

func (w *Worker) Stop() {
	close(w.stopCh)
}

// RunOnce uploads unsynced rows and pulls changes from upstream.
func (w *Worker) RunOnce(ctx context.Context) error {
	log.Printf("[SYNC] RunOnce using DB path: %s", w.cfg.DBPath)
	w.mu.Lock()
	if w.inFlight {
		w.mu.Unlock()
		return errors.New("sync already running")
	}
	w.inFlight = true
	w.mu.Unlock()
	defer func() {
		w.mu.Lock()
		w.inFlight = false
		w.mu.Unlock()
	}()

	if w.cfg.Upstream == "" {
		w.setStatus("offline", "upstream not configured", nil)
		return errors.New("upstream not configured")
	}

	// Upload
	if err := w.upload(ctx); err != nil {
		w.setStatus("offline", err.Error(), nil)
		return err
	}

	// Download
	if err := w.download(ctx); err != nil {
		w.setStatus("offline", err.Error(), nil)
		return err
	}

	now := time.Now()
	_ = MarkSync(w.db, now)
	w.setStatus("online", "", &now)
	return nil
}

func (w *Worker) upload(ctx context.Context) error {
	var (
		products []models.Product
		branches []models.Branch
		sales    []models.Sale
		items    []models.SaleItem
		opnames  []models.StockOpname
		opItems  []models.StockOpnameItem
	)
	w.db.Where("synced = ?", false).Find(&products)
	w.db.Where("synced = ?", false).Find(&branches)
	w.db.Where("synced = ?", false).Find(&sales)
	w.db.Where("synced = ?", false).Find(&items)
	w.db.Where("synced = ?", false).Find(&opnames)
	w.db.Where("synced = ?", false).Find(&opItems)

	payload := map[string]any{
		"branch_id":          w.cfg.BranchID,
		"products":           products,
		"branches":           branches,
		"sales":              sales,
		"sale_items":         items,
		"stock_opnames":      opnames,
		"stock_opname_items": opItems,
	}
	body, _ := json.Marshal(payload)

	url := strings.TrimSuffix(w.cfg.Upstream, "/") + "/api/sync/upload"
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("upload: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("upload failed: %s", resp.Status)
	}

	// mark synced (log RowsAffected and errors for debugging)
	if len(products) > 0 {
		ids := make([]string, len(products))
		for i, p := range products {
			ids[i] = p.ID
		}
		res := w.db.Model(&models.Product{}).Where("id IN ?", ids).Update("synced", true)
		log.Printf("[SYNC] marked products synced: rows=%d, error=%v", res.RowsAffected, res.Error)
	}
	if len(branches) > 0 {
		ids := make([]string, len(branches))
		for i, p := range branches {
			ids[i] = p.ID
		}
		res := w.db.Model(&models.Branch{}).Where("id IN ?", ids).Update("synced", true)
		log.Printf("[SYNC] marked branches synced: rows=%d, error=%v", res.RowsAffected, res.Error)
	}
	if len(sales) > 0 {
		ids := make([]string, len(sales))
		for i, p := range sales {
			ids[i] = p.ID
		}
		res := w.db.Model(&models.Sale{}).Where("id IN ?", ids).Update("synced", true)
		log.Printf("[SYNC] marked sales synced: rows=%d, error=%v", res.RowsAffected, res.Error)
	}
	if len(items) > 0 {
		ids := make([]string, len(items))
		for i, p := range items {
			ids[i] = p.ID
		}
		res := w.db.Model(&models.SaleItem{}).Where("id IN ?", ids).Update("synced", true)
		log.Printf("[SYNC] marked sale_items synced: rows=%d, error=%v", res.RowsAffected, res.Error)
	}
	if len(opnames) > 0 {
		ids := make([]string, len(opnames))
		for i, p := range opnames {
			ids[i] = p.ID
		}
		res := w.db.Model(&models.StockOpname{}).Where("id IN ?", ids).Update("synced", true)
		log.Printf("[SYNC] marked stock_opnames synced: rows=%d, error=%v", res.RowsAffected, res.Error)
	}
	if len(opItems) > 0 {
		ids := make([]string, len(opItems))
		for i, p := range opItems {
			ids[i] = p.ID
		}
		res := w.db.Model(&models.StockOpnameItem{}).Where("id IN ?", ids).Update("synced", true)
		log.Printf("[SYNC] marked stock_opname_items synced: rows=%d, error=%v", res.RowsAffected, res.Error)
	}

	// Prune locally: hard delete rows that are tombstoned and synced
	w.db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.SaleItem{})
	w.db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.Sale{})
	w.db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.Product{})
	w.db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.Branch{})
	w.db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.StockOpnameItem{})
	w.db.Where("is_deleted = ? AND synced = ?", true, true).Delete(&models.StockOpname{})
	return nil
}

func (w *Worker) download(ctx context.Context) error {
	var syncState models.SyncState
	_ = w.db.First(&syncState, "id = ?", "singleton").Error
	last := ""
	if syncState.LastSyncAt != nil {
		last = syncState.LastSyncAt.UTC().Format(time.RFC3339)
	}
	url := strings.TrimSuffix(w.cfg.Upstream, "/") + "/api/sync/changes"
	// build query params
	qs := ""
	if last != "" {
		qs = "?since=" + last
	}
	if w.cfg.BranchID != "" {
		if qs == "" {
			qs = "?branch_id=" + w.cfg.BranchID
		} else {
			qs += "&branch_id=" + w.cfg.BranchID
		}
	}
	if qs != "" {
		url += qs
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("download: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		// upstream not ready
		return nil
	}
	if resp.StatusCode >= 300 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}
	var data struct {
		Products         []models.Product         `json:"products"`
		Branches         []models.Branch          `json:"branches"`
		Sales            []models.Sale            `json:"sales"`
		SaleItems        []models.SaleItem        `json:"sale_items"`
		StockOpnames     []models.StockOpname     `json:"stock_opnames"`
		StockOpnameItems []models.StockOpnameItem `json:"stock_opname_items"`
		LastSyncAt       *time.Time               `json:"last_sync_at"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return fmt.Errorf("decode changes: %w", err)
	}
	// Upsert: gunakan opsi berbeda per model agar tidak merujuk kolom yang tidak ada
	saveOptsBranches := clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"code", "name", "address", "phone", "synced", "updated_at", "created_at"})}
	saveOptsProducts := clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "unit", "stock", "price", "branch_id", "synced", "updated_at", "created_at"})}
	saveOptsSales := clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"receipt_no", "branch_id", "branch_name", "payment_method", "notes", "total", "synced", "updated_at", "created_at"})}
	saveOptsSaleItems := clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"sale_id", "product_id", "qty", "price", "synced", "updated_at", "created_at"})}
	saveOptsOpnames := clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"branch_id", "performed_by", "note", "synced", "updated_at", "created_at"})}
	saveOptsOpItems := clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"stock_opname_id", "product_id", "system_qty", "physical_qty", "synced", "updated_at", "created_at"})}
	// Set synced=true untuk semua data hasil download
	for i := range data.Branches {
		data.Branches[i].Synced = true
	}
	for i := range data.Products {
		data.Products[i].Synced = true
	}
	for i := range data.Sales {
		data.Sales[i].Synced = true
	}
	for i := range data.SaleItems {
		data.SaleItems[i].Synced = true
	}
	for i := range data.StockOpnames {
		data.StockOpnames[i].Synced = true
	}
	for i := range data.StockOpnameItems {
		data.StockOpnameItems[i].Synced = true
	}
	if len(data.Branches) > 0 {
		res := w.db.Clauses(saveOptsBranches).Create(&data.Branches)
		log.Printf("[SYNC] downloaded branches: %d, error: %v", len(data.Branches), res.Error)
	}
	if len(data.Products) > 0 {
		res := w.db.Clauses(saveOptsProducts).Create(&data.Products)
		log.Printf("[SYNC] downloaded products: %d, error: %v", len(data.Products), res.Error)
	}
	if len(data.Sales) > 0 {
		res := w.db.Clauses(saveOptsSales).Create(&data.Sales)
		log.Printf("[SYNC] downloaded sales: %d, error: %v", len(data.Sales), res.Error)
	}
	if len(data.SaleItems) > 0 {
		res := w.db.Clauses(saveOptsSaleItems).Create(&data.SaleItems)
		log.Printf("[SYNC] downloaded sale_items: %d, error: %v", len(data.SaleItems), res.Error)
	}
	if len(data.StockOpnames) > 0 {
		res := w.db.Clauses(saveOptsOpnames).Create(&data.StockOpnames)
		log.Printf("[SYNC] downloaded stock_opnames: %d, error: %v", len(data.StockOpnames), res.Error)
	}
	if len(data.StockOpnameItems) > 0 {
		res := w.db.Clauses(saveOptsOpItems).Create(&data.StockOpnameItems)
		log.Printf("[SYNC] downloaded stock_opname_items: %d, error: %v", len(data.StockOpnameItems), res.Error)
	}

	if data.LastSyncAt != nil {
		_ = MarkSync(w.db, *data.LastSyncAt)
	}
	return nil
}

func (w *Worker) setStatus(status, errMsg string, ts *time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.status = status
	w.lastErr = errMsg
	if ts != nil {
		w.lastRun = ts
	}
}
