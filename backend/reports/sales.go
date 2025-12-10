package reports

import (
    "fmt"
    "os"
    "path/filepath"
    "time"

    "github.com/xuri/excelize/v2"
    "gorm.io/gorm"

    "shosha_mart/backend/config"
    "shosha_mart/backend/models"
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
