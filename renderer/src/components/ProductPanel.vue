<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, type Product } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import { useToast } from '../composables/useToast'
import { toast } from 'vue-sonner'
import * as XLSX from 'xlsx'

const { success, error, warning } = useToast()
const products = ref<Product[]>([])
const loading = ref(false)
const saving = ref(false)
const syncedInfo = ref<Record<string, boolean>>({})
// UI state for inline stock adjust
const adjustingProductId = ref<string | null>(null)
const adjustingAmount = ref<number>(1)
const adjustingDelta = ref<number>(1) // +1 for masuk, -1 for keluar
// List search + pagination (5 per page)
const listSearch = ref('')
const listPage = ref(1)
const listPageSize = 5
const sortUnitDir = ref<'asc' | 'desc'>('asc')
const editingProductId = ref<string | null>(null)
const editForm = ref<Partial<Product>>({})
const filteredList = computed(() => {
  const q = listSearch.value.trim().toLowerCase()
  if (!q) return products.value
  return products.value.filter(p => p.name?.toLowerCase().includes(q) || p.unit?.toLowerCase().includes(q))
})
const filteredListSorted = computed(() => {
  const sorted = [...filteredList.value]
  sorted.sort((a, b) => {
    const aUnit = (a.unit || '').toLowerCase()
    const bUnit = (b.unit || '').toLowerCase()
    const cmp = aUnit.localeCompare(bUnit)
    return sortUnitDir.value === 'asc' ? cmp : -cmp
  })
  return sorted
})
const totalListPages = computed(() => Math.ceil(filteredListSorted.value.length / listPageSize))
const paginatedList = computed(() => {
  if (listPage.value > totalListPages.value) listPage.value = Math.max(1, totalListPages.value)
  const start = (listPage.value - 1) * listPageSize
  return filteredListSorted.value.slice(start, start + listPageSize)
})
// Single-item form removed

// Bulk input table rows (5 fields only)
const bulkRows = ref<Array<Partial<Product>>>([
  { name: '', unit: '', stock: 0, price_investor: 0, price_shosha: 0 },
])
function addRow() {
  bulkRows.value.push({ name: '', unit: '', stock: 0, price_investor: 0, price_shosha: 0 })
}
function removeRow(index: number) {
  bulkRows.value.splice(index, 1)
  if (!bulkRows.value.length) addRow()
}
// Enhancements: paste area and CSV upload
const showPaste = ref(false)
const pasteText = ref('')

function splitCSV(line: string, delim: string): string[] {
  if (delim === '\t') return line.split('\t')
  const out: string[] = []
  let cur = ''
  let inQuotes = false
  for (let i = 0; i < line.length; i++) {
    const ch = line[i]
    if (ch === '"') {
      if (inQuotes && line[i + 1] === '"') { cur += '"'; i++ } else { inQuotes = !inQuotes }
    } else if (ch === ',' && !inQuotes) {
      out.push(cur)
      cur = ''
    } else {
      cur += ch
    }
  }
  out.push(cur)
  return out
}

function parseAndAdd(text: string) {
  // Handle both literal \n strings and real newlines
  let processed = text
    .replace(/\\n/g, '\n')
    .replace(/\\r/g, '\r')
    .trim()
  
  const lines = processed.split(/\r?\n/).filter(l => l.trim().length)
  if (!lines.length) {
    warning('Tidak ada baris yang ditemukan. Format: Nama,Satuan,Stok,Harga Investor,Harga SHOSHA')
    return
  }
  const delim = processed.includes('\t') ? '\t' : ','
  const newRows: Array<Partial<Product>> = []
  for (const line of lines) {
    const cols = splitCSV(line, delim).map(c => c.trim().replace(/^"|"$/g, ''))
    const [name, unit, stockStr, priceInvestorStr, priceSoshaStr] = cols
    if (!name || !unit) {
      console.warn(`[Parse] Skipped row: missing name or unit. Got:`, { name, unit })
      continue
    }
    const stock = parseInt(stockStr ?? '0', 10) || 0
    const priceInvestor = parseFloat(priceInvestorStr ?? '0') || 0
    const priceShosha = parseFloat(priceSoshaStr ?? '0') || 0
    
    // Validasi: minimal salah satu harga harus > 0
    if (priceInvestor <= 0 && priceShosha <= 0) {
      console.warn(`[Parse] Skipped row: at least one price must be > 0. Got:`, { name, priceInvestor, priceShosha })
      continue
    }
    
    console.log(`[Parse] ${name} | unit=${unit} stk=${stock} investor=${priceInvestor} shosha=${priceShosha}`)
    newRows.push({ name, unit, stock, price_investor: priceInvestor, price_shosha: priceShosha })
  }
  console.log(`[Parse] Total parsed rows: ${newRows.length} from ${lines.length} lines`)
  if (newRows.length) {
    const hasContent = bulkRows.value.some(r => r.name || r.unit || (r.price_investor ?? 0) > 0 || (r.price_shosha ?? 0) > 0)
    bulkRows.value = hasContent ? bulkRows.value.concat(newRows) : newRows
    success(`✓ Berhasil menambahkan ${newRows.length} baris`)
  } else {
    warning('Tidak ada baris valid. Minimal salah satu harga (Investor/SHOSHA) harus > 0')
  }
}

function applyPaste() {
  parseAndAdd(pasteText.value)
  pasteText.value = ''
  showPaste.value = false
}

function handleExcelUpload(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  
  const reader = new FileReader()
  reader.onload = (evt) => {
    try {
      const data = evt.target?.result
      const workbook = XLSX.read(data, { type: 'binary' })
      const sheetName = workbook.SheetNames[0]
      const worksheet = workbook.Sheets[sheetName]
      
      // Convert to CSV format then parse
      const csv = XLSX.utils.sheet_to_csv(worksheet)
      parseAndAdd(csv)
      input.value = ''
    } catch (err) {
      error(`Error membaca file Excel: ${(err as Error).message}`)
      input.value = ''
    }
  }
  reader.readAsBinaryString(file)
}

// Stock statistics
const lowStockThreshold = 5
const highStockThreshold = 50

const stockStats = computed(() => {
  const low = products.value.filter(p => p.stock < lowStockThreshold)
  const high = products.value.filter(p => p.stock >= highStockThreshold)
  return {
    lowCount: low.length,
    highCount: high.length,
    lowProducts: low.sort((a, b) => a.stock - b.stock),
    highProducts: high.sort((a, b) => b.stock - a.stock)
  }
})

const showStockDetails = ref(false)
const stockDetailsType = ref<'low' | 'high'>('low')

function openStockDetails(type: 'low' | 'high') {
  stockDetailsType.value = type
  showStockDetails.value = true
}

function closeStockDetails() {
  showStockDetails.value = false
}

// Format price to Rupiah display
function formatRupiah(value: number | undefined): string {
  if (!value) return ''
  return new Intl.NumberFormat('id-ID', {
    style: 'currency',
    currency: 'IDR',
    minimumFractionDigits: 0,
    maximumFractionDigits: 0
  }).format(value)
}

// Parse Rupiah string back to number (remove Rp, dots, spaces)
function parseRupiah(str: string): number {
  const cleaned = str.replace(/[^\d]/g, '')
  return cleaned ? parseInt(cleaned, 10) : 0
}
async function saveAll() {
  console.log('[ProductPanel] saveAll() triggered', { rows: bulkRows.value.length })
  saving.value = true
  try {
    // Filter valid rows: minimal nama, satuan, dan salah satu harga > 0
    console.log('[SaveAll] Total rows in bulkRows:', bulkRows.value.length)
    bulkRows.value.forEach((r, i) => {
      console.log(`[SaveAll] Row ${i}: name=${r.name}, unit=${r.unit}, investor=${r.price_investor}, shosha=${r.price_shosha}`)
    })
    const payload = bulkRows.value
      .filter(r => {
        const hasBasic = r.name?.trim() && r.unit?.trim()
        const hasPrice = (r.price_investor ?? 0) > 0 || (r.price_shosha ?? 0) > 0
        return hasBasic && hasPrice
      })
      .map(r => ({
        name: r.name!,
        unit: r.unit!,
        stock: Number(r.stock ?? 0),
        price_investor: Number(r.price_investor ?? 0),
        price_shosha: Number(r.price_shosha ?? 0)
      }))
    console.log('[ProductPanel] Filtered payload:', payload)
    if (!payload.length) {
      warning('Tidak ada baris valid. Minimal: Nama, Satuan, dan salah satu Harga (Investor/SHOSHA) > 0')
      return
    }
    console.log('[ProductPanel] Sending to API bulkCreateProducts', JSON.stringify(payload))
    // Duplicate checks
    const existingNames = new Set(products.value.map(p => (p.name || '').trim().toLowerCase()))
    const dupExisting: string[] = []
    const seenNew = new Set<string>()
    const dupWithin: string[] = []
    for (const r of payload) {
      const nm = (r.name || '').trim().toLowerCase()
      if (!nm) continue
      if (existingNames.has(nm)) dupExisting.push(r.name)
      if (seenNew.has(nm)) dupWithin.push(r.name)
      seenNew.add(nm)
    }
    if (dupExisting.length || dupWithin.length) {
      const parts = [] as string[]
      if (dupExisting.length) parts.push(`Sudah ada: ${dupExisting.join(', ')}`)
      if (dupWithin.length) parts.push(`Duplikat di input: ${dupWithin.join(', ')}`)
      error(`Duplikat nama barang terdeteksi. ${parts.join(' | ')}`)
      return
    }

    await api.bulkCreateProducts(payload as any)
    console.log('[ProductPanel] bulkCreateProducts succeeded')
    success(`✓ Berhasil menyimpan ${payload.length} produk!`)
    bulkRows.value = [{ name: '', unit: '', stock: 0, price_investor: 0, price_shosha: 0 }]
    await load()
  } catch (err) {
    const errMsg = (err as Error).message
    console.error('[SaveAll] Error:', errMsg)
    error(`Error: ${errMsg}`)
  } finally {
    saving.value = false
  }
}

async function load() {
  loading.value = true
  try {
    products.value = await api.listProducts()
    syncedInfo.value = products.value.reduce((acc, p) => {
      acc[p.id] = Boolean(p.synced)
      return acc
    }, {} as Record<string, boolean>)
  } catch (err) {
    error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function startEdit(product: Product) {
  editingProductId.value = product.id
  editForm.value = {
    name: product.name,
    unit: product.unit,
    stock: product.stock,
    price_investor: product.price_investor,
    price_shosha: product.price_shosha
  }
}

function cancelEdit() {
  editingProductId.value = null
  editForm.value = {}
}

async function saveEdit(product: Product) {
  try {
    await api.updateProduct(product.id, editForm.value)
    syncedInfo.value[product.id] = false // Mark as offline after edit
    success(`✓ ${editForm.value.name} berhasil diperbarui`)
    cancelEdit()
    await load()
  } catch (err) {
    error((err as Error).message)
  }
}

async function remove(id: string) {
  const product = products.value.find(p => p.id === id)
  if (!product) return
  
  // Use Sonner toast with action buttons for confirmation
  toast(`Hapus "${product.name}"?`, {
    description: 'Tindakan ini tidak bisa dibatalkan.',
    action: {
      label: 'Hapus',
      onClick: async () => {
        try {
          await api.deleteProduct(id)
          success(`✓ Terhapus. Menunggu sinkronisasi ke server.`)
          await load()
        } catch (err) {
          error((err as Error).message)
        }
      }
    },
    cancel: {
      label: 'Batal',
      onClick: () => {} // Do nothing on cancel
    },
    duration: 5000
  })
}

function openAdjust(product: Product, delta: number) {
  adjustingProductId.value = product.id
  adjustingDelta.value = delta
  adjustingAmount.value = 1
}

function cancelAdjust() {
  adjustingProductId.value = null
  adjustingAmount.value = 1
}

async function confirmAdjust(product: Product) {
  try {
    const deltaTotal = (adjustingAmount.value || 0) * adjustingDelta.value
    const newStock = Math.max(0, (product.stock || 0) + deltaTotal)
    // Hanya update stock, tidak touch field lain
    await api.updateProduct(product.id, { 
      stock: newStock,
      name: product.name,
      unit: product.unit,
      price_investor: product.price_investor,
      price_shosha: product.price_shosha
    })
    success(`Stok ${product.name} sekarang ${newStock}`)
    cancelAdjust()
    await load()
  } catch (err) {
    error((err as Error).message)
  }
}

// (removed unused adjustStock) use openAdjust/confirmAdjust for stock changes

onMounted(load)
</script>

<template>
  <section class="space-y-4">
    <header class="flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
      <div>
        <p class="text-sm uppercase tracking-[0.2em] text-emerald-500 font-bold">Master Barang</p>
        <h2 class="text-2xl font-semibold">Kelola barang & stok lokal</h2>
      </div>
    </header>

    <!-- Stock Statistics Cards -->
    <div class="grid gap-4 md:grid-cols-2">
      <!-- Low Stock Card -->
      <div class="relative group">
        <Card class="cursor-pointer hover:ring-2 hover:ring-red-500 transition-all" @click="openStockDetails('low')">
          <div class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-slate-600 font-medium">Stock Menipis</p>
                <p class="text-3xl font-bold text-red-500">{{ stockStats.lowCount }}</p>
                <p class="text-xs text-slate-500 mt-1">Produk dengan stok &lt; {{ lowStockThreshold }}</p>
              </div>
              <div class="text-4xl text-red-500/20">⚠️</div>
            </div>
            <p class="text-xs text-slate-400 mt-2 group-hover:text-slate-500">Sorot untuk detail</p>
          </div>
        </Card>
      </div>

      <!-- High Stock Card -->
      <div class="relative group">
        <Card class="cursor-pointer hover:ring-2 hover:ring-green-500 transition-all" @click="openStockDetails('high')">
          <div class="p-4">
            <div class="flex items-center justify-between">
              <div>
                <p class="text-sm text-slate-600 font-medium">Stock Banyak</p>
                <p class="text-3xl font-bold text-green-500">{{ stockStats.highCount }}</p>
                <p class="text-xs text-slate-500 mt-1">Produk dengan stok ≥ {{ highStockThreshold }}</p>
              </div>
              <div class="text-4xl text-green-500/20">✓</div>
            </div>
            <p class="text-xs text-slate-400 mt-2 group-hover:text-slate-500">Sorot untuk detail</p>
          </div>
        </Card>
      </div>
    </div>

    <div class="grid gap-4 lg:grid-cols-[1fr]">

      <Card>
        <div class="p-4">
          <div class="flex items-center justify-between">
            <p class="text-sm font-bold">Input Massal (seperti Excel)</p>
            <div class="flex items-center gap-2">
              <Button variant="ghost" class="text-xs" @click="addRow">Tambah Baris</Button>
              <Button class="text-xs" :disabled="saving" @click="saveAll">Simpan Semua</Button>
              <Button variant="ghost" class="text-xs" @click="showPaste = !showPaste">Paste Excel/CSV</Button>
              <label class="text-xs cursor-pointer inline-flex items-center gap-2">
                <input type="file" accept=".xlsx,.xls" @change="handleExcelUpload" class="hidden" />
                <span class="px-2 py-1 rounded ">Import Excel</span>
              </label>
            </div>
          </div>
          <div v-if="showPaste" class="mt-3 space-y-2">
            <p class="text-xs text-slate-400">Tempel baris dari Excel. Urutan kolom: Nama, Satuan, Stok, Harga Investor, Harga SHOSHA. Pisahkan dengan TAB atau koma.</p>
            <textarea v-model="pasteText" rows="5" class="w-full rounded  p-2 text-sm" placeholder="Contoh:\nSabun,pcs,10,4500,4000\nBeras,kg,20,58000,55000"></textarea>
            <div class="flex items-center gap-2">
              <Button class="text-xs" @click="applyPaste">Tambahkan</Button>
              <Button variant="ghost" class="text-xs" @click="showPaste = false">Batal</Button>
            </div>
          </div>
          <div class="mt-3 overflow-auto">
            <table class="w-full text-sm">
              <thead>
                <tr class="text-left text-slate-400">
                  <th class="px-2 py-2">Nama</th>
                  <th class="px-2 py-2">Satuan</th>
                  <th class="px-2 py-2">Stok</th>
                  <th class="px-2 py-2">Harga Investor</th>
                  <th class="px-2 py-2">Harga SHOSHA</th>
                  <th class="px-2 py-2">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, idx) in bulkRows" :key="idx" class="">
                  <td class="px-2 py-1"><Input v-model="row.name" placeholder="Nama" /></td>
                  <td class="px-2 py-1"><Input v-model="row.unit" placeholder="kg, pcs, liter" /></td>
                  <td class="px-2 py-1"><Input v-model="row.stock" type="number" min="0" /></td>
                  <td class="px-2 py-1">
                    <Input 
                      :value="formatRupiah(row.price_investor)" 
                      @input="(e: InputEvent) => { row.price_investor = parseRupiah((e.target as HTMLInputElement).value) }"
                      placeholder="Rp 0"
                      class="w-full rounded  px-2 py-1 text-sm outline-slate-200 placeholder:text-slate-500"
                    />
                  </td>
                  <td class="px-2 py-1">
                    <Input 
                      :value="formatRupiah(row.price_shosha)" 
                      @input="(e: InputEvent) => { row.price_shosha = parseRupiah((e.target as HTMLInputElement).value) }"
                      placeholder="Rp 0"
                      class="w-full rounded  px-2 py-1 text-sm outline-slate-200 placeholder:text-slate-500"
                    />
                  </td>
                  <td class="px-2 py-1"><Button variant="ghost" class="text-xs text-rose-200" @click="removeRow(idx)">Hapus</Button></td>
                </tr>
              </tbody>
            </table>
          </div>
          <p class="mt-2 text-xs text-slate-500">Minimal: Nama, Satuan, dan salah satu Harga (Investor atau SHOSHA) harus diisi dan > 0.</p>
        </div>
      </Card>

      <Card>
        <div class="p-4">
          <div class="flex items-center justify-between">
            <p class="text-sm font-bold">Daftar Barang</p>
            <span class="text-xs text-slate-500">{{ filteredListSorted.length }} item</span>
          </div>
          <div class="mt-3 flex items-center gap-2">
            <input
              v-model="listSearch"
              placeholder="Cari barang..."
              class="w-full max-w-sm rounded  px-3 py-2 text-sm ring-1 ring-white/10 focus:ring-emerald-400"
              type="search"
            />
            <Button variant="ghost" class="text-xs" @click="sortUnitDir = sortUnitDir === 'asc' ? 'desc' : 'asc'">Sort Unit {{ sortUnitDir === 'asc' ? '↑' : '↓' }}</Button>
            <span class="text-xs text-slate-500">Hal {{ listPage }} / {{ totalListPages || 1 }}</span>
          </div>
          <div v-if="loading" class="py-6 text-sm text-slate-400">Memuat...</div>
          <div v-else class="mt-3 space-y-2">
            <div
              v-for="product in paginatedList"
              :key="product.id"
              class="rounded-xl  px-3 py-2 ring-1 ring-white/5"
            >
              <!-- View mode -->
              <template v-if="editingProductId !== product.id">
                <div class="flex items-center justify-between">
                  <div>
                    <p class="font-semibold text-black">{{ product.name }}</p>
                    <p class="text-xs text-slate-400">{{ product.unit }} • Stok {{ product.stock }}</p>
                    <p class="text-xs text-slate-400">Investor: {{ formatRupiah(product.price_investor) }} | SHOSHA: {{ formatRupiah(product.price_shosha) }}</p>
                  </div>
                  <div class="flex items-center gap-2">
                    <span
                      class="rounded-full px-2 py-1 text-[10px] uppercase tracking-wide"
                      :class="syncedInfo[product.id] ? 'bg-emerald-500 text-white' : 'bg-amber-500 text-white'"
                    >
                      {{ syncedInfo[product.id] ? 'online (synced)' : 'offline (pending sync)' }}
                    </span>
                    <template v-if="adjustingProductId !== product.id">
                      <Button variant="ghost" class="text-xs" @click.prevent="() => openAdjust(product, 1)">↑ Stok</Button>
                      <Button variant="ghost" class="text-xs" @click.prevent="() => openAdjust(product, -1)">↓ Stok</Button>
                      <Button variant="ghost" class="text-xs text-blue-300" @click="startEdit(product)">Edit</Button>
                    </template>
                    <template v-else>
                      <div class="flex items-center gap-2">
                        <input type="number" v-model.number="adjustingAmount" min="1" class="w-16 rounded  px-2 py-1 text-sm text-white" />
                        <Button size="sm" class="text-xs" @click.prevent="() => confirmAdjust(product)">OK</Button>
                        <Button variant="ghost" size="sm" class="text-xs" @click.prevent="cancelAdjust">Batal</Button>
                      </div>
                    </template>
                    <Button variant="ghost" class="text-xs text-rose-200" @click="remove(product.id)">Hapus</Button>
                  </div>
                </div>
              </template>
              <!-- Edit mode -->
              <template v-else>
                <div class="space-y-2">
                  <div class="grid grid-cols-2 gap-2">
                    <div>
                      <label class="text-xs font-bold">Nama</label>
                      <Input v-model="editForm.name" placeholder="Nama" class="mt-1" />
                    </div>
                    <div>
                      <label class="text-xs font-bold">Satuan</label>
                      <Input v-model="editForm.unit" placeholder="kg, pcs, liter" class="mt-1" />
                    </div>
                    <div>
                      <label class="text-xs font-bold">Stok</label>
                      <Input v-model.number="editForm.stock" type="number" min="0" class="mt-1" />
                    </div>
                    <div>
                      <label class="text-xs font-bold">Harga Investor</label>
                      <Input 
                        :value="formatRupiah(editForm.price_investor)" 
                        @input="(e: InputEvent) => { editForm.price_investor = parseRupiah((e.target as HTMLInputElement).value) }"
                        placeholder="Rp 0"
                        class="mt-1"
                      />
                    </div>
                    <div class="col-span-2">
                      <label class="text-xs font-bold">Harga SHOSHA</label>
                      <Input 
                        :value="formatRupiah(editForm.price_shosha)" 
                        @input="(e: InputEvent) => { editForm.price_shosha = parseRupiah((e.target as HTMLInputElement).value) }"
                        placeholder="Rp 0"
                        class="mt-1"
                      />
                    </div>
                  </div>
                  <div class="flex items-center gap-2">
                    <Button class="text-xs" @click="saveEdit(product)">Simpan</Button>
                    <Button variant="ghost" class="text-xs" @click="cancelEdit">Batal</Button>
                  </div>
                </div>
              </template>
            </div>
            <p v-if="!products.length" class="py-4 text-sm text-slate-400">Belum ada data.</p>
            <div v-else class="flex items-center justify-between py-2">
              <Button variant="ghost" size="sm" :disabled="listPage <= 1" @click="listPage--">Sebelumnya</Button>
              <div class="text-xs text-slate-500">Menampilkan {{ paginatedList.length }} dari {{ filteredListSorted.length }}</div>
              <Button variant="ghost" size="sm" :disabled="listPage >= totalListPages" @click="listPage++">Berikutnya</Button>
            </div>
          </div>
        </div>
      </Card>
    </div>

    <!-- Stock Details Modal -->
    <div v-if="showStockDetails" class="fixed inset-0 z-50 flex items-center justify-center bg-black/50" @click="closeStockDetails">
      <Card class="max-w-2xl w-full max-h-96 overflow-y-auto" @click.stop>
        <div class="p-6 space-y-4">
          <div class="flex items-center justify-between border-b pb-3 sticky top-0 bg-white">
            <h3 class="text-lg font-semibold text-slate-900">
              {{ stockDetailsType === 'low' ? 'Produk Stock Menipis' : 'Produk Stock Banyak' }}
            </h3>
            <Button variant="ghost" size="sm" @click="closeStockDetails">✕</Button>
          </div>

          <div class="space-y-2">
            <div v-if="stockDetailsType === 'low' && stockStats.lowProducts.length === 0" class="py-6 text-center text-slate-400">
              Tidak ada produk dengan stock menipis
            </div>
            <div v-else-if="stockDetailsType === 'high' && stockStats.highProducts.length === 0" class="py-6 text-center text-slate-400">
              Tidak ada produk dengan stock banyak
            </div>
            <div v-else class="overflow-x-auto">
              <table class="w-full text-sm">
                <thead class="bg-slate-50 sticky top-0">
                  <tr>
                    <th class="px-3 py-2 text-left">Nama Produk</th>
                    <th class="px-3 py-2 text-center">Stok</th>
                    <th class="px-3 py-2 text-left">Satuan</th>
                    <th class="px-3 py-2 text-right">Harga Investor</th>
                    <th class="px-3 py-2 text-right">Harga SHOSHA</th>
                  </tr>
                </thead>
                <tbody class="divide-y">
                  <tr 
                    v-for="product in (stockDetailsType === 'low' ? stockStats.lowProducts : stockStats.highProducts)" 
                    :key="product.id"
                    :class="stockDetailsType === 'low' ? 'bg-red-50' : 'bg-green-50'"
                  >
                    <td class="px-3 py-2 font-medium text-slate-900">{{ product.name }}</td>
                    <td class="px-3 py-2 text-center">
                      <span :class="[
                        'font-bold px-2 py-1 rounded text-sm',
                        stockDetailsType === 'low' ? 'bg-red-200 text-red-800' : 'bg-green-200 text-green-800'
                      ]">
                        {{ product.stock }}
                      </span>
                    </td>
                    <td class="px-3 py-2 text-slate-600">{{ product.unit }}</td>
                    <td class="px-3 py-2 text-right">{{ formatRupiah(product.price_investor) }}</td>
                    <td class="px-3 py-2 text-right">{{ formatRupiah(product.price_shosha) }}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="flex justify-end gap-2 pt-3 border-t">
            <Button @click="closeStockDetails">Tutup</Button>
          </div>
        </div>
      </Card>
    </div>  </section>
</template>