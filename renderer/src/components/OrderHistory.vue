<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, type Sale, type SaleItem, type Product, type Branch } from '../api'
import { toast } from 'vue-sonner'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Label from './ui/Label.vue'
import Input from './ui/Input.vue'
import Select from './ui/Select.vue'
import BranchSelect from './ui/BranchSelect.vue'
import Accordion from './ui/Accordion.vue'
import AccordionItem from './ui/AccordionItem.vue'
import AccordionTrigger from './ui/AccordionTrigger.vue'
import AccordionContent from './ui/AccordionContent.vue'

const sales = ref<Sale[]>([])
const loading = ref(false)
const error = ref('')
const exporting = ref(false)

const showDetail = ref(false)
const selected = ref<Sale | null>(null)
type AugmentedItem = SaleItem & { name?: string; unit?: string; subtotal?: number }
const items = ref<AugmentedItem[]>([])

const showEditDialog = ref(false)
const editingSale = ref<Sale | null>(null)
const editForm = ref<{
  created_at: string
  branch_id: string
  payment_method: string
  notes: string
}>({
  created_at: '',
  branch_id: '',
  payment_method: 'cash',
  notes: ''
})
const savingEdit = ref(false)

const products = ref<Product[]>([])
const branches = ref<Branch[]>([])

const searchCode = ref('')

const showPrintDialog = ref(false)
const printData = ref<any>(null)

const dateRange = ref<{ start: string; end: string }>({ start: '', end: '' })

// Products edit dialog state
const showProductsEditDialog = ref(false)
const editingProductsForSale = ref<Sale | null>(null)
const editingItems = ref<(SaleItem & { name?: string; unit?: string })[]>([])
const newItemForm = ref<{ product_id: string; qty: number; price: number }>({
  product_id: '',
  qty: 1,
  price: 0
})
const savingProducts = ref(false)


const branchMap = computed(() => Object.fromEntries(branches.value.map(b => [b.id, b])))

const displayedSales = computed(() => {
  const start = dateRange.value.start ? new Date(dateRange.value.start) : null
  const end = dateRange.value.end ? new Date(dateRange.value.end) : null
  const codeQuery = searchCode.value.trim().toLowerCase()

  return sales.value
    .map(s => ({
      ...s,
      branch_code: branchMap.value[s.branch_id]?.code || '',
      branch_name_resolved: branchMap.value[s.branch_id]?.name || s.branch_name || s.branch_id,
    }))
    .filter(s => {
      const d = new Date(s.created_at)
      if (start && d < start) return false
      if (end) {
        // include the end date through 23:59:59
        const endInclusive = new Date(end)
        endInclusive.setHours(23, 59, 59, 999)
        if (d > endInclusive) return false
      }
      if (codeQuery) {
        const code = (s.branch_code || '').toLowerCase()
        if (!code.includes(codeQuery)) return false
      }
      return true
    })
})

// Group sales by branch
const groupedByBranch = computed(() => {
  const groups = new Map<string, typeof displayedSales.value>()
  
  displayedSales.value.forEach(sale => {
    const branchKey = sale.branch_name_resolved || 'Unknown'
    if (!groups.has(branchKey)) {
      groups.set(branchKey, [])
    }
    groups.get(branchKey)!.push(sale)
  })
  
  // Convert to array and sort by branch name
  return Array.from(groups.entries())
    .map(([branchName, salesList]) => {
      const total = salesList.reduce((sum, s) => sum + (s.total || 0), 0)
      const branchCode = salesList[0]?.branch_code || ''
      return {
        branchName,
        branchCode,
        sales: salesList,
        total,
        count: salesList.length
      }
    })
    .sort((a, b) => a.branchName.localeCompare(b.branchName))
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    sales.value = await api.listSales()
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    loading.value = false
  }
}

function setDefaultMonth() {
  const now = new Date()
  const start = new Date(now.getFullYear(), now.getMonth(), 1)
  const end = new Date(now.getFullYear(), now.getMonth() + 1, 0)
  const toLocalISO = (d: Date) => {
    const year = d.getFullYear()
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }
  dateRange.value = { start: toLocalISO(start), end: toLocalISO(end) }
}

function setDefaultYear() {
  const now = new Date()
  const start = new Date(now.getFullYear(), 0, 1) // 1 Januari tahun ini
  const end = new Date(now.getFullYear(), 11, 31) // 31 Desember tahun ini
  const toLocalISO = (d: Date) => {
    const year = d.getFullYear()
    const month = String(d.getMonth() + 1).padStart(2, '0')
    const day = String(d.getDate()).padStart(2, '0')
    return `${year}-${month}-${day}`
  }
  dateRange.value = { start: toLocalISO(start), end: toLocalISO(end) }
}

function clearDateFilter() {
  dateRange.value = { start: '', end: '' }
}

async function openDetail(id: string) {
  try {
    const sale = await api.getSale(id)
    selected.value = sale
    const productMap = Object.fromEntries(products.value.map(p => [p.id, p]))
    items.value = (sale.items || []).map(i => ({
      ...i,
      name: productMap[i.product_id]?.name || i.product_id,
      unit: productMap[i.product_id]?.unit || 'PCS',
      subtotal: i.qty * i.price,
    }))
    showDetail.value = true
  } catch (e) {
    error.value = (e as Error).message
  }
}

function closeDetail() {
  showDetail.value = false
  selected.value = null
  items.value = []
}

function openEditDialog(sale: Sale) {
  editingSale.value = sale
  editForm.value = {
    created_at: sale.created_at.slice(0, 10), // Format to YYYY-MM-DD
    branch_id: sale.branch_id,
    payment_method: sale.payment_method,
    notes: sale.notes || ''
  }
  showEditDialog.value = true
}

function closeEditDialog() {
  showEditDialog.value = false
  editingSale.value = null
  editForm.value = {
    created_at: '',
    branch_id: '',
    payment_method: 'cash',
    notes: ''
  }
}

async function saveEdit() {
  if (!editingSale.value) return
  
  savingEdit.value = true
  try {
    await api.updateSale(editingSale.value.id, {
      created_at: editForm.value.created_at,
      branch_id: editForm.value.branch_id,
      payment_method: editForm.value.payment_method,
      notes: editForm.value.notes
    })
    
    toast.success('Transaksi berhasil diupdate')
    closeEditDialog()
    await load() // Reload data
  } catch (err) {
    toast.error((err as Error).message)
  } finally {
    savingEdit.value = false
  }
}

// Products edit dialog functions
function openProductsEditDialog(sale: Sale) {
  editingProductsForSale.value = sale
  const productMap = Object.fromEntries(products.value.map(p => [p.id, p]))
  editingItems.value = sale.items.map(item => {
    const product = productMap[item.product_id]
    return {
      ...item,
      name: product?.name,
      unit: product?.unit
    }
  })
  newItemForm.value = { product_id: '', qty: 1, price: 0 }
  showProductsEditDialog.value = true
}

function closeProductsEditDialog() {
  showProductsEditDialog.value = false
  editingProductsForSale.value = null
  editingItems.value = []
  newItemForm.value = { product_id: '', qty: 1, price: 0 }
}

async function addProductToSale() {
  if (!editingProductsForSale.value || !newItemForm.value.product_id || newItemForm.value.qty <= 0) {
    toast.error('Pilih produk dan masukkan qty > 0')
    return
  }

  savingProducts.value = true
  try {
    const product = products.value.find(p => p.id === newItemForm.value.product_id)
    if (!product) {
      toast.error('Produk tidak ditemukan')
      return
    }

    const price = newItemForm.value.price > 0 ? newItemForm.value.price : product.price
    
    const newItem = await api.addSaleItem(editingProductsForSale.value.id, {
      product_id: newItemForm.value.product_id,
      qty: newItemForm.value.qty,
      price: price
    })

    toast.success('Produk berhasil ditambahkan')
    // Update editing items list
    editingItems.value.push({
      ...newItem,
      name: product.name,
      unit: product.unit
    })
    newItemForm.value = { product_id: '', qty: 1, price: 0 }
  } catch (err) {
    toast.error((err as Error).message)
  } finally {
    savingProducts.value = false
  }
}

async function updateProductInSale(item: SaleItem, newQty: number, newPrice: number) {
  if (!editingProductsForSale.value || newQty <= 0 || newPrice <= 0) {
    toast.error('Qty dan price harus > 0')
    return
  }

  savingProducts.value = true
  try {
    await api.updateSaleItem(editingProductsForSale.value.id, item.id, {
      qty: newQty,
      price: newPrice
    })
    
    toast.success('Produk berhasil diupdate')
    
    // Update in local list
    const idx = editingItems.value.findIndex(i => i.id === item.id)
    if (idx >= 0) {
      editingItems.value[idx].qty = newQty
      editingItems.value[idx].price = newPrice
    }
  } catch (err) {
    toast.error((err as Error).message)
  } finally {
    savingProducts.value = false
  }
}

async function deleteProductFromSale(itemId: string) {
  if (!editingProductsForSale.value) return

  if (!confirm('Yakin hapus produk ini?')) return

  savingProducts.value = true
  try {
    await api.deleteSaleItem(editingProductsForSale.value.id, itemId)
    
    toast.success('Produk berhasil dihapus')
    editingItems.value = editingItems.value.filter(i => i.id !== itemId)
    
    // Reload full data to get updated total
    await load()
  } catch (err) {
    toast.error((err as Error).message)
  } finally {
    savingProducts.value = false
  }
}

async function closeAndReloadProducts() {
  closeProductsEditDialog()
  await load()
}


async function exportData() {
  exporting.value = true
  error.value = ''
  try {
    const xlsx = await import('xlsx')

    // Fetch detailed items for the current filtered/sorted sales
    const details = await Promise.all(
      displayedSales.value.map(async (s) => {
        const detail = await api.getSale(s.id)
        return { sale: s, detail }
      })
    )

    const productMap = Object.fromEntries(products.value.map(p => [p.id, p]))

    const headers = ['Tanggal', 'Cabang', 'Nama Barang', 'Qty', 'Harga', 'Subtotal']
    const rows: (string | number)[][] = [headers]

    details.forEach(({ sale, detail }) => {
      const branchName = sale.branch_name_resolved || sale.branch_name || sale.branch_id || '-'
      const dateStr = fmtDate(sale.created_at)
      const totalInvoice = detail.total ?? sale.total ?? 0
      const itemsList = detail.items || []

      itemsList.forEach((item: any, idx: number) => {
        const subtotal = (item.qty || 0) * (item.price || 0)
        const productName = productMap[item.product_id]?.name || item.name || item.product_id || '-'
        rows.push([
          idx === 0 ? dateStr : '',
          idx === 0 ? branchName : '',
          productName,
          item.qty || 0,
          item.price || 0,
          subtotal,
        ])
      })

      // Add total row after all items
      rows.push(['', '', 'TOTAL', '', '', totalInvoice])
      // Add blank row for spacing
      rows.push([])
    })

    const ws = xlsx.utils.aoa_to_sheet(rows)

    // Column widths for readability
    ws['!cols'] = [
      { wch: 12 }, // Tanggal
      { wch: 18 }, // Cabang
      { wch: 28 }, // Nama Barang
      { wch: 8 },  // Qty
      { wch: 12 }, // Harga
      { wch: 14 }, // Subtotal
    ]

    // Number formats and styling
    const range = xlsx.utils.decode_range(ws['!ref'] || `A1:F${rows.length}`)
    for (let r = 1; r <= range.e.r; r++) {
      const qtyCell = xlsx.utils.encode_cell({ r, c: 3 })
      const priceCell = xlsx.utils.encode_cell({ r, c: 4 })
      const subCell = xlsx.utils.encode_cell({ r, c: 5 })
      const nameCell = xlsx.utils.encode_cell({ r, c: 2 })

      // Bold TOTAL rows
      if (ws[nameCell]?.v === 'TOTAL') {
        if (ws[nameCell]) ws[nameCell].s = { font: { bold: true } }
        if (ws[subCell]) { ws[subCell].s = { font: { bold: true }, numFmt: '#,##0' }; ws[subCell].t = 'n' }
      } else {
        if (ws[qtyCell]) ws[qtyCell].t = 'n'
        if (ws[priceCell]) { ws[priceCell].t = 'n'; ws[priceCell].z = '#,##0' }
        if (ws[subCell]) { ws[subCell].t = 'n'; ws[subCell].z = '#,##0' }
      }
    }

    const wb = xlsx.utils.book_new()
    xlsx.utils.book_append_sheet(wb, ws, 'History')
    const wbout = xlsx.write(wb, { bookType: 'xlsx', type: 'array' })
    const blob = new Blob([wbout], { type: 'application/octet-stream' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `history-order-${new Date().toISOString().slice(0, 10)}.xlsx`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
  } catch (e) {
    error.value = (e as Error).message
  } finally {
    exporting.value = false
  }
}

function preparePrint() {
  if (!selected.value) return
  const branchMap = Object.fromEntries(branches.value.map(b => [b.id, b]))
  printData.value = {
    ...selected.value,
    branch: branchMap[selected.value.branch_id] || { name: selected.value.branch_name, address: '' },
    items: items.value,
  }
  showPrintDialog.value = true
  printReceipt()
}

function printReceipt() {
  const printWindow = window.open('', '_blank')
  if (!printWindow || !printData.value) return

  const txDate = printData.value.created_at ? new Date(printData.value.created_at) : new Date()
  const date = txDate.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })


  let itemsHtml = ''
  printData.value.items.forEach((item: any, idx: number) => {
    itemsHtml += `
      <tr>
        <td style="text-align: center; padding: 3px 2px;">${idx + 1}</td>
        <td style="padding: 3px 2px;">${item.name || item.product_id}</td>
        <td style="text-align: center; padding: 3px 2px;">${item.qty}</td>
        <td style="text-align: center; padding: 3px 2px;">${item.unit || 'PCS'}</td>
        <td style="text-align: right; padding: 3px 2px;">Rp ${Number(item.price || 0).toLocaleString('id-ID')}</td>
        <td style="text-align: right; padding: 3px 2px;">Rp ${Number(item.subtotal || 0).toLocaleString('id-ID')}</td>
        <td style="padding: 3px 2px;"></td>
      </tr>
    `
  })

  const grandTotal = printData.value.items.reduce((sum: number, item: any) => sum + (item.subtotal || 0), 0)
  const isHutang = printData.value.payment_method === 'hutang'
  const cashierName = ''


  const html = `
     <!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8" />
    <title>${isHutang ? 'Surat Jalan' : 'Struk Pembayaran'}</title>
    <style>
      * {
        margin: 0;
        padding: 0;
        box-sizing: border-box;
      }
      @page {
        size: A4;
        margin: 12mm;
      }
      body {
        font-family: Arial, Helvetica, sans-serif;
        font-size: 9pt;
        padding: 8mm;
        color: #000;
      }
      .container-head {
        display: flex;
        justify-content: space-between;
      }
      .header {
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: flex-start;
        margin-bottom: 6px;
        width: 40%;
      }
      .header-cust {
        display: flex;
        flex-direction: column;
        justify-content: center;
        align-items: flex-end;
        margin-bottom: 6px;
        width: 40%;
      }
      .title {
        font-size: 20pt;
        font-weight: 800;
        text-align: center;
      }
        .title-cust {
            font-size: 14pt;
            font-weight: 700;
            margin-top: 4px;
        }
      table.items {
        width: 100%;
        border-collapse: collapse;
        margin-top: 8px;
        border-top: 1px solid #000;
        border-bottom: 1px solid #000;
      }
      table.items th {
        padding: 4px 6px;
        border-left: 1px solid #000;
        border-right: 1px solid #000;
        border-bottom: 1px solid #000;
        text-align: center;
        background: #fff;
        font-weight: 700;
      }
      table.items td {
        padding: 4px 6px;
        border-left: 1px solid #000;
        border-right: 1px solid #000;
        border-bottom: none;
        vertical-align: top;
        line-height: 1.2;
      }
      table.items td.name {
        min-height: 44px;
      }
      table.items th.col-no,
      table.items td.col-no {
        width: 20px;
        text-align: center;
      }
      table.items th.col-name,
      table.items td.col-name {
        width: 30%;
      }
      table.items th.col-qty,
      table.items td.col-qty {
        width: 8%;
        text-align: center;
      }
      table.items th.col-unit,
      table.items td.col-unit {
        width: 8%;
        text-align: center;
      }
      table.items th.col-price,
      table.items td.col-price {
        width: 18%;
        text-align: right;
      }
      table.items th.col-sum,
      table.items td.col-sum {
        width: 18%;
        text-align: right;
      }
      table.items th.col-ket,
      table.items td.col-ket {
        width: 10%;
      }
      .items tfoot td {
        border-top: 1px solid #000;
        padding: 8px;
      }
      .notes {
        margin-top: 6px;
        display: flex;
        gap: 8px;
      }
      .notes .left {
        width: 60%;
      }
      .notes .right {
        width: 40%;
        border: 1px solid #000;
        padding: 6px;
      }
      .signatures {
        margin-top: 24px;
        display: flex;
        justify-content: space-between;
      }
      .signature {
        width: 32%;
        text-align: center;
      }
      .signature .line {
        border-top: 1px solid #000;
        margin-top: 36px;
      }
    </style>
  </head>
  <body>
    <div class="container-head">
      <div class="header">
        <div class="title">SHO SHA MART</div>
        <i>
          Jl. Pahlawan No.33, RT.10/RW.4, Sukabumi Sel., Kec. Kb. Jeruk, Kota
          Jakarta Barat, Daerah Khusus Ibukota Jakarta 11560</i
        >
      </div>

      <div class="header-cust">
          <p>${date}</p>
          <i>${printData.value.branch?.code || '-'}</i>
        <span>Kepada Yth,</span>
        <div class="title-cust">${printData.value.branch?.name || '-'}</div> 
        <i>${printData.value.branch?.address || '-'}</i>
      </div>
    </div>

    <table class="items">
      <thead>
        <tr>
          <th class="col-no">NO</th>
          <th class="col-name">PESANAN</th>
          <th class="col-qty">QTY</th>
          <th class="col-unit">SATUAN</th>
          <th class="col-price">HARGA</th>
          <th class="col-sum">JUMLAH</th>
          <th class="col-ket">KET</th>
        </tr>
      </thead>
      <tbody>
        ${itemsHtml}
      </tbody>
      <tfoot>
        <tr>
          <td colspan="4"></td>
          <td style="text-align: right">&nbsp;</td>
          <td style="text-align: right; padding-right: 6px">
            Rp ${grandTotal.toLocaleString('id-ID')}
          </td>
          <td></td>
        </tr>
      </tfoot>
    </table>

    <div class="notes">
      <div class="left">
        <div style="font-weight: bold">CATATAN/KETERANGAN:</div>
        <div style="margin-top: 6px">${printData.value.notes || '-'}</div>
      </div>
      <div class="right">
        <strong>PERHATIAN:</strong>
        <ol style="margin-top: 6px; padding-left: 18px; font-size: 9pt">
          <li>Surat Jalan ini merupakan bukti resmi penerimaan barang</li>
          <li>Surat Jalan ini bukti penjualan</li>
        </ol>
      </div>
    </div>

    <div class="signatures">
      <div class="signature">
        <div>PELANGGAN</div>
        <div class="line">
          ${isHutang ? (printData.value.branch?.name || '') : ''}
        </div>
      </div>
      <div class="signature">
        <div>&nbsp;</div>
        <div class="line">Bagian Pengiriman</div>
      </div>
      <div class="signature">
        <div>SHO-SHA MART</div>
        <div class="line">${cashierName}</div>
      </div>
    </div>
  </body>
</html>
  `
  printWindow.document.write(html)
  printWindow.document.close()
  
  // Auto-trigger print dialog
  printWindow.onload = () => {
    printWindow.focus()
    printWindow.print()
  }
}

async function downloadPDF() {
  if (!selected.value) return
  
  try {
    // Prepare data (like preparePrint does)
    const branchMap = Object.fromEntries(branches.value.map(b => [b.id, b]))
    const pdfData = {
      ...selected.value,
      branch: branchMap[selected.value.branch_id] || { name: selected.value.branch_name, address: '' },
      items: items.value,
    }
    
    // Dynamically import libraries
    const [{ default: jsPDF }, { default: html2canvas }] = await Promise.all([
      import('jspdf'),
      import('html2canvas')
    ])
    
    // Create temporary container with print HTML
    const tempDiv = document.createElement('div')
    tempDiv.style.position = 'absolute'
    tempDiv.style.left = '-9999px'
    tempDiv.style.width = '210mm' // A4 width
    tempDiv.style.padding = '8mm'
    tempDiv.style.backgroundColor = 'white'
    
    const txDate = pdfData.created_at ? new Date(pdfData.created_at) : new Date()
    const date = txDate.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
    
    let itemsHtml = ''
    pdfData.items.forEach((item: any, idx: number) => {
      itemsHtml += `
        <tr>
          <td style="text-align: center; padding: 3px 2px; border: 1px solid #ddd;">${idx + 1}</td>
          <td style="padding: 3px 2px; border: 1px solid #ddd;">${item.name || item.product_id}</td>
          <td style="text-align: center; padding: 3px 2px; border: 1px solid #ddd;">${item.qty}</td>
          <td style="text-align: center; padding: 3px 2px; border: 1px solid #ddd;">${item.unit || 'PCS'}</td>
          <td style="text-align: right; padding: 3px 2px; border: 1px solid #ddd;">Rp ${Number(item.price || 0).toLocaleString('id-ID')}</td>
          <td style="text-align: right; padding: 3px 2px; border: 1px solid #ddd;">Rp ${Number(item.subtotal || 0).toLocaleString('id-ID')}</td>
        </tr>
      `
    })
    
    const grandTotal = pdfData.items.reduce((sum: number, item: any) => sum + (item.subtotal || 0), 0)
    
    tempDiv.innerHTML = `
      <div style="font-family: Arial, sans-serif; font-size: 10pt; color: #000;">
        <div style="display: flex; justify-content: space-between; margin-bottom: 12px;">
          <div style="width: 45%;">
            <div style="font-size: 18pt; font-weight: 800;">SHO SHA MART</div>
            <div style="font-size: 8pt; margin-top: 4px;">Jl. Pahlawan No.33, RT.10/RW.4, Sukabumi Sel., Jakarta Barat 11560</div>
          </div>
          <div style="width: 45%; text-align: right;">
            <div>${date}</div>
            <div style="font-weight: 700; margin-top: 4px;">${pdfData.branch?.name || '-'}</div>
            <div style="font-size: 8pt;">${pdfData.branch?.address || ''}</div>
          </div>
        </div>
        <table style="width: 100%; border-collapse: collapse; margin-top: 12px;">
          <thead>
            <tr style="background: #f0f0f0;">
              <th style="padding: 6px; border: 1px solid #000; text-align: center;">NO</th>
              <th style="padding: 6px; border: 1px solid #000;">PESANAN</th>
              <th style="padding: 6px; border: 1px solid #000; text-align: center;">QTY</th>
              <th style="padding: 6px; border: 1px solid #000; text-align: center;">SATUAN</th>
              <th style="padding: 6px; border: 1px solid #000; text-align: right;">HARGA</th>
              <th style="padding: 6px; border: 1px solid #000; text-align: right;">JUMLAH</th>
            </tr>
          </thead>
          <tbody>
            ${itemsHtml}
            <tr>
              <td colspan="5" style="text-align: right; padding: 8px; border: 1px solid #000; font-weight: bold;">TOTAL</td>
              <td style="text-align: right; padding: 8px; border: 1px solid #000; font-weight: bold;">Rp ${grandTotal.toLocaleString('id-ID')}</td>
            </tr>
          </tbody>
        </table>
        <div style="margin-top: 12px; font-size: 9pt;">
          <strong>CATATAN:</strong> ${pdfData.notes || '-'}
        </div>
      </div>
    `
    
    document.body.appendChild(tempDiv)
    
    const canvas = await html2canvas(tempDiv, { scale: 2 })
    document.body.removeChild(tempDiv)
    
    const imgData = canvas.toDataURL('image/png')
    const pdf = new jsPDF('p', 'mm', 'a4')
    const pdfWidth = pdf.internal.pageSize.getWidth()
    const pdfHeight = (canvas.height * pdfWidth) / canvas.width
    
    pdf.addImage(imgData, 'PNG', 0, 0, pdfWidth, pdfHeight)
    pdf.save(`Invoice-${pdfData.receipt_no || Date.now()}.pdf`)
    
    toast.success('✓ PDF berhasil diunduh')
  } catch (err) {
    toast.error(`Error: ${(err as Error).message}`)
  }
}

async function pruneDeleted() {
  try {
    const res = await fetch('/api/sync/prune-deleted', { method: 'POST' })
    if (!res.ok) throw new Error(`HTTP ${res.status}`)
    toast.success('✓ Bersihkan data terhapus dikirim')
    await load()
  } catch (e) {
    toast.error((e as Error).message)
  }
}

async function removeSale() {
  if (!selected.value) return
  
  const saleInfo = `${selected.value.receipt_no || selected.value.id}`
  
  toast(`Hapus transaksi ${saleInfo}?`, {
    description: 'Tindakan ini tidak bisa dibatalkan.',
    action: {
      label: 'Hapus',
      onClick: async () => {
        try {
          await api.deleteSale(selected.value!.id)
          toast.success(`✓ Terhapus. Menunggu sinkronisasi ke server.`)
          closeDetail()
          await load()
        } catch (err) {
          toast.error(`Error: ${(err as Error).message}`)
        }
      }
    },
    cancel: {
      label: 'Batal',
      onClick: () => {}
    },
    duration: 5000
  })
}

function fmtCurrency(n: number) {
  return `Rp ${Number(n || 0).toLocaleString('id-ID')}`
}

function fmtDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
}

onMounted(async () => {
  // Don't set default date range - show all data by default
  // User can click "Bulan ini" or "Tahun ini" to filter
  await Promise.all([
    load(),
    api.listProducts().then(p => (products.value = p)),
    api.listBranches().then(b => (branches.value = b)),
  ])
})
</script>

<template>
  <section class="space-y-4">
    <header class="flex items-center justify-between">
      <div>
        <p class="text-lg font-bold uppercase tracking-[0.2em] text-emerald-600">Riwayat</p>
        <h2 class="text-2xl font-semibold text-slate-900">History Order</h2>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" :disabled="exporting" @click="exportData">
          {{ exporting ? 'Exporting...' : 'Export Excel' }}
        </Button>
        <Button variant="ghost" @click="load">Refresh</Button>
        <Button variant="ghost" class="text-slate-400" @click="pruneDeleted">Bersihkan Data Terhapus</Button>
      </div>
    </header>

    <Card>
      <div class="p-4 space-y-3">
        <div v-if="error" class="text-rose-300 text-sm">{{ error }}</div>
        <div v-if="loading" class=" text-sm">Memuat...</div>

        <div class="flex flex-wrap gap-3 items-end">
          <div>
            <label class="block text-xs text-slate-500">Tanggal Mulai</label>
            <input v-model="dateRange.start" type="date" class="border border-slate-200 rounded px-2 py-1 text-sm" />
          </div>
          <div>
            <label class="block text-xs text-slate-500">Tanggal Akhir</label>
            <input v-model="dateRange.end" type="date" class="border border-slate-200 rounded px-2 py-1 text-sm" />
          </div>
          <Button variant="ghost" size="sm" class="text-emerald-600" @click="setDefaultMonth">Bulan ini</Button>
          <Button variant="ghost" size="sm" class="text-blue-600" @click="setDefaultYear">Tahun ini</Button>
          <Button variant="ghost" size="sm" class="text-slate-600" @click="clearDateFilter">Tampilkan Semua</Button>
          <div>
            <label class="block text-xs text-slate-500">Filter Kode Cabang</label>
            <input v-model="searchCode" type="text" placeholder="Mis. SHOSHA" class="border border-slate-200 rounded px-2 py-1 text-sm" />
          </div>
        </div>

        <!-- Grouped by Branch with Accordion -->
        <div v-if="!loading && groupedByBranch.length > 0" class="mt-4">
          <div class="mb-3 text-sm text-slate-600">
            Total: {{ groupedByBranch.length }} cabang, {{ displayedSales.length }} transaksi
          </div>
          
          <Accordion type="multiple" class="w-full">
            <AccordionItem v-for="group in groupedByBranch" :key="group.branchName" :value="group.branchName">
              <AccordionTrigger>
                <div class="flex items-center justify-between w-full pr-4">
                  <div class="flex items-center gap-3">
                    <span class="font-bold text-emerald-600">{{ group.branchCode || 'N/A' }}</span>
                    <span class="font-semibold text-slate-900">{{ group.branchName }}</span>
                    <span class="text-xs bg-slate-100 px-2 py-1 rounded">{{ group.count }} transaksi</span>
                  </div>
                  <span class="font-semibold text-slate-900">{{ fmtCurrency(group.total) }}</span>
                </div>
              </AccordionTrigger>
              
              <AccordionContent>
                <div class="px-4">
                  <table class="w-full text-sm">
                    <thead>
                      <tr class="text-left text-slate-500 border-b">
                        <th class="py-2 px-2">Tanggal</th>
                        <th class="py-2 px-2">No. Invoice</th>
                        <th class="py-2 px-2">Metode</th>
                        <th class="py-2 px-2 text-right">Total</th>
                        <th class="py-2 px-2">Aksi</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="s in group.sales" :key="s.id" class="border-b hover:bg-slate-50">
                        <td class="py-2 px-2">{{ fmtDate(s.created_at) }}</td>
                        <td class="py-2 px-2 font-mono text-xs">{{ s.receipt_no }}</td>
                        <td class="py-2 px-2">
                          <span class="uppercase text-xs px-2 py-1 rounded" :class="s.payment_method === 'cash' ? 'bg-green-100 text-green-700' : 'bg-orange-100 text-orange-700'">
                            {{ s.payment_method }}
                          </span>
                        </td>
                        <td class="py-2 px-2 text-right font-semibold">{{ fmtCurrency(s.total) }}</td>
                        <td class="py-2 px-2">
                          <div class="flex gap-1">
                            <Button size="sm" variant="ghost" @click="openEditDialog(s)">Edit</Button>
                            <Button size="sm" variant="ghost" @click="openProductsEditDialog(s)">Edit Produk</Button>
                            <Button size="sm" variant="outline" @click="openDetail(s.id)">Detail</Button>
                          </div>
                        </td>
                      </tr>
                    </tbody>
                  </table>
                </div>
              </AccordionContent>
            </AccordionItem>
          </Accordion>
        </div>

        <div v-if="!loading && displayedSales.length === 0" class="py-8 text-center text-slate-500">
          Belum ada transaksi
        </div>
      </div>
    </Card>

    <!-- Detail Modal -->
    <div v-if="showDetail" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click="closeDetail">
      <Card class="max-w-2xl w-full" @click.stop>
        <div class="p-6 space-y-4">
          <div class="flex items-center justify-between">
            <h3 class="text-lg font-semibold text-white">Detail Pembelian</h3>
            <Button variant="ghost" @click="closeDetail">Tutup</Button>
          </div>
          <div class="text-sm font-bold">
            <p><span class="font-semibold">Cabang:</span> {{ selected?.branch_name || selected?.branch_id }}</p>
            <p><span class="font-semibold">Tanggal:</span> {{ selected?.created_at ? fmtDate(selected!.created_at) : '-'
            }}</p>
            <p><span class="font-semibold">Invoice:</span> {{ selected?.receipt_no }}</p>
            <p><span class="font-semibold">Metode:</span> {{ selected?.payment_method }}</p>
          </div>
          <div class="overflow-x-auto">
            <table class="min-w-full text-sm text-left text-slate-700">
              <thead>
                <tr class="bg-white border border-slate-200">
                  <th class="px-3 py-2">No</th>
                  <th class="px-3 py-2">Nama Barang</th>
                  <th class="px-3 py-2">Qty</th>
                  <th class="px-3 py-2">Harga</th>
                  <th class="px-3 py-2 text-right">Jumlah</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(it, i) in items" :key="it.id" class="border border-slate-200">
                  <td class="px-3 py-2">{{ i + 1 }}</td>
                  <td class="px-3 py-2">{{ it.name || it.product_id }}</td>
                  <td class="px-3 py-2">{{ it.qty }}</td>
                  <td class="px-3 py-2">{{ fmtCurrency(it.price) }}</td>
                  <td class="px-3 py-2 text-right">{{ fmtCurrency(it.subtotal) }}</td>
                </tr>
                <tr>
                  <td colspan="4" class="px-3 py-2 text-right font-semibold">Total</td>
                  <td class="px-3 py-2 text-right font-bold">{{ fmtCurrency(selected?.total || 0) }}</td>
                </tr>
              </tbody>
            </table>
          </div>

          <div class="flex justify-between items-center">
            <Button variant="ghost" class="text-rose-400" @click="removeSale">Hapus Transaksi</Button>
            <div class="flex gap-2">
              <Button variant="outline" @click="preparePrint">Cetak Faktur</Button>
              <Button @click="downloadPDF">Download PDF</Button>
              <Button variant="ghost" @click="closeDetail">Tutup</Button>
            </div>
          </div>
        </div>
      </Card>
    </div>

    <!-- Edit Transaction Modal -->
    <div v-if="showEditDialog" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click="closeEditDialog">
      <Card class="max-w-lg w-full" @click.stop>
        <div class="p-6 space-y-4">
          <div class="flex items-center justify-between border-b pb-3">
            <h3 class="text-lg font-semibold text-slate-900">Edit Transaksi</h3>
            <Button variant="ghost" size="sm" @click="closeEditDialog">✕</Button>
          </div>

          <div class="space-y-4">
            <div class="space-y-1">
              <Label>Tanggal Transaksi</Label>
              <Input v-model="editForm.created_at" type="date" />
            </div>

            <div class="space-y-1">
              <Label>Cabang Pembeli</Label>
              <BranchSelect v-model="editForm.branch_id" :branches="branches" />
            </div>

            <div class="space-y-1">
              <Label>Metode Pembayaran</Label>
              <Select v-model="editForm.payment_method">
                <option value="cash">Cash (Tunai)</option>
                <option value="hutang">Hutang</option>
              </Select>
            </div>

            <div class="space-y-1">
              <Label>Catatan</Label>
              <textarea 
                v-model="editForm.notes" 
                class="w-full border border-slate-200 rounded px-3 py-2 text-sm"
                rows="3"
                placeholder="Catatan tambahan..."
              ></textarea>
            </div>
          </div>

          <div class="flex justify-end gap-2 pt-3 border-t">
            <Button variant="ghost" @click="closeEditDialog" :disabled="savingEdit">Batal</Button>
            <Button @click="saveEdit" :disabled="savingEdit">
              {{ savingEdit ? 'Menyimpan...' : 'Simpan Perubahan' }}
            </Button>
          </div>
        </div>
      </Card>
    </div>

    <!-- Edit Products Modal -->
    <div v-if="showProductsEditDialog" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click="closeProductsEditDialog">
      <Card class="max-w-2xl w-full max-h-96 overflow-y-auto" @click.stop>
        <div class="p-6 space-y-4">
          <div class="flex items-center justify-between border-b pb-3 sticky top-0 bg-white">
            <h3 class="text-lg font-semibold text-slate-900">Edit Produk - {{ editingProductsForSale?.receipt_no }}</h3>
            <Button variant="ghost" size="sm" @click="closeProductsEditDialog">✕</Button>
          </div>

          <!-- Current Products Table -->
          <div class="space-y-2">
            <h4 class="font-semibold text-slate-700">Produk Saat Ini</h4>
            <div class="overflow-x-auto">
              <table class="w-full text-sm border-collapse">
                <thead class="bg-slate-50">
                  <tr>
                    <th class="py-2 px-2 text-left">Produk</th>
                    <th class="py-2 px-2 text-center">Qty</th>
                    <th class="py-2 px-2 text-right">Harga</th>
                    <th class="py-2 px-2 text-right">Subtotal</th>
                    <th class="py-2 px-2 text-center">Aksi</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="item in editingItems" :key="item.id" class="border-t">
                    <td class="py-2 px-2">
                      <div>
                        <div class="font-medium text-slate-900">{{ item.name }}</div>
                        <div class="text-xs text-slate-500">{{ item.unit || '-' }}</div>
                      </div>
                    </td>
                    <td class="py-2 px-2">
                      <Input 
                        :model-value="item.qty" 
                        @update:model-value="(v) => {
                          const num = parseInt(v as string) || 0
                          if (num > 0) updateProductInSale(item, num, item.price)
                        }"
                        type="number" 
                        min="1"
                        class="w-20"
                      />
                    </td>
                    <td class="py-2 px-2 text-right">
                      <Input 
                        :model-value="item.price.toString()" 
                        @update:model-value="(v) => {
                          const num = parseFloat(v as string) || 0
                          if (num > 0) updateProductInSale(item, item.qty, num)
                        }"
                        type="number" 
                        min="0"
                        step="0.01"
                        class="w-28 text-right"
                      />
                    </td>
                    <td class="py-2 px-2 text-right font-semibold">
                      {{ fmtCurrency(item.qty * item.price) }}
                    </td>
                    <td class="py-2 px-2 text-center">
                      <Button 
                        size="sm" 
                        variant="ghost" 
                        @click="deleteProductFromSale(item.id)"
                        :disabled="savingProducts"
                        class="text-red-600 hover:text-red-700"
                      >
                        Hapus
                      </Button>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <!-- Add New Product -->
          <div class="border-t pt-4 space-y-3">
            <h4 class="font-semibold text-slate-700">Tambah Produk</h4>
            <div class="grid grid-cols-3 gap-2">
              <div class="space-y-1">
                <Label>Produk</Label>
                <Select v-model="newItemForm.product_id">
                  <option value="">Pilih Produk</option>
                  <option v-for="p in products" :key="p.id" :value="p.id">
                    {{ p.name }}
                  </option>
                </Select>
              </div>
              <div class="space-y-1">
                <Label>Qty</Label>
                <Input v-model.number="newItemForm.qty" type="number" min="1" />
              </div>
              <div class="space-y-1">
                <Label>Harga</Label>
                <Input v-model.number="newItemForm.price" type="number" min="0" step="0.01" />
              </div>
            </div>
            <Button 
              @click="addProductToSale" 
              :disabled="savingProducts || !newItemForm.product_id"
              class="w-full"
            >
              {{ savingProducts ? 'Menyimpan...' : 'Tambah Produk' }}
            </Button>
          </div>

          <div class="flex justify-end gap-2 pt-3 border-t">
            <Button variant="ghost" @click="closeProductsEditDialog" :disabled="savingProducts">Tutup</Button>
            <Button @click="closeAndReloadProducts" :disabled="savingProducts">
              {{ savingProducts ? 'Menyimpan...' : 'Selesai' }}
            </Button>
          </div>
        </div>
      </Card>
    </div>
  </section>
</template>
