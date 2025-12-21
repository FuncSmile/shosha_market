<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue'
import DashboardPanel from './components/DashboardPanel.vue'
import ProductPanel from './components/ProductPanel.vue'
import SalesPanel from './components/SalesPanel.vue'
import OrderHistory from './components/OrderHistory.vue'
import BranchPanel from './components/BranchPanel.vue'
import StockOpnamePanel from './components/StockOpnamePanel.vue'
import ReportsPanel from './components/ReportsPanel.vue'
import { api, type SyncSummary } from './api'
import SyncPanel from './components/SyncPanel.vue'
import Button from './components/ui/Button.vue'

const tabs = [
  { key: 'dashboard', label: 'Dashboard' },
  { key: 'products', label: 'Barang' },
  { key: 'sales', label: 'Penjualan' },
  { key: 'orders', label: 'History Order' },
  { key: 'branches', label: 'Cabang' },
  { key: 'opname', label: 'Stock Opname' },
  { key: 'reports', label: 'Laporan' },
]

const active = ref('dashboard')
const components = {
  dashboard: DashboardPanel,
  products: ProductPanel,
  sales: SalesPanel,
  orders: OrderHistory,
  branches: BranchPanel,
  opname: StockOpnamePanel,
  reports: ReportsPanel,
}
const current = computed(() => components[active.value as keyof typeof components])

function go(key: string) {
  active.value = key
}

const syncSummary = ref<SyncSummary | null>(null)
const syncing = ref(false)
let timer: ReturnType<typeof setInterval> | null = null

async function loadSummary() {
  try {
    syncSummary.value = await api.syncSummary()
  } catch (err) {
    // ignore for now
  }
}

async function runSync() {
  syncing.value = true
  try {
    await api.syncRun()
  } catch (err) {
    // ignore
  } finally {
    syncing.value = false
    await loadSummary()
  }
}

onMounted(async () => {
  // trigger sync awal supaya laptop lain bisa pull data dari upstream
  await runSync()
  timer = setInterval(loadSummary, 30000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})

// Use Toaster from vue-sonner (shadcn CLI integrates Sonner)
import { Toaster } from './components/ui/sonner'
</script>

<template>
  <Toaster position="top-right" />
  <div class="flex min-h-screen bg-gray-50 text-slate-900">
    <aside class="flex w-72 flex-col border-r border-slate-200 bg-white px-4 py-6">
      <div class="mb-8 rounded-2xl bg-emerald-50 p-4 ring-1 ring-emerald-200">
        <p class="text-xs uppercase tracking-[0.3em] text-emerald-600">Shosha Mart</p>
        <h1 class="text-xl font-semibold text-slate-900">POS Offline-first</h1>
        <p class="text-xs text-slate-600">Electron + Vue (TS) + Go sidecar</p>
      </div>

      <nav class="flex-1 space-y-1">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="flex w-full items-center justify-between rounded-xl px-3 py-2 text-sm font-semibold transition"
          :class="active === tab.key ? 'bg-emerald-600 text-white shadow-sm shadow-emerald-200' : 'bg-transparent text-slate-700 hover:bg-slate-100'"
          @click="go(tab.key)"
        >
          <span>{{ tab.label }}</span>
          <span v-if="active === tab.key" class="h-2 w-2 rounded-full bg-emerald-100" />
        </button>
      </nav>

      <div class="mt-6 rounded-xl bg-white p-4 text-xs text-slate-600 ring-1 ring-slate-200">
        <p class="font-semibold text-slate-900">Offline-first</p>
        <p class="mt-1 text-slate-600">SQLite lokal, flag synced untuk antrean sinkron ke cloud PostgreSQL.</p>
      </div>
    </aside>

    <main class="flex-1 overflow-y-auto px-8 py-8">
      <header class="mb-6 flex flex-wrap items-center justify-between gap-4 rounded-3xl bg-white p-6 ring-1 ring-slate-200">
        <div>
          <p class="text-xs uppercase tracking-[0.3em] text-emerald-500 font-bold">Control Panel</p>
          <p class="text-sm text-slate-600">Kelola data, POS, opname, dan laporan.</p>
        </div>
        <div class="flex items-center gap-3">
          <div
            class="rounded-full px-3 py-1 text-xs font-semibold"
            :class="syncSummary?.status === 'online' ? 'bg-emerald-500 text-white' : 'bg-amber-500 text-white'"
          >
            {{ syncSummary?.status ?? 'offline' }}
            <span v-if="syncSummary?.queuedChanges">· {{ syncSummary?.queuedChanges }} antrian</span>
          </div>
          <Button variant="outline" :disabled="syncing" @click="runSync">
            {{ syncing ? 'Sync...' : 'Sync sekarang' }}
          </Button>
        </div>
      </header>

      <div class="space-y-6">
        <SyncPanel />
        <div class="rounded-2xl border border-slate-200 bg-white p-6 shadow-sm">
          <component :is="current" @navigate="go" />
        </div>
      </div>
    </main>
  </div>
</template>
