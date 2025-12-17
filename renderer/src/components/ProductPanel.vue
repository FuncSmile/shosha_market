<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { api, type Product } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import { useToast } from '../composables/useToast'

const { success, error, warning } = useToast()
const products = ref<Product[]>([])
const loading = ref(false)
const saving = ref(false)
const syncedInfo = ref<Record<string, boolean>>({})
// List search + pagination (5 per page)
const listSearch = ref('')
const listPage = ref(1)
const listPageSize = 5
const filteredList = computed(() => {
  const q = listSearch.value.trim().toLowerCase()
  if (!q) return products.value
  return products.value.filter(p => p.name?.toLowerCase().includes(q) || p.unit?.toLowerCase().includes(q))
})
const totalListPages = computed(() => Math.ceil(filteredList.value.length / listPageSize))
const paginatedList = computed(() => {
  if (listPage.value > totalListPages.value) listPage.value = Math.max(1, totalListPages.value)
  const start = (listPage.value - 1) * listPageSize
  return filteredList.value.slice(start, start + listPageSize)
})
// Single-item form removed

// Bulk input table rows
const bulkRows = ref<Array<Partial<Product>>>([
  { name: '', unit: '', stock: 0, price: 0 },
])
function addRow() {
  bulkRows.value.push({ name: '', unit: '', stock: 0, price: 0 })
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
  // First replace escaped sequences: \\n, \\r\\n (from pasted text with literal backslash-n)
  let processed = text
    .replace(/\\n/g, '\n')
    .replace(/\\r/g, '\r')
    .trim()
  
  const lines = processed.split(/\r?\n/).filter(l => l.trim().length)
  if (!lines.length) {
    warning('Tidak ada baris yang ditemukan. Pastikan format: Nama,Satuan,Stok,Harga')
    return
  }
  const delim = processed.includes('\t') ? '\t' : ','
  const newRows: Array<Partial<Product>> = []
  for (const line of lines) {
    const cols = splitCSV(line, delim).map(c => c.trim().replace(/^"|"$/g, ''))
    const [name, unit, stockStr, priceStr] = cols
    if (!name || !unit) {
      console.warn(`[Parse] Skipped row: missing name or unit. Got:`, { name, unit, stock: stockStr, price: priceStr })
      continue
    }
    const stock = parseInt(stockStr ?? '0', 10) || 0
    const price = parseFloat(priceStr ?? '0') || 0
    console.log(`[Parse] ${name} | unit=${unit} stk=${stock} price=${price}`)
    newRows.push({ name, unit, stock, price })
  }
  console.log(`[Parse] Total parsed rows: ${newRows.length} from ${lines.length} lines`)
  if (newRows.length) {
    const hasContent = bulkRows.value.some(r => r.name || r.unit || (r.price ?? 0))
    bulkRows.value = hasContent ? bulkRows.value.concat(newRows) : newRows
    success(`✓ Berhasil menambahkan ${newRows.length} baris`)
  } else {
    warning('Tidak ada baris valid. Periksa Satuan dan Harga minimal > 0')
  }
}

function applyPaste() {
  parseAndAdd(pasteText.value)
  pasteText.value = ''
  showPaste.value = false
}

function handleCSVUpload(e: Event) {
  const input = e.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  const reader = new FileReader()
  reader.onload = () => {
    const text = String(reader.result || '')
    parseAndAdd(text)
    input.value = ''
  }
  reader.readAsText(file)
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
    // Filter valid rows
    console.log('[SaveAll] Total rows in bulkRows:', bulkRows.value.length)
    bulkRows.value.forEach((r, i) => {
      console.log(`[SaveAll] Row ${i}: name=${r.name}, unit=${r.unit}, price=${r.price}`)
    })
    const payload = bulkRows.value
      .filter(r => (r.name && r.unit && (r.price ?? 0) > 0))
      .map(r => ({
        name: r.name!,
        unit: r.unit!,
        stock: Number(r.stock ?? 0),
        price: Number(r.price ?? 0)
      }))
    console.log('[ProductPanel] Filtered payload:', payload)
    if (!payload.length) {
      warning('Tidak ada baris valid untuk disimpan (minimal: Nama, Satuan, Harga > 0)')
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
    bulkRows.value = [{ name: '', unit: '', stock: 0, price: 0 }]
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

// Single-item form removed; editing disabled in list

async function remove(id: string) {
  await api.deleteProduct(id)
  await load()
}

// Adjust stock by delta (positive = masuk, negative = keluar)
async function adjustStock(product: Product, delta: number) {
  try {
    const newStock = Math.max(0, (product.stock || 0) + delta)
    await api.updateProduct(product.id, { stock: newStock })
    success(`Stok ${product.name} sekarang ${newStock}`)
    await load()
  } catch (err) {
    error((err as Error).message)
  }
}

onMounted(load)
</script>

<template>
  <section class="space-y-4">
    <header class="flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
      <div>
        <p class="text-sm uppercase tracking-[0.2em] text-emerald-200/80">Master Barang</p>
        <h2 class="text-2xl font-semibold text-white">Kelola barang & stok lokal</h2>
      </div>
    </header>

    <div class="grid gap-4 lg:grid-cols-[1fr]">

      <Card>
        <div class="p-4">
          <div class="flex items-center justify-between">
            <p class="text-sm text-slate-300">Input Massal (seperti Excel)</p>
            <div class="flex items-center gap-2">
              <Button variant="ghost" class="text-xs" @click="addRow">Tambah Baris</Button>
              <Button class="text-xs" :disabled="saving" @click="saveAll">Simpan Semua</Button>
              <Button variant="ghost" class="text-xs" @click="showPaste = !showPaste">Paste Excel/CSV</Button>
              <label class="text-xs cursor-pointer inline-flex items-center gap-2">
                <input type="file" accept=".csv" @change="handleCSVUpload" class="hidden" />
                <span class="px-2 py-1 rounded bg-slate-800/60">Import CSV</span>
              </label>
            </div>
          </div>
          <div v-if="showPaste" class="mt-3 space-y-2">
            <p class="text-xs text-slate-400">Tempel baris dari Excel/CSV. Urutan kolom: Nama, Satuan, Stok, Harga. Pisahkan dengan TAB atau koma.</p>
            <textarea v-model="pasteText" rows="5" class="w-full rounded bg-slate-900/60 p-2 text-sm" placeholder="Contoh:\nSabun,pcs,10,5000\nBeras,kg,20,60000"></textarea>
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
                  <th class="px-2 py-2">Harga</th>
                  <th class="px-2 py-2">Aksi</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(row, idx) in bulkRows" :key="idx" class="odd:bg-slate-800/40">
                  <td class="px-2 py-1"><Input v-model="row.name" placeholder="Nama" /></td>
                  <td class="px-2 py-1"><Input v-model="row.unit" placeholder="kg, pcs, liter" /></td>
                  <td class="px-2 py-1"><Input v-model="row.stock" type="number" min="0" /></td>
                  <td class="px-2 py-1">
                    <input 
                      :value="formatRupiah(row.price)" 
                      @input="(e) => { row.price = parseRupiah((e.target as HTMLInputElement).value) }"
                      placeholder="Rp 0"
                      class="w-full rounded bg-slate-700 px-2 py-1 text-sm text-white placeholder:text-slate-500"
                    />
                  </td>
                  <td class="px-2 py-1"><Button variant="ghost" class="text-xs text-rose-200" @click="removeRow(idx)">Hapus</Button></td>
                </tr>
              </tbody>
            </table>
          </div>
          <p class="mt-2 text-xs text-slate-500">Minimal: Nama, Satuan, Harga</p>
        </div>
      </Card>

      <Card>
        <div class="p-4">
          <div class="flex items-center justify-between">
            <p class="text-sm text-slate-300">Daftar Barang</p>
            <span class="text-xs text-slate-500">{{ filteredList.length }} item</span>
          </div>
          <div class="mt-3 flex items-center gap-2">
            <input
              v-model="listSearch"
              placeholder="Cari barang..."
              class="w-full max-w-sm rounded bg-slate-900/60 px-3 py-2 text-sm ring-1 ring-white/10 focus:ring-emerald-400"
              type="search"
            />
            <span class="text-xs text-slate-500">Hal {{ listPage }} / {{ totalListPages || 1 }}</span>
          </div>
          <div v-if="loading" class="py-6 text-sm text-slate-400">Memuat...</div>
          <div v-else class="mt-3 space-y-2">
            <div
              v-for="product in paginatedList"
              :key="product.id"
              class="flex items-center justify-between rounded-xl bg-slate-800/60 px-3 py-2 ring-1 ring-white/5"
            >
              <div>
                <p class="font-semibold text-white">{{ product.name }}</p>
                <p class="text-xs text-slate-400">{{ product.unit }} • Stok {{ product.stock }} • {{ formatRupiah(product.price) }}</p>
              </div>
              <div class="flex items-center gap-2">
                <span
                  class="rounded-full px-2 py-1 text-[10px] uppercase tracking-wide"
                  :class="syncedInfo[product.id] ? 'bg-emerald-500/20 text-emerald-100' : 'bg-amber-500/20 text-amber-100'"
                >
                  {{ syncedInfo[product.id] ? 'online (synced)' : 'offline (pending sync)' }}
                </span>
                <Button variant="ghost" class="text-xs" @click="() => adjustStock(product, 1)">Stock Masuk</Button>
                <Button variant="ghost" class="text-xs" @click="() => adjustStock(product, -1)">Stock Keluar</Button>
                <Button variant="ghost" class="text-xs text-rose-200" @click="remove(product.id)">Hapus</Button>
              </div>
            </div>
            <p v-if="!products.length" class="py-4 text-sm text-slate-400">Belum ada data.</p>
            <div v-else class="flex items-center justify-between py-2">
              <Button variant="ghost" size="sm" :disabled="listPage <= 1" @click="listPage--">Sebelumnya</Button>
              <div class="text-xs text-slate-500">Menampilkan {{ paginatedList.length }} dari {{ filteredList.length }}</div>
              <Button variant="ghost" size="sm" :disabled="listPage >= totalListPages" @click="listPage++">Berikutnya</Button>
            </div>
          </div>
        </div>
      </Card>
    </div>
  </section>
</template>
