package exports

import (
	"fmt"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"

	"shosha_mart_backend/models"
)

// ExportSalesByBranch creates an Excel file with one sheet per branch containing sales and items
func ExportSalesByBranch(db *gorm.DB) (*excelize.File, error) {
	f := excelize.NewFile()
	defer f.Close()

	// Get all sales with items
	var sales []models.Sale
	if err := db.Order("created_at DESC").Find(&sales).Error; err != nil {
		return nil, err
	}

	// Load all sale items in one query for efficiency
	var allItems []models.SaleItem
	if err := db.Order("sale_id, created_at").Find(&allItems).Error; err != nil {
		return nil, err
	}

	// Map sale_id -> items
	itemsBySale := make(map[string][]models.SaleItem)
	for _, item := range allItems {
		itemsBySale[item.SaleID] = append(itemsBySale[item.SaleID], item)
	}

	// Load all products for name lookup
	var products []models.Product
	if err := db.Find(&products).Error; err != nil {
		return nil, err
	}
	productMap := make(map[string]models.Product)
	for _, p := range products {
		productMap[p.ID] = p
	}

	// Group sales by branch
	salesByBranch := make(map[string][]models.Sale)
	for _, sale := range sales {
		branchKey := sale.BranchName
		if branchKey == "" {
			branchKey = "Tanpa Cabang"
		}
		salesByBranch[branchKey] = append(salesByBranch[branchKey], sale)
	}

	// Create a sheet for each branch
	sheetIndex := 0
	for branchName, branchSales := range salesByBranch {
		var sheetName string
		if sheetIndex == 0 {
			// Rename default Sheet1
			sheetName = sanitizeSheetName(branchName)
			f.SetSheetName("Sheet1", sheetName)
		} else {
			sheetName = sanitizeSheetName(branchName)
			_, err := f.NewSheet(sheetName)
			if err != nil {
				return nil, err
			}
		}

		// Write header
		headers := []string{"Tanggal", "No. Invoice", "Metode", "Total", "Nama Barang", "Qty", "Harga", "Jumlah"}
		for col, h := range headers {
			cell, _ := excelize.CoordinatesToCellName(col+1, 1)
			f.SetCellValue(sheetName, cell, h)
		}

		row := 2
		for _, sale := range branchSales {
			items := itemsBySale[sale.ID]
			if len(items) == 0 {
				// Empty sale, still show header info
				dateStr := sale.CreatedAt.Format("02 Jan 2006")
				f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), dateStr)
				f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), sale.ReceiptNo)
				f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), sale.PaymentMethod)
				f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), sale.Total)
				row++
			} else {
				for idx, item := range items {
					dateStr := sale.CreatedAt.Format("02 Jan 2006")
					productName := productMap[item.ProductID].Name
					if productName == "" {
						productName = item.ProductID
					}
					subtotal := float64(item.Qty) * item.Price

					if idx == 0 {
						// First row shows sale header
						f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), dateStr)
						f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), sale.ReceiptNo)
						f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), sale.PaymentMethod)
						f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), sale.Total)
					}
					// Item details
					f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), productName)
					f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), item.Qty)
					f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.Price)
					f.SetCellValue(sheetName, fmt.Sprintf("H%d", row), subtotal)
					row++
				}
			}
		}

		sheetIndex++
	}

	return f, nil
}

func sanitizeSheetName(name string) string {
	// Excel sheet names must be <= 31 chars and cannot contain: \ / ? * [ ]
	if len(name) > 31 {
		name = name[:31]
	}
	invalid := []rune{'\\', '/', '?', '*', '[', ']'}
	runes := []rune(name)
	for i, r := range runes {
		for _, inv := range invalid {
			if r == inv {
				runes[i] = '_'
				break
			}
		}
	}
	return string(runes)
}
