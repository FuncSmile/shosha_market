<script setup lang="ts">
import { computed, ref } from 'vue'
import DashboardPanel from './components/DashboardPanel.vue'
import ProductPanel from './components/ProductPanel.vue'
import SalesPanel from './components/SalesPanel.vue'
import BranchPanel from './components/BranchPanel.vue'
import StockOpnamePanel from './components/StockOpnamePanel.vue'
import ReportsPanel from './components/ReportsPanel.vue'

const tabs = [
  { key: 'dashboard', label: 'Dashboard' },
  { key: 'products', label: 'Barang' },
  { key: 'sales', label: 'Penjualan' },
  { key: 'branches', label: 'Cabang' },
  { key: 'opname', label: 'Stock Opname' },
  { key: 'reports', label: 'Laporan' },
]

const active = ref('dashboard')
const components = {
  dashboard: DashboardPanel,
  products: ProductPanel,
  sales: SalesPanel,
  branches: BranchPanel,
  opname: StockOpnamePanel,
  reports: ReportsPanel,
}
const current = computed(() => components[active.value as keyof typeof components])

function go(key: string) {
  active.value = key
}
</script>

<template>
  <div class="flex min-h-screen bg-slate-950 text-slate-100">
    <aside class="flex w-72 flex-col border-r border-white/5 bg-slate-900/70 px-4 py-6">
      <div class="mb-8 rounded-2xl bg-emerald-500/10 p-4 ring-1 ring-emerald-400/20">
        <p class="text-xs uppercase tracking-[0.3em] text-emerald-200/80">Shosha Mart</p>
        <h1 class="text-xl font-semibold text-white">POS Offline-first</h1>
        <p class="text-xs text-emerald-100/80">Electron + Vue (TS) + Go sidecar</p>
      </div>

      <nav class="flex-1 space-y-1">
        <button
          v-for="tab in tabs"
          :key="tab.key"
          class="flex w-full items-center justify-between rounded-xl px-3 py-2 text-sm font-semibold transition"
          :class="active === tab.key ? 'bg-emerald-500 text-emerald-50 shadow-lg shadow-emerald-500/20' : 'bg-transparent text-slate-200 hover:bg-slate-800/80'"
          @click="go(tab.key)"
        >
          <span>{{ tab.label }}</span>
          <span v-if="active === tab.key" class="h-2 w-2 rounded-full bg-emerald-100" />
        </button>
      </nav>

      <div class="mt-6 rounded-xl bg-slate-800/60 p-4 text-xs text-slate-300 ring-1 ring-white/5">
        <p class="font-semibold text-white">Offline-first</p>
        <p class="mt-1 text-slate-400">SQLite lokal, flag synced untuk antrean sinkron ke cloud PostgreSQL.</p>
      </div>
    </aside>

    <main class="flex-1 overflow-y-auto bg-gradient-to-br from-slate-950 via-slate-900 to-slate-950 px-8 py-8">
      <header class="mb-6 flex flex-wrap items-center justify-between gap-4 rounded-3xl bg-slate-900/70 p-6 ring-1 ring-white/5">
        <div>
          <p class="text-xs uppercase tracking-[0.3em] text-emerald-200/80">Control Panel</p>
          <p class="text-sm text-slate-300">Kelola data, POS, opname, dan laporan.</p>
        </div>
        <span class="rounded-full bg-emerald-500/20 px-3 py-1 text-xs font-semibold text-emerald-100">Sidecar: 127.0.0.1:8080</span>
      </header>

      <div class="rounded-2xl border border-white/5 bg-slate-900/60 p-6 shadow-xl shadow-emerald-500/5">
        <component :is="current" @navigate="go" />
      </div>
    </main>
  </div>
</template>
