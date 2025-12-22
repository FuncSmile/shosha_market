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
const selected = ref<Sale | null>(null)
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
}

function fmtCurrency(n: number) {
  return `Rp ${Number(n || 0).toLocaleString('id-ID')}`
}

function fmtDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
}

onMounted(async () => {
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
      </div>
    </header>

    <Card>
      <div class="p-4 space-y-3">
        <div v-if="error" class="text-rose-300 text-sm">{{ error }}</div>
        <div v-if="loading" class=" text-sm">Memuat...</div>

        <div class="overflow-x-auto">
          <table class="min-w-full text-sm text-left text-slate-700">
            <thead>
              <tr class="bg-white border border-slate-200">
                <th class="px-3 py-2">Tanggal</th>
                <th class="px-3 py-2">Cabang</th>
                <th class="px-3 py-2">No. Invoice</th>
                <th class="px-3 py-2">Metode</th>
                <th class="px-3 py-2 text-right">Total</th>
                <th class="px-3 py-2">Aksi</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="s in sales" :key="s.id" class="border border-slate-200 hover:bg-slate-50">
                <td class="px-3 py-2">{{ fmtDate(s.created_at) }}</td>
                <td class="px-3 py-2">
                  <button class="text-emerald-600 hover:underline" @click="openDetail(s.id)">
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
                <td colspan="6" class="px-3 py-4 text-center text-slate-500">Belum ada transaksi</td>
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

          <div class="flex justify-end gap-2">
            <Button variant="outline" @click="preparePrint">Cetak Faktur Penjualan</Button>
            <Button variant="ghost" @click="closeDetail">Tutup</Button>
          </div>
        </div>
      </Card>
    </div>
  </section>
</template>
