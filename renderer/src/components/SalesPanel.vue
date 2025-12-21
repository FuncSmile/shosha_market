<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api, type Branch, type Product } from '../api'
import { useToast } from '../composables/useToast'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'
import Select from './ui/Select.vue'
import BranchSelect from './ui/BranchSelect.vue'

const { success, error, warning } = useToast()
const products = ref<Product[]>([])
const branches = ref<Branch[]>([])
const saving = ref(false)
const searchProduct = ref('')
const currentProductPage = ref(1)
const productPageSize = 5
const showPrintDialog = ref(false)
const printData = ref<any>(null)

const form = reactive({
  branch_id: '',
  // default to today (YYYY-MM-DD) for date input
  created_at: new Date().toISOString().slice(0, 10),
  receipt_no: '',
  payment_method: 'cash',
  notes: '',
  jumlah_bayar: 0,
  items: [] as { product_id: string; qty: number; price: number }[],
})

// threshold to consider stock 'low' (show warning toast)
const lowStockThreshold = 5

const total = computed(() => form.items.reduce((sum, item) => sum + (item.qty || 0) * (item.price || 0), 0))
const kembalian = computed(() => {
  if (form.payment_method !== 'cash') return 0
  return Math.max(0, form.jumlah_bayar - total.value)
})
const selectedBranch = computed(() => branches.value.find((b) => b.id === form.branch_id))
const canEditPrice = computed(() => selectedBranch.value?.code?.toLowerCase() === 'shosha')

const cartItems = computed(() => {
  return form.items.map((item) => {
    const product = products.value.find((p) => p.id === item.product_id)
    return {
      ...item,
      name: product?.name || 'Loading...',
      unit: product?.unit || 'PCS',
    }
  })
})

const isValid = computed(
  () =>
    form.branch_id &&
    form.items.length > 0 &&
    form.items.every((item) => item.product_id && item.qty > 0 && item.price >= 0) &&
    (form.payment_method === 'hutang' || form.jumlah_bayar >= total.value),
)

const filteredProducts = computed(() => {
  if (!searchProduct.value) return products.value
  const query = searchProduct.value.toLowerCase()
  return products.value.filter((p) => p.name?.toLowerCase().includes(query))
})

const totalProductPages = computed(() => Math.ceil(filteredProducts.value.length / productPageSize))
const paginatedProducts = computed(() => {
  const start = (currentProductPage.value - 1) * productPageSize
  return filteredProducts.value.slice(start, start + productPageSize)
})

async function loadProducts() {
  products.value = await api.listProducts()
}

async function loadBranches() {
  branches.value = await api.listBranches()
}

function addToCart(product: Product) {
  // legacy name kept for internal use; prefer handleAddToCart
  return handleAddToCart(product)
}

function handleAddToCart(product: Product) {
  // Prevent adding out-of-stock products
  if ((product.stock ?? 0) <= 0) {
    warning(`Stok ${product.name} kosong, tidak dapat dipilih.`)
    return
  }

  // Warn when stock is low
  if ((product.stock ?? 0) <= lowStockThreshold) {
    warning(`Stok ${product.name} menipis (${product.stock}).`)
  }

  const existingItem = form.items.find((item) => item.product_id === product.id)
  if (existingItem) {
    // don't allow exceeding stock
    const maxAllowed = product.stock ?? Infinity
    if ((existingItem.qty || 0) + 1 > maxAllowed) {
      warning(`Tidak dapat menambahkan. Stok ${product.name} hanya ${product.stock}.`)
      return
    }
    existingItem.qty++
  } else {
    form.items.push({
      product_id: product.id,
      qty: 1,
      price: product.price,
    })
  }
}

function removeItem(idx: number) {
  form.items.splice(idx, 1)
}

function updateQty(idx: number, delta: number) {
  const newQty = (form.items[idx]?.qty || 0) + delta
  const pid = form.items[idx]?.product_id
  const product = products.value.find(p => p.id === pid)
  const maxAllowed = product ? (product.stock ?? Infinity) : Infinity
  if (newQty > 0) {
    if (newQty > maxAllowed) {
      warning(`Jumlah melebihi stok. Stok ${product?.name} = ${product?.stock}`)
      return
    }
    form.items[idx].qty = newQty
  } else {
    removeItem(idx)
  }
}

function getProductName(productId: string): string {
  return products.value.find((p) => p.id === productId)?.name || 'Unknown'
}

function getProductUnit(productId: string): string {
  return products.value.find((p) => p.id === productId)?.unit || 'PCS'
}

async function submit() {
  if (!isValid.value) {
    warning('Lengkapi cabang dan pilih minimal 1 barang!')
    return
  }
  saving.value = true
  try {
    const sale = await api.createSale({
      branch_id: form.branch_id,
      receipt_no: form.receipt_no || `INV-${Date.now()}`,
      payment_method: form.payment_method,
      notes: form.notes,
      // send created_at if user selected a date (backend accepts optional created_at)
      created_at: form.created_at,
      items: form.items,
    })
    
    success('Transaksi berhasil disimpan!')
    
    // Prepare print data
    printData.value = {
      ...sale,
      branch: selectedBranch.value,
      items: form.items.map(item => ({
        ...item,
        name: getProductName(item.product_id),
        unit: getProductUnit(item.product_id),
        subtotal: item.qty * item.price,
      })),
    }
    
    // Show print dialog
    showPrintDialog.value = true
    
    // Reset form
    form.receipt_no = ''
    form.payment_method = 'cash'
    form.notes = ''
    form.items = []
  } catch (err) {
    error((err as Error).message)
  } finally {
    saving.value = false
  }
}

function printReceipt() {
  const printWindow = window.open('', '_blank')
  if (!printWindow) return
  
  const isHutang = printData.value.payment_method === 'hutang'
  const txDate = printData.value.created_at ? new Date(printData.value.created_at) : new Date()
  const date = txDate.toLocaleDateString('id-ID', {
    day: 'numeric',
    month: 'long',
    year: 'numeric'
  })
  const cashierName = 'Fadli'
  
  let itemsHtml = ''
  printData.value.items.forEach((item: any, idx: number) => {
    itemsHtml += `
      <tr>
        <td style="text-align: center; padding: 3px 2px;">${idx + 1}</td>
        <td style="padding: 3px 2px;">${item.name}</td>
        <td style="text-align: center; padding: 3px 2px;">${item.qty}</td>
        <td style="text-align: center; padding: 3px 2px;">${item.unit}</td>
        <td style="text-align: right; padding: 3px 2px;">Rp ${item.price.toLocaleString('id-ID')}</td>
        <td style="text-align: right; padding: 3px 2px;">Rp ${item.subtotal.toLocaleString('id-ID')}</td>
        <td style="padding: 3px 2px;"></td>
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
        *{margin:0;padding:0;box-sizing:border-box}
        @page{size:A4;margin:12mm}
        body{font-family:Arial,Helvetica,sans-serif;font-size:10pt;padding:8mm;color:#000}
        .header{display:flex;justify-content:center;align-items:flex-start;margin-bottom:6px}
        .title{font-size:20pt;font-weight:800;text-align:center}
        .info{margin-top:6px;margin-bottom:6px}
        .info table{width:100%;border-collapse:collapse}
        .info td{padding:2px 0;font-size:9pt}
        table.items{width:100%;border-collapse:collapse;margin-top:8px;border-top:1px solid #000;border-bottom:1px solid #000}
        table.items th{padding:4px 6px;border-left:1px solid #000;border-right:1px solid #000;border-bottom:1px solid #000;text-align:center;background:#fff;font-weight:700}
        table.items td{padding:4px 6px;border-left:1px solid #000;border-right:1px solid #000;border-bottom:none;vertical-align:top;line-height:1.2}
        table.items td.name{min-height:44px}
        table.items th.col-no,table.items td.col-no{width:40px;text-align:center}
        table.items th.col-name,table.items td.col-name{width:46%}
        table.items th.col-qty,table.items td.col-qty{width:8%;text-align:center}
        table.items th.col-unit,table.items td.col-unit{width:10%;text-align:center}
        table.items th.col-price,table.items td.col-price{width:10%;text-align:right}
        table.items th.col-sum,table.items td.col-sum{width:10%;text-align:right}
        table.items th.col-ket,table.items td.col-ket{width:8%}
        .items tfoot td{border-top:1px solid #000;padding:8px}
        .notes{margin-top:6px;display:flex;gap:8px}
        .notes .left{width:60%}
        .notes .right{width:40%;border:1px solid #000;padding:6px}
        .signatures{margin-top:24px;display:flex;justify-content:space-between}
        .signature{width:32%;text-align:center}
        .signature .line{border-top:1px solid #000;margin-top:36px}
      </style>
    </head>
    <body>
      <div class="header"><div class="title">SHO SHA MART</div></div>

      <div class="info">
        <table>
          <tr><td style="width:80px">NAMA</td><td>: ${printData.value.branch?.name || '-'}</td></tr>
          <tr><td>TANGGAL</td><td>: ${date}</td></tr>
          <tr><td>ALAMAT</td><td>: ${printData.value.branch?.address || '-'}</td></tr>
        </table>
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
            <td style="text-align: right;">&nbsp;</td>
            <td style="text-align: right; padding-right: 6px;">Rp ${grandTotal.toLocaleString('id-ID')}</td>
            <td></td>
          </tr>
        </tfoot>
      </table>

      <div class="notes">
        <div class="left">
          <div style="font-weight:bold">CATATAN/KETERANGAN:</div>
          <div style="margin-top:6px">${printData.value.notes || '-'}</div>
        </div>
        <div class="right">
          <strong>PERHATIAN:</strong>
          <ol style="margin-top:6px;padding-left:18px;font-size:9pt">
            <li>Surat Jalan ini merupakan bukti resmi penerimaan barang</li>
            <li>Surat Jalan ini bukan bukti penjualan</li>
            <li>Surat Jalan ini akan dilengkapi Invoice sebagai bukti penjualan</li>
          </ol>
        </div>
      </div>

      <div class="signatures">
        <div class="signature">
          <div>PELANGGAN</div>
          <div class="line">${isHutang ? (printData.value.branch?.name || '') : ''}</div>
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

      <script type="text/javascript">
        window.onload = function() { window.print(); window.onafterprint = function(){ window.close() } }
      <\/script>
    </body>
    </html>
  `
  
  printWindow.document.write(html)
  printWindow.document.close()
}

function closePrintDialog() {
  showPrintDialog.value = false
  printData.value = null
}

onMounted(async () => {
  await Promise.all([loadProducts(), loadBranches()])
})
</script>

<template>
  <section class="space-y-4">
    <header class="flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
      <div>
        <p class="text-sm uppercase tracking-[0.2em] text-emerald-500">Penjualan</p>
        <h2 class="text-2xl font-semibold">Point of Sale (POS)</h2>
      </div>
    </header>

    <div class="grid gap-4 lg:grid-cols-[1fr_400px]">
      <!-- Left: Product Selection -->
      <Card>
        <div class="p-4 space-y-4">
          <p class="text-sm font-semibold text-slate-300 border-b border-slate-700 pb-3">1. PILIH BARANG & TENTUKAN QTY</p>

          <!-- Search Products -->
          <div class="space-y-1">
            <Label>Cari Barang</Label>
            <Input
              v-model="searchProduct"
              placeholder="Ketik nama barang..."
              type="search"
            />
          </div>

          <!-- Product List -->
          <div class="grid gap-2">
            <div class="flex items-center justify-between text-xs">
              <span>Menampilkan {{ paginatedProducts.length }} dari {{ filteredProducts.length }}</span>
              <span>Hal {{ currentProductPage }} / {{ totalProductPages || 1 }}</span>
            </div>
            <div
              v-for="product in paginatedProducts"
              :key="product.id"
              :class="[ 'flex items-center justify-between rounded-lg p-3 ring-1 transition-all text-emerald-500/40', (product.stock ?? 0) > 0 ? 'ring-white/10 hover:ring-emerald-400/50 cursor-pointer text text-red-500/40' : 'ring-red-500/5 cursor-not-allowed opacity-60' ]"
              @click="product.stock > 0 ? addToCart(product) : null"
            >
              <div>
                <p :class="['text-sm font-semibold ', (product.stock ?? 0) > 0 ? 'text-black' : ' text-red-500/40']">{{ product.name }}</p>
                <p class="text-xs text-slate-400">
                  Stok: {{ product.stock }} {{ product.unit }} • Rp{{ product.price.toLocaleString('id-ID') }}
                </p>
              </div>
              <Button variant="ghost" size="sm" class="text-emerald-200" :disabled="(product.stock ?? 0) <= 0">+ Tambah</Button>
            </div>
            <div v-if="filteredProducts.length === 0" class="text-center py-4 text-sm text-slate-400">
              {{ searchProduct ? 'Barang tidak ditemukan' : 'Cari barang untuk menambahkan ke keranjang' }}
            </div>
            <div v-else class="flex items-center justify-between pt-2">
              <Button variant="ghost" size="sm" :disabled="currentProductPage <= 1" @click="currentProductPage--">Sebelumnya</Button>
              <div class="text-xs text-slate-500">5 per halaman</div>
              <Button variant="ghost" size="sm" :disabled="currentProductPage >= totalProductPages" @click="currentProductPage++">Berikutnya</Button>
            </div>
          </div>
        </div>
      </Card>

      <!-- Right: Cart & Checkout -->
      <div class="space-y-4">
        <!-- Step 2: Branch Selection -->
        <Card>
          <div class="p-4 space-y-3">
            <p class="text-sm font-semibold text-slate-300 border-b border-slate-700 pb-3">2. PILIH CABANG PEMBELI</p>
            <div class="space-y-1">
              <Label>Cabang</Label>
              <BranchSelect v-model="form.branch_id" :branches="branches" />
            </div>
            <div v-if="selectedBranch" class="text-xs text-slate-400">
              <p>Kode: {{ selectedBranch.code }}</p>
              <p v-if="canEditPrice" class="text-emerald-400">✓ Harga dapat diubah (cabang SHOSHA)</p>
            </div>
          </div>
        </Card>

        <!-- Cart Items -->
        <Card>
          <div class="p-4 space-y-4">
            <p class="text-sm font-semibold text-slate-300 border-b border-slate-700 pb-3">KERANJANG BELANJA</p>

            <div class="space-y-2 max-h-64 overflow-y-auto">
              <div
                v-for="(item, idx) in cartItems"
                :key="idx"
                class="flex flex-col gap-2 rounded-lg bg-emerald-600/50 p-2 ring-1 ring-white/10"
              >
                <div class="flex items-start justify-between">
                  <div class="flex-1">
                    <p class="text-sm text-white font-semibold">{{ item.name }}</p>
                    <p class="text-xs text-white">{{ item.unit }}</p>
                  </div>
                  <Button
                    variant="ghost"
                    size="sm"
                    class="text-rose-400"
                    @click="removeItem(idx)"
                  >
                    ✕
                  </Button>
                </div>
                
                <div class="flex items-center gap-2">
                  <div class="flex items-center gap-1">
                    <Button
                      variant="ghost"
                      size="sm"
                      class="h-6 w-6 p-0 text-slate-400"
                      @click="updateQty(idx, -1)"
                    >
                      −
                    </Button>
                    <input
                      :value="item.qty"
                      type="number"
                      min="1"
                      class="w-14 h-6 bg-slate-700 text-center text-white text-xs rounded border border-slate-600"
                      @change="form.items[idx].qty = Math.max(1, parseInt(($event.target as HTMLInputElement).value) || 1)"
                    />
                    <Button
                      variant="ghost"
                      size="sm"
                      class="h-6 w-6 p-0 text-slate-400"
                      @click="updateQty(idx, 1)"
                    >
                      +
                    </Button>
                  </div>
                  
                  <div class="flex-1">
                    <input
                      v-model.number="form.items[idx].price"
                      type="number"
                      :disabled="!canEditPrice"
                      :class="canEditPrice ? 'bg-slate-700' : 'bg-slate-800/50 opacity-60'"
                      class="w-full h-6 text-right text-white text-xs rounded border border-slate-600 px-2"
                      min="0"
                      step="100"
                    />
                  </div>
                </div>
                
                <div class="text-right text-xs text-emerald-400 font-semibold">
                  = Rp{{ (item.qty * item.price).toLocaleString('id-ID') }}
                </div>
              </div>
              
              <div v-if="form.items.length === 0" class="text-center py-6 text-sm text-slate-400">
                Keranjang kosong
              </div>
            </div>

            <!-- Total -->
            <div class="border-t border-slate-700 pt-3">
              <div class="flex items-center justify-between text-lg">
                <span class="text-white font-bold">TOTAL</span>
                <span class="text-emerald-400 font-bold">Rp{{ total.toLocaleString('id-ID') }}</span>
              </div>
            </div>
          </div>
        </Card>

        <!-- Step 3: Payment Method -->
        <Card>
          <div class="p-4 space-y-3">
            <p class="text-sm font-semibold text-black border-b border-slate-700 pb-3">3. PILIH METODE PEMBAYARAN</p>
            
            <div class="space-y-1">
              <Label>Metode Pembayaran</Label>
              <Select v-model="form.payment_method">
                <option value="cash">Cash (Tunai)</option>
                <option value="hutang">Hutang</option>
              </Select>
            </div>

            <!-- Cash Payment Amount (only show if payment_method is 'cash') -->
            <div v-if="form.payment_method === 'cash'" class="space-y-1">
              <Label>Jumlah Bayar</Label>
              <Input
                v-model.number="form.jumlah_bayar"
                type="number"
                placeholder="0"
                min="0"
                step="1000"
              />
            </div>

            <!-- Change Display (only show if payment_method is 'cash' and payment is valid) -->
            <div v-if="form.payment_method === 'cash'" class="rounded-lg bg-emerald-500/10 border border-emerald-400/30 p-3">
              <div class="flex items-center justify-between">
                <span class="text-sm font-semibold text-emerald-200">Kembalian</span>
                <span class="text-lg font-bold text-emerald-400">Rp{{ kembalian.toLocaleString('id-ID') }}</span>
              </div>
            </div>

            <div class="space-y-1">
              <Label>No. Invoice (Opsional)</Label>
              <Input v-model="form.receipt_no" placeholder="Auto-generate" />
            </div>

            <div class="space-y-1">
              <Label>Tanggal Transaksi</Label>
              <input
                v-model="form.created_at"
                type="date"
                class="w-full h-10 rounded-lg bg-slate-800/70 px-3 py-2 text-sm text-white ring-1 ring-white/10 focus:ring-emerald-400"
              />
            </div>

            <div v-if="form.payment_method === 'hutang'" class="space-y-1">
              <Label>Catatan (Opsional)</Label>
              <textarea
                v-model="form.notes"
                rows="2"
                placeholder="Catatan tambahan..."
                class="w-full rounded-lg bg-slate-800/70 px-3 py-2 text-sm text-white ring-1 ring-white/10 focus:ring-emerald-400"
              />
            </div>

            <!-- Checkout Button -->
            <div class="space-y-2 pt-2">
              <Button
                :disabled="saving || !isValid"
                class="w-full"
                @click="submit"
              >
                {{ saving ? 'Menyimpan...' : 'Checkout & Cetak' }}
              </Button>
              <Button
                variant="ghost"
                class="w-full text-slate-400"
                @click="form.items = []"
              >
                Bersihkan Keranjang
              </Button>
            </div>
          </div>
        </Card>
      </div>
    </div>

    <!-- Print Dialog -->
    <div
      v-if="showPrintDialog"
      class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
      @click="closePrintDialog"
    >
      <Card class="max-w-md" @click.stop>
        <div class="p-6 space-y-4">
          <h3 class="text-lg font-semibold text-white">Transaksi Berhasil!</h3>
          <p class="text-sm text-slate-300">
            {{ printData?.payment_method === 'hutang' ? 'Surat jalan siap dicetak' : 'Struk pembayaran siap dicetak' }}
          </p>
          <div class="flex gap-2">
            <Button class="flex-1" @click="printReceipt">
              Cetak {{ printData?.payment_method === 'hutang' ? 'Surat Jalan' : 'Struk' }}
            </Button>
            <Button variant="ghost" @click="closePrintDialog">
              Tutup
            </Button>
          </div>
        </div>
      </Card>
    </div>
  </section>
</template>
