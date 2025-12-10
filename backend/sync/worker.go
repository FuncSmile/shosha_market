package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"shosha_mart/backend/config"
	"shosha_mart/backend/models"
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
		client:   &http.Client{Timeout: 5 * time.Second},
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

	// mark synced
	if len(products) > 0 {
		ids := make([]string, len(products))
		for i, p := range products {
			ids[i] = p.ID
		}
		w.db.Model(&models.Product{}).Where("id IN ?", ids).Update("synced", true)
	}
	if len(branches) > 0 {
		ids := make([]string, len(branches))
		for i, p := range branches {
			ids[i] = p.ID
		}
		w.db.Model(&models.Branch{}).Where("id IN ?", ids).Update("synced", true)
	}
	if len(sales) > 0 {
		ids := make([]string, len(sales))
		for i, p := range sales {
			ids[i] = p.ID
		}
		w.db.Model(&models.Sale{}).Where("id IN ?", ids).Update("synced", true)
	}
	if len(items) > 0 {
		ids := make([]string, len(items))
		for i, p := range items {
			ids[i] = p.ID
		}
		w.db.Model(&models.SaleItem{}).Where("id IN ?", ids).Update("synced", true)
	}
	if len(opnames) > 0 {
		ids := make([]string, len(opnames))
		for i, p := range opnames {
			ids[i] = p.ID
		}
		w.db.Model(&models.StockOpname{}).Where("id IN ?", ids).Update("synced", true)
	}
	if len(opItems) > 0 {
		ids := make([]string, len(opItems))
		for i, p := range opItems {
			ids[i] = p.ID
		}
		w.db.Model(&models.StockOpnameItem{}).Where("id IN ?", ids).Update("synced", true)
	}
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
	if last != "" {
		url += "?since=" + last
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
	// Upsert
	saveOpts := clause.OnConflict{Columns: []clause.Column{{Name: "id"}}, DoUpdates: clause.AssignmentColumns([]string{"name", "address", "phone", "stock", "price", "branch_id", "total", "synced", "updated_at", "created_at"})}
	if len(data.Branches) > 0 {
		w.db.Clauses(saveOpts).Create(&data.Branches)
	}
	if len(data.Products) > 0 {
		w.db.Clauses(saveOpts).Create(&data.Products)
	}
	if len(data.Sales) > 0 {
		w.db.Clauses(saveOpts).Create(&data.Sales)
	}
	if len(data.SaleItems) > 0 {
		w.db.Clauses(saveOpts).Create(&data.SaleItems)
	}
	if len(data.StockOpnames) > 0 {
		w.db.Clauses(saveOpts).Create(&data.StockOpnames)
	}
	if len(data.StockOpnameItems) > 0 {
		w.db.Clauses(saveOpts).Create(&data.StockOpnameItems)
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
