<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, type Sale, type SaleItem, type Product, type Branch } from '../api'
import { toast } from 'vue-sonner'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'

const sales = ref<Sale[]>([])
const loading = ref(false)
const error = ref('')
const exporting = ref(false)

const showDetail = ref(false)
const selected = ref<Sale | null>(null)
type AugmentedItem = SaleItem & { name?: string; unit?: string; subtotal?: number }
const items = ref<AugmentedItem[]>([])

const products = ref<Product[]>([])
const branches = ref<Branch[]>([])

const searchCode = ref('')

const showPrintDialog = ref(false)
const printData = ref<any>(null)

const sortKey = ref<'created_at' | 'branch_code'>('created_at')
const sortDir = ref<'asc' | 'desc'>('desc')
const dateRange = ref<{ start: string; end: string }>({ start: '', end: '' })

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
    .sort((a, b) => {
      const dir = sortDir.value === 'asc' ? 1 : -1
      if (sortKey.value === 'branch_code') {
        return a.branch_code.localeCompare(b.branch_code) * dir
      }
      return (new Date(a.created_at).getTime() - new Date(b.created_at).getTime()) * dir
    })
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
  const toISO = (d: Date) => d.toISOString().slice(0, 10)
  dateRange.value = { start: toISO(start), end: toISO(end) }
}

function toggleSort(key: 'created_at' | 'branch_code') {
  if (sortKey.value === key) {
    sortDir.value = sortDir.value === 'asc' ? 'desc' : 'asc'
  } else {
    sortKey.value = key
    sortDir.value = 'asc'
  }
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
  setDefaultMonth()
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
          <Button variant="ghost" class="text-emerald-600" @click="setDefaultMonth">Bulan ini</Button>
          <div>
            <label class="block text-xs text-slate-500">Filter Kode Cabang</label>
            <input v-model="searchCode" type="text" placeholder="Mis. SHOSHA" class="border border-slate-200 rounded px-2 py-1 text-sm" />
          </div>
        </div>

        <div class="overflow-x-auto">
          <table class="min-w-full text-sm text-left text-slate-700">
            <thead>
              <tr class="bg-white border border-slate-200">
                <th class="px-3 py-2 cursor-pointer select-none" @click="toggleSort('created_at')">
                  Tanggal
                  <span v-if="sortKey === 'created_at'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
                </th>
                <th class="px-3 py-2">Kode</th>
                <th class="px-3 py-2">Cabang</th>
                <th class="px-3 py-2">No. Invoice</th>
                <th class="px-3 py-2">Metode</th>
                <th class="px-3 py-2 text-right">Total</th>
                <th class="px-3 py-2">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="s in displayedSales" :key="s.id" class="border border-slate-200 hover:bg-slate-50">
                <td class="px-3 py-2">{{ fmtDate(s.created_at) }}</td>
                <td class="px-3 py-2 cursor-pointer select-none" @click="toggleSort('branch_code')">
                  <span class="font-semibold">{{ s.branch_code || '-' }}</span>
                  <span v-if="sortKey === 'branch_code'">{{ sortDir === 'asc' ? '↑' : '↓' }}</span>
                </td>
                <td class="px-3 py-2">
                  <button class="text-emerald-600 hover:underline" @click="openDetail(s.id)">
                    {{ s.branch_name_resolved || s.branch_id || '-' }}
                  </button>
                </td>
                <td class="px-3 py-2">{{ s.receipt_no }}</td>
                <td class="px-3 py-2 uppercase">{{ s.payment_method }}</td>
                <td class="px-3 py-2 text-right">{{ fmtCurrency(s.total) }}</td>
                <td class="px-3 py-2">
                  <Button size="sm" @click="openDetail(s.id)">Detail</Button>
                </td>
              </tr>
              <tr v-if="displayedSales.length === 0 && !loading">
                <td colspan="7" class="px-3 py-4 text-center text-slate-500">Belum ada transaksi</td>
              </tr>
            </tbody>
          </table>
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
  </section>
</template>
