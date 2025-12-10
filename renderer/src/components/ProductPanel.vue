<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type Product } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'

const products = ref<Product[]>([])
const loading = ref(false)
const saving = ref(false)
const message = ref('')
const syncedInfo = ref<Record<string, boolean>>({})
const form = reactive<Partial<Product>>({
  name: '',
  sku: '',
  stock: 0,
  price: 0,
})

async function load() {
  loading.value = true
  try {
    products.value = await api.listProducts()
    syncedInfo.value = products.value.reduce((acc, p) => {
      acc[p.id] = Boolean(p.synced)
      return acc
    }, {} as Record<string, boolean>)
  } catch (err) {
    message.value = (err as Error).message
  } finally {
    loading.value = false
  }
}

async function save() {
  saving.value = true
  message.value = ''
  try {
    if (form.id) {
      await api.updateProduct(form.id, form)
    } else {
      await api.createProduct(form)
    }
    Object.assign(form, { id: undefined, name: '', sku: '', stock: 0, price: 0 })
    await load()
  } catch (err) {
    message.value = (err as Error).message
  } finally {
    saving.value = false
  }
}

async function edit(item: Product) {
  Object.assign(form, item)
}

async function remove(id: string) {
  await api.deleteProduct(id)
  await load()
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
      <span v-if="message" class="rounded-lg bg-rose-500/20 px-3 py-1 text-sm text-rose-100">{{ message }}</span>
    </header>

    <div class="grid gap-4 lg:grid-cols-[360px_1fr]">
      <Card>
        <div class="p-4">
          <p class="text-sm text-slate-300">Tambah / edit barang</p>
          <form class="mt-3 space-y-3" @submit.prevent="save">
            <div class="space-y-1">
              <Label>Nama</Label>
              <Input v-model="form.name" required />
            </div>
            <div class="space-y-1">
              <Label>SKU</Label>
              <Input v-model="form.sku" required />
            </div>
            <div class="grid grid-cols-2 gap-3">
              <div class="space-y-1">
                <Label>Stok</Label>
                <Input v-model="form.stock" type="number" min="0" />
              </div>
              <div class="space-y-1">
                <Label>Harga</Label>
                <Input v-model="form.price" type="number" min="0" step="100" />
              </div>
            </div>
            <div class="flex items-center gap-2">
              <Button type="submit" :disabled="saving">{{ form.id ? 'Simpan Perubahan' : 'Tambah Barang' }}</Button>
              <Button type="button" variant="ghost" @click="Object.assign(form, { id: undefined, name: '', sku: '', stock: 0, price: 0 })">
                Reset
              </Button>
            </div>
          </form>
        </div>
      </Card>

      <Card>
        <div class="p-4">
          <div class="flex items-center justify-between">
            <p class="text-sm text-slate-300">Daftar Barang</p>
            <span class="text-xs text-slate-500">{{ products.length }} item</span>
          </div>
          <div v-if="loading" class="py-6 text-sm text-slate-400">Memuat...</div>
          <div v-else class="mt-3 space-y-2">
            <div
              v-for="product in products"
              :key="product.id"
              class="flex items-center justify-between rounded-xl bg-slate-800/60 px-3 py-2 ring-1 ring-white/5"
            >
              <div>
                <p class="font-semibold text-white">{{ product.name }}</p>
                <p class="text-xs text-slate-400">SKU {{ product.sku }} • Stok {{ product.stock }} • Rp{{ product.price }}</p>
              </div>
              <div class="flex items-center gap-2">
                <span
                  class="rounded-full px-2 py-1 text-[10px] uppercase tracking-wide"
                  :class="syncedInfo[product.id] ? 'bg-emerald-500/20 text-emerald-100' : 'bg-amber-500/20 text-amber-100'"
                >
                  {{ syncedInfo[product.id] ? 'online (synced)' : 'offline (pending sync)' }}
                </span>
                <Button variant="ghost" class="text-xs" @click="edit(product)">Edit</Button>
                <Button variant="ghost" class="text-xs text-rose-200" @click="remove(product.id)">Hapus</Button>
              </div>
            </div>
            <p v-if="!products.length" class="py-4 text-sm text-slate-400">Belum ada data.</p>
          </div>
        </div>
      </Card>
    </div>
  </section>
</template>
