<template>
  <div class="rounded-xl border border-slate-200 bg-white p-4 ring-1 ring-slate-200 space-y-3">
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div>
        <p class="text-xs uppercase tracking-[0.2em] text-emerald-500 font-bold">Sinkronisasi</p>
        <h2 class="text-base font-semibold text-slate-900">Tarik perubahan dari Postgres ke SQLite</h2>
      </div>
      <button
        class="px-4 py-2 text-sm font-semibold rounded-lg transition disabled:opacity-50"
        :class="loading
          ? 'bg-slate-100 text-slate-700 cursor-not-allowed'
          : 'bg-emerald-600 text-white hover:bg-emerald-700 shadow-sm'"
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
        <p class="text-xs text-slate-500 mb-1">Antrian Lokal</p>
        <strong class="text-slate-900">{{ summary?.queuedChanges ?? 0 }}</strong>
      </div>
      <div>
        <p class="text-xs text-slate-500 mb-1">Terakhir Sinkron</p>
        <strong class="text-slate-700 text-xs">{{ lastSync }}</strong>
      </div>
      <div>
        <p class="text-xs text-slate-500 mb-1">Database</p>
        <code class="text-xs text-emerald-700">{{ summary?.dbPath || '-' }}</code>
      </div>
    </div>

    <div v-if="summary?.lastError" class="text-xs text-red-700 bg-red-50 rounded px-3 py-2 border border-red-200">
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
