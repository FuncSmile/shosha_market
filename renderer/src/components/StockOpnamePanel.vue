<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type Product } from '../api'
import { useToast } from '../composables/useToast'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'

const { success, error } = useToast()
const products = ref<Product[]>([])
const saving = ref(false)
const form = reactive({
  branch_id: '',
  note: '',
  items: [] as { product_id: string; qty_system: number; qty_physical: number }[],
})

async function loadProducts() {
  products.value = await api.listProducts()
  form.items = products.value.map((p) => ({
    product_id: p.id,
    qty_system: p.stock,
    qty_physical: p.stock,
  }))
}

function updatePhysical(idx: number, delta: number) {
  const item = form.items[idx]
  if (!item) return
  item.qty_physical = Math.max(0, item.qty_physical + delta)
}

async function submit() {
  saving.value = true
  try {
    await api.createStockOpname(form)
    success('Opname tersimpan & stok disesuaikan.')
  } catch (err) {
    error((err as Error).message)
  } finally {
    saving.value = false
  }
}

onMounted(loadProducts)
</script>

<template>
  <section class="space-y-4">
    <header class="flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
      <div>
        <p class="text-sm uppercase tracking-[0.2em] text-emerald-500 font-bold">Stock Opname</p>
        <h2 class="text-2xl font-semibold">Sesuaikan stok fisik dan export Excel</h2>
      </div>
    </header>

    <Card>
      <form class="space-y-4 p-4" @submit.prevent="submit">
        <div class="grid gap-3 sm:grid-cols-[240px_1fr]">
          <div class="space-y-1">
            <Label>ID Cabang</Label>
            <Input v-model="form.branch_id" required />
          </div>
          <div class="space-y-1">
            <Label>Catatan</Label>
            <Input v-model="form.note" />
          </div>
        </div>

        <Card>
          <div class="flex items-center justify-between px-4 py-3">
            <p class="text-sm">Hitung stok</p>
            <p class="text-xs text-slate-500">{{ form.items.length }} barang</p>
          </div>
          <div class="space-y-2 p-3">
            <div
              v-for="(item, idx) in form.items"
              :key="item.product_id"
              class="grid gap-2 rounded-lg  p-3 ring-1 ring-white/5 sm:grid-cols-[1.5fr_1fr_1fr_auto]"
            >
              <div>
                <p class="text-sm font-semibold">
                  {{ products.find((p) => p.id === item.product_id)?.name ?? 'Produk' }}
                </p>
                <p class="text-xs text-slate-500">{{ products.find((p) => p.id === item.product_id)?.unit }}</p>
              </div>
              <div class="space-y-1">
                <p class="text-xs text-slate-500">Stok Sistem</p>
                <div class="rounded-lg bg-slate-100 px-3 py-2 text-sm ring-1 ring-white/10">
                  {{ item.qty_system }}
                </div>
              </div>
              <div class="space-y-1">
                <p class="text-xs text-slate-500">Stok Fisik</p>
                <div class="flex items-center gap-2">
                  <Button type="button" variant="ghost" class="px-2 py-1" @click="updatePhysical(idx, -1)">-</Button>
                  <Input v-model="item.qty_physical" type="number" min="0" />
                  <Button type="button" variant="ghost" class="px-2 py-1" @click="updatePhysical(idx, 1)">+</Button>
                </div>
              </div>
              <div class="space-y-1 text-right">
                <p class="text-xs text-slate-500">Selisih</p>
                <p class="text-sm font-semibold" :class="item.qty_physical - item.qty_system >= 0 ? 'text-emerald-200' : 'text-rose-200'">
                  {{ item.qty_physical - item.qty_system }}
                </p>
              </div>
            </div>
          </div>
        </Card>

        <div class="flex items-center gap-3">
          <Button type="submit" :disabled="saving">Simpan & update stok</Button>
          <p class="text-xs text-slate-400">
            Data disimpan di SQLite, flag synced=false, siap di-export ke Excel & dikirim saat online.
          </p>
        </div>
      </form>
    </Card>
  </section>
</template>
