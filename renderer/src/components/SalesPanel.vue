<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api, type Branch, type Product } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Select from './ui/Select.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'

const products = ref<Product[]>([])
const branches = ref<Branch[]>([])
const message = ref('')
const saving = ref(false)
const form = reactive({
  branch_id: '',
  receipt_no: '',
  items: [{ product_id: '', qty: 1, price: 0 }],
})

const total = computed(() => form.items.reduce((sum, item) => sum + (item.qty || 0) * (item.price || 0), 0))
const isValid = computed(
  () =>
    form.branch_id &&
    form.items.length > 0 &&
    form.items.every((item) => item.product_id && item.qty > 0 && item.price >= 0),
)

async function loadProducts() {
  products.value = await api.listProducts()
}
async function loadBranches() {
  branches.value = await api.listBranches()
  if (!form.branch_id && branches.value.length) {
    form.branch_id = branches.value[0].id
  }
}

function addRow() {
  form.items.push({ product_id: '', qty: 1, price: 0 })
}

function removeRow(idx: number) {
  form.items.splice(idx, 1)
}

function setDefaults(idx: number) {
  const item = form.items[idx]
  const product = products.value.find((p) => p.id === item.product_id)
  if (product) {
    if (!item.price || item.price === 0) item.price = product.price
    if (!item.qty || item.qty <= 0) item.qty = 1
  }
}

async function submit() {
  if (!isValid.value) {
    message.value = 'Lengkapi cabang, pilih barang, qty > 0.'
    return
  }
  saving.value = true
  message.value = ''
  try {
    await api.createSale(form)
    message.value = 'Transaksi tersimpan offline.'
    form.receipt_no = ''
    form.items = [{ product_id: '', qty: 1, price: 0 }]
  } catch (err) {
    message.value = (err as Error).message
  } finally {
    saving.value = false
  }
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
        <h2 class="text-2xl font-semibold text-white">Checkout offline, sync nanti</h2>
      </div>
      <span v-if="message" class="rounded-lg bg-emerald-500/20 px-3 py-1 text-sm text-emerald-100">{{ message }}</span>
    </header>

    <Card>
      <form class="space-y-4 p-4" @submit.prevent="submit">
        <div class="grid gap-3 sm:grid-cols-3">
          <div class="space-y-1">
            <Label>Cabang</Label>
            <Select v-model="form.branch_id" required>
              <option value="" disabled>Pilih cabang</option>
              <option v-for="branch in branches" :key="branch.id" :value="branch.id">
                {{ branch.name }} ({{ branch.id || branch.code }})
              </option>
            </Select>
          </div>
          <div class="space-y-1">
            <Label>No Nota</Label>
            <Input v-model="form.receipt_no" required />
          </div>
          <div class="space-y-1">
            <Label>Total</Label>
            <div class="rounded-lg bg-slate-800/70 px-3 py-2 text-sm font-semibold text-emerald-100 ring-1 ring-white/10">
              Rp{{ total.toLocaleString('id-ID') }}
            </div>
          </div>
        </div>

        <div class="space-y-2">
            <div class="flex items-center justify-between">
              <p class="text-sm text-slate-300">Item</p>
              <Button variant="ghost" type="button" class="px-3 py-1" @click="addRow">Tambah baris</Button>
            </div>
            <div class="space-y-2">
              <div
                v-for="(item, idx) in form.items"
                :key="idx"
                class="grid gap-2 rounded-xl bg-slate-800/60 p-3 ring-1 ring-white/10 sm:grid-cols-[1.4fr_1fr_1fr_auto]"
              >
                <Select v-model="item.product_id" required @change="setDefaults(idx)">
                  <option value="" disabled>Pilih barang</option>
                  <option v-for="product in products" :key="product.id" :value="product.id">
                    {{ product.name }} (Stok {{ product.stock }})
                  </option>
                </Select>
                <Input v-model="item.qty" type="number" min="1" />
                <Input v-model="item.price" type="number" min="0" step="100" />
                <Button type="button" variant="ghost" class="text-rose-200" @click="removeRow(idx)">Hapus</Button>
              </div>
            </div>
          </div>

        <div class="flex items-center gap-3">
          <Button type="submit" :disabled="saving || !isValid">Simpan transaksi</Button>
          <p class="text-xs text-slate-400">Disimpan ke SQLite. Flag synced=false untuk antrean upload.</p>
        </div>
      </form>
    </Card>
  </section>
</template>
