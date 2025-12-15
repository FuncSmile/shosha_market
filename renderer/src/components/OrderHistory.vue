<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, type Sale, type SaleItem, type Product, type Branch } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'

const sales = ref<Sale[]>([])
const loading = ref(false)
const error = ref('')
const exporting = ref(false)

const showDetail = ref(false)
const selected: any = ref<Sale | null>(null)
type AugmentedItem = SaleItem & { name?: string; unit?: string; subtotal?: number }
const items = ref<AugmentedItem[]>([])

const products = ref<Product[]>([])
const branches = ref<Branch[]>([])

const showPrintDialog = ref(false)
const printData = ref<any>(null)

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
    const url = await api.exportSales()
    const a = document.createElement('a')
    a.href = url
    a.download = `sales-export-${new Date().toISOString().slice(0, 10)}.xlsx`
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

  const isHutang = printData.value.payment_method === 'hutang'
  const txDate = printData.value.created_at ? new Date(printData.value.created_at) : new Date()
  const date = txDate.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
  const customerName = (printData.value.notes || '').trim() || '-'
  const cashierName = 'SHOSHA MART'

  let itemsHtml = ''
  printData.value.items.forEach((item: any, idx: number) => {
    itemsHtml += `
      <tr>
        <td style="text-align: center;">${idx + 1}</td>
        <td>${item.name || item.product_id}</td>
        <td style="text-align: center;">${item.qty}</td>
        <td style="text-align: center;">${item.unit || 'PCS'}</td>
        <td style="text-align: right;">Rp ${Number(item.price || 0).toLocaleString('id-ID')}</td>
        <td style="text-align: right;">Rp ${Number(item.subtotal || 0).toLocaleString('id-ID')}</td>
        <td></td>
      </tr>
    `
  })
  const grandTotal = printData.value.items.reduce((sum: number, item: any) => sum + (item.subtotal || 0), 0)

  const html = `
    <!DOCTYPE html>
    <html>
    <head>
      <meta charset="UTF-8">
      <title>${isHutang ? 'Surat Jalan' : 'Struk Pembayaran'}</title>
      <style>
        @page { size: A4; margin: 20mm; }
        body { font-family: Arial, sans-serif; font-size: 11pt; }
        .header { text-align: center; margin-bottom: 20px; }
        .header h1 { margin: 0; font-size: 18pt; }
        .info { margin-bottom: 15px; }
        .info table { width: 100%; }
        .info td { padding: 3px 0; }
        .info td:first-child { width: 100px; font-weight: bold; }
        table.items { width: 100%; border-collapse: collapse; margin-top: 15px; }
        table.items th, table.items td { border: 1px solid #000; padding: 5px; }
        table.items th { background: #f0f0f0; font-weight: bold; }
        .footer { margin-top: 40px; }
        .footer table { width: 100%; }
        .footer td { text-align: center; padding-top: 60px; }
      </style>
    </head>
    <body>
      <div class="header">
        <h1>SHO SHA MART</h1>
      </div>
      
      <div class="info">
        <table>
          <tr>
            <td>NAMA</td>
            <td>: ${printData.value.branch?.name || '-'}</td>
          </tr>
          <tr>
            <td>TANGGAL</td>
            <td>: ${date}</td>
          </tr>
          <tr>
            <td>ALAMAT</td>
            <td>: ${printData.value.branch?.address || '-'}</td>
          </tr>
        </table>
      </div>
      
      <table class="items">
        <thead>
          <tr>
            <th style="width: 30px;">NO</th>
            <th>PESANAN</th>
            <th style="width: 50px;">QTY</th>
            <th style="width: 50px;">SATUAN</th>
            <th style="width: 90px;">HARGA</th>
            <th style="width: 90px;">JUMLAH</th>
            <th style="width: 100px;">KET</th>
          </tr>
        </thead>
        <tbody>
          ${itemsHtml}
          <tr>
            <td colspan="4"></td>
            <td></td>
            <td style="text-align: right; font-weight: bold;">Rp ${grandTotal.toLocaleString('id-ID')}</td>
            <td></td>
          </tr>
        </tbody>
      </table>
      
      <div class="footer">
        ${isHutang ? `
          <table style="margin-top: 15px; width: 100%;">
            <tr>
              <td style="width: 50%; vertical-align: top; padding-right: 10px;">
                <p>CATATAN/KETERANGAN :</p>
                <p>${printData.value.notes || '-'}</p>
              </td>
              <td style="width: 25%; text-align: center;">
                <p>PELANGGAN</p>
                <p style="margin-top: 60px; border-top: 1px solid #000; display: inline-block; padding-top: 5px; min-width: 120px;">
                  ${customerName}
                </p>
              </td>
              <td style="width: 25%; text-align: center;">
                <p>SHO-SHA MART</p>
                <p style="margin-top: 60px; border-top: 1px solid #000; display: inline-block; padding-top: 5px; min-width: 120px;">
                  ${cashierName}
                </p>
              </td>
            </tr>
          </table>
        ` : `
          <p style="text-align: center; margin-top: 30px; font-size: 14pt; font-weight: bold;">
            TERIMA KASIH
          </p>
          <p style="text-align: center; margin-top: 10px; font-size: 10pt;">
            No. Invoice: ${printData.value.receipt_no || 'Auto-generated'}
          </p>
        `}
      </div>
      
      <script type="text/javascript">
        window.onload = function() {
          window.print();
          window.onafterprint = function() { window.close(); };
        };
      <\/script>
    </body>
    </html>
  `
  printWindow.document.write(html)
  printWindow.document.close()
}

function fmtCurrency(n: number) {
  return `Rp ${Number(n || 0).toLocaleString('id-ID')}`
}

function fmtDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
}

onMounted(async () => {
  await Promise.all([load(), api.listProducts().then(p => products.value = p), api.listBranches().then(b => branches.value = b)])
})
</script>

<template>
  <section class="space-y-4">
    <header class="flex items-center justify-between">
      <div>
        <p class="text-sm uppercase tracking-[0.2em] text-emerald-200/80">Riwayat</p>
        <h2 class="text-2xl font-semibold text-white">History Order</h2>
      </div>
      <div class="flex gap-2">
        <Button variant="outline" :disabled="exporting" @click="exportData">
          {{ exporting ? 'Exporting...' : 'Export Excel' }}
        </Button>
        <Button variant="ghost" @click="load">Refresh</Button>
      </div>
    </header>

    <Card>
      <div class="p-4 space-y-3">
        <div v-if="error" class="text-rose-300 text-sm">{{ error }}</div>
        <div v-if="loading" class="text-slate-300 text-sm">Memuat...</div>

        <div class="overflow-x-auto">
          <table class="min-w-full text-sm text-left text-slate-300">
            <thead>
              <tr class="bg-slate-800/60 border border-slate-700">
                <th class="px-3 py-2">Tanggal</th>
                <th class="px-3 py-2">Cabang</th>
                <th class="px-3 py-2">No. Invoice</th>
                <th class="px-3 py-2">Metode</th>
                <th class="px-3 py-2 text-right">Total</th>
                <th class="px-3 py-2">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="s in sales" :key="s.id" class="border border-slate-700 hover:bg-slate-800/40">
                <td class="px-3 py-2">{{ fmtDate(s.created_at) }}</td>
                <td class="px-3 py-2">
                  <button class="text-emerald-300 hover:underline" @click="openDetail(s.id)">
                    {{ s.branch_name || s.branch_id || '-' }}
                  </button>
                </td>
                <td class="px-3 py-2">{{ s.receipt_no }}</td>
                <td class="px-3 py-2 uppercase">{{ s.payment_method }}</td>
                <td class="px-3 py-2 text-right">{{ fmtCurrency(s.total) }}</td>
                <td class="px-3 py-2">
                  <Button size="sm" @click="openDetail(s.id)">Detail</Button>
                </td>
              </tr>
              <tr v-if="sales.length === 0 && !loading">
                <td colspan="6" class="px-3 py-4 text-center text-slate-400">Belum ada transaksi</td>
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
          <div class="text-sm text-slate-300">
            <p><span class="font-semibold">Cabang:</span> {{ selected?.branch_name || selected?.branch_id }}</p>
            <p><span class="font-semibold">Tanggal:</span> {{ selected?.created_at ? fmtDate(selected!.created_at) : '-' }}</p>
            <p><span class="font-semibold">Invoice:</span> {{ selected?.receipt_no }}</p>
            <p><span class="font-semibold">Metode:</span> {{ selected?.payment_method }}</p>
          </div>
          <div class="overflow-x-auto">
            <table class="min-w-full text-sm text-left text-slate-300">
              <thead>
                <tr class="bg-slate-800/60 border border-slate-700">
                  <th class="px-3 py-2">No</th>
                  <th class="px-3 py-2">Nama Barang</th>
                  <th class="px-3 py-2">Qty</th>
                  <th class="px-3 py-2">Harga</th>
                  <th class="px-3 py-2 text-right">Jumlah</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(it, i) in items" :key="it.id" class="border border-slate-700">
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

          <div class="flex justify-end gap-2">
            <Button variant="outline" @click="preparePrint">Cetak Struk/Surat Jalan</Button>
            <Button variant="ghost" @click="closeDetail">Tutup</Button>
          </div>
        </div>
      </Card>
    </div>
  </section>
</template>
