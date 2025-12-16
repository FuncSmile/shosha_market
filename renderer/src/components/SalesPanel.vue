<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api, type Branch, type Product } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'
import Select from './ui/Select.vue'
import BranchSelect from './ui/BranchSelect.vue'

const products = ref<Product[]>([])
const branches = ref<Branch[]>([])
const message = ref('')
const saving = ref(false)
const searchProduct = ref('')
const showPrintDialog = ref(false)
const printData = ref<any>(null)

const form = reactive({
  branch_id: '',
  receipt_no: '',
  payment_method: 'cash',
  notes: '',
  jumlah_bayar: 0,
  items: [] as { product_id: string; qty: number; price: number }[],
})

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

async function loadProducts() {
  products.value = await api.listProducts()
}

async function loadBranches() {
  branches.value = await api.listBranches()
}

function addToCart(product: Product) {
  const existingItem = form.items.find((item) => item.product_id === product.id)
  if (existingItem) {
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
  if (newQty > 0) {
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
    message.value = 'Lengkapi cabang dan pilih minimal 1 barang!'
    return
  }
  saving.value = true
  message.value = ''
  try {
    const sale = await api.createSale({
      branch_id: form.branch_id,
      receipt_no: form.receipt_no || `INV-${Date.now()}`,
      payment_method: form.payment_method,
      notes: form.notes,
      items: form.items,
    })
    
    message.value = 'Transaksi berhasil disimpan!'
    
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
    message.value = (err as Error).message
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
  const customerName = (printData.value.notes || '').trim() || '-'
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
        * { margin: 0; padding: 0; }
        @page { size: A4; margin: 12mm; }
        body { 
          font-family: Arial, sans-serif; 
          font-size: 10pt;
        }
        .header { text-align: center; margin-bottom: 8px; }
        .header h1 { margin: 0; font-size: 16pt; font-weight: bold; }
        .info { margin-bottom: 6px; }
        .info table { width: 100%; border-collapse: collapse; }
        .info td { padding: 2px 0; font-size: 9pt; }
        .info td:first-child { width: 80px; font-weight: bold; }
        table.items { 
          width: 100%; 
          border-collapse: collapse; 
          margin: 6px 0;
        }
        table.items th { 
          background: #f0f0f0; 
          font-weight: bold;
          border: 1px solid #000;
          padding: 3px 2px;
          font-size: 9pt;
        }
        table.items td { 
          border: 1px solid #000; 
          padding: 2px;
          font-size: 9pt;
        }
        .total-row { 
          font-weight: bold;
          text-align: right;
          padding-right: 3px;
        }
        .footer { 
          margin-top: 8px;
          page-break-inside: avoid;
        }
        .footer table { width: 100%; border-collapse: collapse; }
        .footer td { 
          text-align: center; 
          padding: 3px;
          font-size: 9pt;
          vertical-align: top;
        }
        .signature-area {
          min-height: 35px;
          border-top: 1px solid #000;
          display: inline-block;
          min-width: 90px;
          padding-top: 2px;
          margin-top: 3px;
        }
        .notes-section { padding-top: 3px; }
        .notes-section p { margin: 2px 0; font-size: 9pt; }
        .thank-you { text-align: center; font-weight: bold; margin: 6px 0; }
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
            <td>: ${printData.value.branch?.address || '-'}</td>
          </tr>
          <tr>
            <td>TANGGAL</td>
            <td>: ${date}</td>
          </tr>
          <tr>
            <td>ALAMAT</td>
            <td>: ${printData.value.branch?.name || '-'}</td>
          </tr>
        </table>
      </div>
      
      <table class="items">
        <thead>
          <tr>
            <th style="width: 25px;">NO</th>
            <th>PESANAN</th>
            <th style="width: 40px;">QTY</th>
            <th style="width: 45px;">SATUAN</th>
            <th style="width: 75px;">HARGA</th>
            <th style="width: 75px;">JUMLAH</th>
            <th style="width: 60px;">KET</th>
          </tr>
        </thead>
        <tbody>
          ${itemsHtml}
          <tr style="font-weight: bold;">
            <td colspan="4"></td>
            <td></td>
            <td style="text-align: right; padding-right: 3px;">Rp ${grandTotal.toLocaleString('id-ID')}</td>
            <td></td>
          </tr>
        </tbody>
      </table>
      
      <div class="footer">
          ${isHutang ? `
            <table>
              <tr>
                <td style="width: 50%; text-align: left; vertical-align: top;">
                  <div class="notes-section">
                    <p style="font-weight: bold; margin-bottom: 3px;">CATATAN/KETERANGAN:</p>
                    <p>${printData.value.notes || '-'}</p>
                  </div>
                </td>
                <td style="width: 25%;">
                  <p style="font-weight: bold; font-size: 8pt; margin-bottom: 80px;">PELANGGAN</p>
                  <div class="signature-area">${isHutang ? printData.value.branch?.name : customerName}</div>
                </td>
                <td style="width: 25%;">
                  <p style="font-weight: bold; font-size: 8pt; margin-bottom: 80px;">SHO-SHA MART</p>
                  <div class="signature-area">${cashierName}</div>
                </td>
              </tr>
            </table>
          ` : `
            <p class="thank-you">TERIMA KASIH</p>
            <p style="text-align: center; font-size: 9pt;">
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
        <p class="text-sm uppercase tracking-[0.2em] text-emerald-200/80">Penjualan</p>
        <h2 class="text-2xl font-semibold text-white">Point of Sale (POS)</h2>
      </div>
      <span
        v-if="message"
        :class="message.includes('berhasil') ? 'bg-emerald-500/20 text-emerald-100' : 'bg-rose-500/20 text-rose-100'"
        class="rounded-lg px-3 py-1 text-sm"
      >
        {{ message }}
      </span>
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
          <div class="grid gap-2 max-h-96 overflow-y-auto">
            <div
              v-for="product in filteredProducts"
              :key="product.id"
              class="flex items-center justify-between rounded-lg bg-slate-800/60 p-3 ring-1 ring-white/10 hover:ring-emerald-400/50 transition-all cursor-pointer"
              @click="addToCart(product)"
            >
              <div>
                <p class="text-sm font-semibold text-white">{{ product.name }}</p>
                <p class="text-xs text-slate-400">
                  Stok: {{ product.stock }} {{ product.unit }} • Rp{{ product.price.toLocaleString('id-ID') }}
                </p>
              </div>
              <Button variant="ghost" size="sm" class="text-emerald-200">+ Tambah</Button>
            </div>
            <div v-if="filteredProducts.length === 0" class="text-center py-4 text-sm text-slate-400">
              {{ searchProduct ? 'Barang tidak ditemukan' : 'Cari barang untuk menambahkan ke keranjang' }}
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
                class="flex flex-col gap-2 rounded-lg bg-slate-800/60 p-2 ring-1 ring-white/10"
              >
                <div class="flex items-start justify-between">
                  <div class="flex-1">
                    <p class="text-sm text-white font-semibold">{{ item.name }}</p>
                    <p class="text-xs text-slate-400">{{ item.unit }}</p>
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
            <p class="text-sm font-semibold text-slate-300 border-b border-slate-700 pb-3">3. PILIH METODE PEMBAYARAN</p>
            
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
