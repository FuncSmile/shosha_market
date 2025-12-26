package reports

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"shosha_mart_backend/config"
	"shosha_mart_backend/models"
)

// GenerateSalesReport builds an Excel export for sales between start/end.
func GenerateSalesReport(db *gorm.DB, cfg config.AppConfig, start, end time.Time) (string, error) {
	var sales []models.Sale
	if err := db.Where("created_at BETWEEN ? AND ?", start, end).Preload("Items").Find(&sales).Error; err != nil {
		return "", err
	}

	if err := os.MkdirAll(cfg.ExportDir, 0o755); err != nil {
		return "", err
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)

	title := fmt.Sprintf("Laporan Penjualan %s - %s", start.Format("02 Jan 2006"), end.Format("02 Jan 2006"))
	f.SetCellValue(sheet, "A1", "POS Offline-First")
	f.SetCellValue(sheet, "A2", title)
	f.SetCellValue(sheet, "A3", fmt.Sprintf("Cabang: %s", cfg.BranchID))

	headers := []string{"No", "Tanggal", "No Nota", "Total", "Items"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 5)
		f.SetCellValue(sheet, cell, h)
	}

	row := 6
	for idx, sale := range sales {
		dateCell, _ := excelize.CoordinatesToCellName(2, row)
		noCell, _ := excelize.CoordinatesToCellName(1, row)
		receiptCell, _ := excelize.CoordinatesToCellName(3, row)
		totalCell, _ := excelize.CoordinatesToCellName(4, row)
		itemsCell, _ := excelize.CoordinatesToCellName(5, row)

		f.SetCellValue(sheet, noCell, idx+1)
		f.SetCellValue(sheet, dateCell, sale.CreatedAt.Format("02-01-2006 15:04"))
		f.SetCellValue(sheet, receiptCell, sale.ReceiptNo)
		f.SetCellValue(sheet, totalCell, sale.Total)
		f.SetCellValue(sheet, itemsCell, len(sale.Items))
		row++
	}

	filename := fmt.Sprintf("sales_%s_%s.xlsx", start.Format("20060102"), end.Format("20060102"))
	path := filepath.Join(cfg.ExportDir, filename)
	if err := f.SaveAs(path); err != nil {
		return "", err
	}
	return path, nil
}

// GenerateSalesReportByBranch builds an Excel export for a single branch, grouped by date.
func GenerateSalesReportByBranch(db *gorm.DB, cfg config.AppConfig, branchID string, start, end time.Time) (string, error) {
	var sales []models.Sale
	query := db.Where("created_at BETWEEN ? AND ? AND is_deleted = false", start, end)
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	if err := query.Preload("Items").Order("created_at ASC").Find(&sales).Error; err != nil {
		return "", err
	}

	if err := os.MkdirAll(cfg.ExportDir, 0o755); err != nil {
		return "", err
	}

	f := excelize.NewFile()
	sheet := f.GetSheetName(0)
	f.SetSheetName(sheet, "Penjualan")

	// Group sales by date
	grouped := groupSalesByDate(sales)
	
	branchName := branchID
	if branchID != "" {
		var branch models.Branch
		if err := db.First(&branch, "id = ?", branchID).Error; err == nil {
			branchName = branch.Name
		}
	}

	if err := writeSalesSheet(f, "Penjualan", branchName, start, end, grouped); err != nil {
		return "", err
	}

	filename := fmt.Sprintf("sales_branch_%s_%s_%s.xlsx", branchID, start.Format("20060102"), end.Format("20060102"))
	path := filepath.Join(cfg.ExportDir, filename)
	if err := f.SaveAs(path); err != nil {
		return "", err
	}
	return path, nil
}

// GenerateSalesReportGlobal builds a multi-sheet Excel with one sheet per branch, grouped by date.
func GenerateSalesReportGlobal(db *gorm.DB, cfg config.AppConfig, start, end time.Time) (string, error) {
	var sales []models.Sale
	if err := db.Where("created_at BETWEEN ? AND ? AND is_deleted = false", start, end).
		Preload("Items").
		Order("branch_id ASC, created_at ASC").
		Find(&sales).Error; err != nil {
		return "", err
	}

	if err := os.MkdirAll(cfg.ExportDir, 0o755); err != nil {
		return "", err
	}

	// Group sales by branch
	branchSales := make(map[string][]models.Sale)
	for _, s := range sales {
		branchSales[s.BranchID] = append(branchSales[s.BranchID], s)
	}

	// Get branch names
	var branches []models.Branch
	db.Find(&branches)
	branchNames := make(map[string]string)
	for _, b := range branches {
		branchNames[b.ID] = b.Name
	}

	f := excelize.NewFile()
	f.DeleteSheet("Sheet1")

	// Sort branch IDs for consistent ordering
	branchIDs := make([]string, 0, len(branchSales))
	for bid := range branchSales {
		branchIDs = append(branchIDs, bid)
	}
	sort.Strings(branchIDs)

	// Create one sheet per branch
	for _, branchID := range branchIDs {
		salesForBranch := branchSales[branchID]
		branchName := branchNames[branchID]
		if branchName == "" {
			branchName = branchID
		}

		// Sanitize sheet name (Excel limits: 31 chars, no special chars)
		sheetName := sanitizeSheetName(branchName)
		f.NewSheet(sheetName)

		grouped := groupSalesByDate(salesForBranch)
		if err := writeSalesSheet(f, sheetName, branchName, start, end, grouped); err != nil {
			return "", err
		}
	}

	filename := fmt.Sprintf("sales_global_%s_%s.xlsx", start.Format("20060102"), end.Format("20060102"))
	path := filepath.Join(cfg.ExportDir, filename)
	if err := f.SaveAs(path); err != nil {
		return "", err
	}
	return path, nil
}

// groupSalesByDate groups sales by date (YYYY-MM-DD format)
func groupSalesByDate(sales []models.Sale) map[string][]models.Sale {
	grouped := make(map[string][]models.Sale)
	for _, s := range sales {
		dateKey := s.CreatedAt.Format("2006-01-02")
		grouped[dateKey] = append(grouped[dateKey], s)
	}
	return grouped
}

// writeSalesSheet writes a sales report to a specific sheet with grouping by date and proper column widths
func writeSalesSheet(f *excelize.File, sheetName, branchName string, start, end time.Time, grouped map[string][]models.Sale) error {
	// Header
	title := fmt.Sprintf("Laporan Penjualan %s - %s", start.Format("02 Jan 2006"), end.Format("02 Jan 2006"))
	f.SetCellValue(sheetName, "A1", "POS Offline-First")
	f.SetCellValue(sheetName, "A2", title)
	f.SetCellValue(sheetName, "A3", fmt.Sprintf("Cabang: %s", branchName))

	// Table headers
	headers := []string{"Tanggal", "No", "No Nota", "Metode Bayar", "Total", "Jumlah Item"}
	headerRow := 5
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		f.SetCellValue(sheetName, cell, h)
	}

	// Sort dates
	dates := make([]string, 0, len(grouped))
	for date := range grouped {
		dates = append(dates, date)
	}
	sort.Strings(dates)

	row := headerRow + 1
	for _, date := range dates {
		salesForDate := grouped[date]
		
		for idx, sale := range salesForDate {
			// Date (only on first row of each date group)
			if idx == 0 {
				dateCell, _ := excelize.CoordinatesToCellName(1, row)
				f.SetCellValue(sheetName, dateCell, date)
			}

			noCell, _ := excelize.CoordinatesToCellName(2, row)
			receiptCell, _ := excelize.CoordinatesToCellName(3, row)
			paymentCell, _ := excelize.CoordinatesToCellName(4, row)
			totalCell, _ := excelize.CoordinatesToCellName(5, row)
			itemsCell, _ := excelize.CoordinatesToCellName(6, row)

			f.SetCellValue(sheetName, noCell, idx+1)
			f.SetCellValue(sheetName, receiptCell, sale.ReceiptNo)
			f.SetCellValue(sheetName, paymentCell, sale.PaymentMethod)
			f.SetCellValue(sheetName, totalCell, sale.Total)
			f.SetCellValue(sheetName, itemsCell, len(sale.Items))
			row++
		}

		// Add subtotal for this date
		subtotalRow := row
		dateCell, _ := excelize.CoordinatesToCellName(1, subtotalRow)
		labelCell, _ := excelize.CoordinatesToCellName(3, subtotalRow)
		subtotalCell, _ := excelize.CoordinatesToCellName(5, subtotalRow)

		f.SetCellValue(sheetName, dateCell, "")
		f.SetCellValue(sheetName, labelCell, "Subtotal")
		
		subtotal := 0.0
		for _, s := range salesForDate {
			subtotal += s.Total
		}
		f.SetCellValue(sheetName, subtotalCell, subtotal)
		row++
	}

	// Set column widths
	f.SetColWidth(sheetName, "A", "A", 12) // Tanggal
	f.SetColWidth(sheetName, "B", "B", 6)  // No
	f.SetColWidth(sheetName, "C", "C", 15) // No Nota
	f.SetColWidth(sheetName, "D", "D", 14) // Metode Bayar
	f.SetColWidth(sheetName, "E", "E", 12) // Total
	f.SetColWidth(sheetName, "F", "F", 12) // Jumlah Item

	return nil
}

// sanitizeSheetName ensures sheet name is valid for Excel (max 31 chars, no special chars)
func sanitizeSheetName(name string) string {
	// Replace invalid characters
	invalid := []string{"\\", "/", "*", "[", "]", ":", "?"}
	for _, char := range invalid {
		name = replaceAll(name, char, "_")
	}
	
	// Limit to 31 characters
	if len(name) > 31 {
		name = name[:31]
	}
	
	return name
}

func replaceAll(s, old, new string) string {
	result := ""
	for _, c := range s {
		if string(c) == old {
			result += new
		} else {
			result += string(c)
		}
	}
	return result
}
