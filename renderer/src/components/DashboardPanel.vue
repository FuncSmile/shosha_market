<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { api, type SalesAnalytics } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Select from './ui/Select.vue'
import Label from './ui/Label.vue'

const emit = defineEmits<{ (e: 'navigate', key: string): void }>()

const loading = ref(false)
const message = ref('')
const analytics = ref<SalesAnalytics | null>(null)
const filters = reactive({
  start: '',
  end: '',
  sort: 'asc',
})

const shortcuts = [
  { label: 'Tambah Barang', key: 'products' },
  { label: 'Checkout Cepat', key: 'sales' },
  { label: 'Stock Opname', key: 'opname' },
  { label: 'Export Laporan', key: 'reports' },
]

const perDaySorted = computed(() => {
  if (!analytics.value) return []
  return [...analytics.value.perDay].sort((a, b) =>
    filters.sort === 'asc' ? a.day.localeCompare(b.day) : b.day.localeCompare(a.day),
  )
})

const maxRevenue = computed(() => {
  const list = perDaySorted.value
  return list.length ? Math.max(...list.map((d) => d.revenue)) || 1 : 1
})

async function loadAnalytics() {
  loading.value = true
  message.value = ''
  try {
    analytics.value = await api.salesAnalytics(filters.start, filters.end)
    if (!filters.start) filters.start = analytics.value.start
    if (!filters.end) filters.end = analytics.value.end
  } catch (err) {
    message.value = (err as Error).message
  } finally {
    loading.value = false
  }
}

onMounted(loadAnalytics)
</script>

<template>
<section class="space-y-5">
  <div class="grid gap-4 lg:grid-cols-12">
    <Card class="lg:col-span-7">
      <div class="p-5 space-y-4">
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-[11px] uppercase tracking-[0.3em] text-emerald-200/80">Shadcn Vue</p>
            <h2 class="text-2xl font-semibold text-white">Offline-first POS</h2>
            <p class="text-sm text-slate-300">
              Electron + Vue (TS) + Tailwind + shadcn-inspired UI. Sidecar Go + SQLite, sinkron ke PostgreSQL.
            </p>
          </div>
          <span v-if="message" class="rounded-full bg-rose-500/20 px-3 py-1 text-xs text-rose-100">{{ message }}</span>
        </div>

        <div class="grid gap-3 md:grid-cols-2">
          <div class="space-y-2">
            <Label>Filter tanggal</Label>
            <div class="grid grid-cols-2 gap-2">
              <Input v-model="filters.start" type="date" @change="loadAnalytics" />
              <Input v-model="filters.end" type="date" @change="loadAnalytics" />
            </div>
          </div>
          <div class="grid gap-2 md:grid-cols-[1fr_auto]">
            <div>
              <Label>Urut</Label>
              <Select v-model="filters.sort" @change="loadAnalytics">
                <option value="asc">Tanggal naik</option>
                <option value="desc">Tanggal turun</option>
              </Select>
            </div>
            <div class="flex items-end">
              <Button variant="outline" class="w-full" @click="loadAnalytics">Refresh</Button>
            </div>
          </div>
        </div>

        <div class="flex flex-wrap gap-2">
          <Button
            v-for="shortcut in shortcuts"
            :key="shortcut.key"
            variant="ghost"
            class="border border-white/5"
            @click="emit('navigate', shortcut.key)"
          >
            {{ shortcut.label }}
          </Button>
        </div>
      </div>
    </Card>

    <Card class="lg:col-span-5">
      <div class="grid gap-3 p-4 md:grid-cols-3 sm:grid-cols-2">
        <div class="rounded-xl border border-white/5 bg-slate-900/70 p-3">
          <p class="text-xs uppercase tracking-wide text-slate-400">Nilai Penjualan</p>
          <p class="mt-2 text-2xl font-semibold text-white">
            Rp{{ (analytics?.totalRevenue ?? 0).toLocaleString('id-ID') }}
          </p>
          <p class="text-[11px] text-slate-500">Rentang {{ filters.start || '...' }} s.d {{ filters.end || '...' }}</p>
        </div>
        <div class="rounded-xl border border-white/5 bg-slate-900/70 p-3">
          <p class="text-xs uppercase tracking-wide text-slate-400">Jumlah Order</p>
          <p class="mt-2 text-2xl font-semibold text-white">{{ analytics?.totalOrders ?? 0 }}</p>
          <p class="text-[11px] text-slate-500">Transaksi tercatat di SQLite</p>
        </div>
        <div class="rounded-xl border border-white/5 bg-slate-900/70 p-3">
          <p class="text-xs uppercase tracking-wide text-slate-400">Item Terjual</p>
          <p class="mt-2 text-2xl font-semibold text-white">{{ analytics?.totalItems ?? 0 }}</p>
          <p class="text-[11px] text-slate-500">Flag synced menunggu upload</p>
        </div>
      </div>
    </Card>
  </div>

  <Card>
    <div class="flex flex-wrap items-center justify-between gap-3 border-b border-white/5 px-6 py-4">
      <div>
        <p class="text-[11px] uppercase tracking-[0.3em] text-emerald-200/80">Analitik Harian</p>
        <p class="text-sm text-slate-300">Performa penjualan per tanggal</p>
      </div>
      <span v-if="loading" class="text-xs text-slate-400">Memuat...</span>
    </div>
    <div class="grid gap-6 p-6 lg:grid-cols-[2fr_1.2fr]">
      <div>
        <p class="text-xs text-slate-400">Grafik Revenue</p>
        <div class="mt-3 flex items-end gap-3 overflow-x-auto pb-2">
          <div
            v-for="row in perDaySorted"
            :key="row.day"
            class="flex min-w-[68px] flex-col items-center gap-1"
          >
            <div
              class="w-full rounded-t-md bg-gradient-to-t from-emerald-600 via-emerald-400 to-cyan-300 shadow-md shadow-emerald-500/30"
              :style="{ height: `${Math.max(10, (row.revenue / maxRevenue) * 200)}px` }"
            ></div>
            <p class="text-[11px] text-slate-300">{{ row.day.slice(5) }}</p>
            <p class="text-[11px] text-emerald-100">Rp{{ row.revenue.toLocaleString('id-ID') }}</p>
          </div>
          <p v-if="!perDaySorted.length" class="text-sm text-slate-400">Belum ada data penjualan.</p>
        </div>
      </div>

      <div class="overflow-hidden rounded-xl border border-white/5 bg-slate-900/60">
        <table class="min-w-full divide-y divide-white/5 text-sm">
          <thead class="bg-slate-900/80">
            <tr>
              <th class="px-4 py-2 text-left text-slate-300">Tanggal</th>
              <th class="px-4 py-2 text-right text-slate-300">Order</th>
              <th class="px-4 py-2 text-right text-slate-300">Item</th>
              <th class="px-4 py-2 text-right text-slate-300">Revenue</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="row in perDaySorted" :key="row.day" class="hover:bg-slate-800/40">
              <td class="px-4 py-2 text-slate-100">{{ row.day }}</td>
              <td class="px-4 py-2 text-right text-slate-100">{{ row.orders }}</td>
              <td class="px-4 py-2 text-right text-slate-100">{{ row.items }}</td>
              <td class="px-4 py-2 text-right text-emerald-100">Rp{{ row.revenue.toLocaleString('id-ID') }}</td>
            </tr>
            <tr v-if="!perDaySorted.length">
              <td colspan="4" class="px-4 py-4 text-center text-slate-400">Belum ada data.</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </Card>
</section>
</template>
