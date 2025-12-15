<template>
  <div class="rounded-xl border border-white/5 bg-slate-900/70 p-4 ring-1 ring-white/5 space-y-3">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <p class="text-xs uppercase tracking-[0.2em] text-emerald-200/80">Sinkronisasi</p>
        <h2 class="text-base font-semibold text-white">Tarik perubahan dari Postgres ke SQLite</h2>
      </div>
      <button
        class="px-4 py-2 text-sm font-semibold rounded-lg transition disabled:opacity-50"
        :class="loading
          ? 'bg-slate-700 text-slate-400 cursor-not-allowed'
          : 'bg-emerald-500 text-white hover:bg-emerald-600 shadow-lg shadow-emerald-500/20'"
        :disabled="loading"
        @click="runSync"
      >
        {{ loading ? 'Menyinkronkan…' : 'Tarik Perubahan (Pull)' }}
      </button>
    </div>

    <div class="grid grid-cols-2 md:grid-cols-4 gap-4 text-sm">
      <div>
        <p class="text-xs text-slate-400 mb-1">Status</p>
        <span :class="statusClass" class="font-semibold">{{ statusLabel }}</span>
      </div>
      <div>
        <p class="text-xs text-slate-400 mb-1">Antrian Lokal</p>
        <strong class="text-white">{{ summary?.queuedChanges ?? 0 }}</strong>
      </div>
      <div>
        <p class="text-xs text-slate-400 mb-1">Terakhir Sinkron</p>
        <strong class="text-white text-xs">{{ lastSync }}</strong>
      </div>
      <div>
        <p class="text-xs text-slate-400 mb-1">Database</p>
        <code class="text-xs text-emerald-400">{{ summary?.dbPath || '-' }}</code>
      </div>
    </div>

    <div v-if="summary?.lastError" class="text-xs text-red-400 bg-red-500/10 rounded px-3 py-2 border border-red-500/20">
      Error: {{ summary.lastError }}
    </div>
  </div>
</template>

<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue'
import { api, type SyncSummary } from '../api'

const summary = ref<SyncSummary | null>(null)
const loading = ref(false)
const status = computed(() => summary.value?.status ?? 'offline')
const statusLabel = computed(() => status.value === 'online' ? 'Online' : 'Offline')
const statusClass = computed(() => status.value === 'online' ? 'text-emerald-400' : 'text-amber-400')
const lastSync = computed(() => {
  const ls = summary.value?.lastSyncAt
  if (!ls) return '-'
  const d = new Date(ls)
  return isNaN(d.getTime()) ? ls : d.toLocaleString()
})

async function refresh() {
  try {
    summary.value = await api.syncSummary()
  } catch (e) {
    console.error('syncSummary error', e)
  }
}

async function runSync() {
  loading.value = true
  try {
    await api.syncRun()
  } catch (e) {
    alert(`Gagal sinkron: ${(e as Error).message}`)
  } finally {
    loading.value = false
    await refresh()
  }
}

let timer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  refresh()
  timer = setInterval(refresh, 5000)
})

onUnmounted(() => {
  if (timer) clearInterval(timer)
})
</script>

<style scoped>
code {
  font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, "Liberation Mono", "Courier New", monospace;
}
</style>
