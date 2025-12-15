<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { api, type Sale, type SaleItem } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'

const sales = ref<Sale[]>([])
const loading = ref(false)
const error = ref('')
const exporting = ref(false)

const showDetail = ref(false)
const selected: any = ref<Sale | null>(null)
const items = ref<SaleItem[]>([])

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
    items.value = (sale.items || []).map(i => ({
      ...i,
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

function fmtCurrency(n: number) {
  return `Rp ${Number(n || 0).toLocaleString('id-ID')}`
}

function fmtDate(s: string) {
  const d = new Date(s)
  return d.toLocaleDateString('id-ID', { day: 'numeric', month: 'long', year: 'numeric' })
}

onMounted(load)
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
                  <td class="px-3 py-2">{{ it.product_id }}</td>
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
        </div>
      </Card>
    </div>
  </section>
</template>
