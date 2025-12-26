<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../api'
import type { Branch } from '../api'
import { useToast } from '../composables/useToast'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'

const { success, error } = useToast()
const start = ref('')
const end = ref('')
const opnameId = ref('')
const selectedBranchId = ref('')
const branches = ref<Branch[]>([])

onMounted(async () => {
  try {
    branches.value = await api.listBranches()
  } catch (err) {
    error((err as Error).message)
  }
})

async function download(url: string, filename: string) {
  const a = document.createElement('a')
  a.href = url
  a.download = filename
  a.click()
  URL.revokeObjectURL(url)
}

async function exportSales() {
  try {
    const url = await api.downloadSalesReport(start.value, end.value)
    await download(url, `sales_${start.value}_${end.value}.xlsx`)
    success('Laporan penjualan diunduh.')
  } catch (err) {
    error((err as Error).message)
  }
}

async function exportSalesByBranch() {
  if (!selectedBranchId.value) {
    error('Pilih cabang terlebih dahulu.')
    return
  }
  try {
    const url = await api.downloadSalesReportByBranch(selectedBranchId.value, start.value, end.value)
    await download(url, `sales_branch_${selectedBranchId.value}_${start.value}_${end.value}.xlsx`)
    success('Laporan penjualan per cabang diunduh.')
  } catch (err) {
    error((err as Error).message)
  }
}

async function exportSalesGlobal() {
  try {
    const url = await api.downloadSalesReportGlobal(start.value, end.value)
    await download(url, `sales_global_${start.value}_${end.value}.xlsx`)
    success('Laporan penjualan global (semua cabang) diunduh.')
  } catch (err) {
    error((err as Error).message)
  }
}

async function exportOpname() {
  try {
    const url = await api.downloadOpnameReport(opnameId.value)
    await download(url, `stock_opname_${opnameId.value}.xlsx`)
    success('Laporan stock opname diunduh.')
  } catch (err) {
    error((err as Error).message)
  }
}
</script>

<template>
  <section class="space-y-4">
    <header class="flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
      <div>
        <p class="text-sm uppercase tracking-[0.2em] text-emerald-500 font-bold">Laporan Excel</p>
        <h2 class="text-2xl font-semibold">Export penjualan & stock opname</h2>
      </div>
    </header>

    <div class="grid gap-4 lg:grid-cols-2">
      <!-- Export per cabang -->
      <Card>
        <div class="p-4 space-y-3">
          <p class="text-sm font-bold">Laporan Penjualan Per Cabang (.xlsx)</p>
          <div class="space-y-2">
            <Label>Pilih Cabang</Label>
            <select v-model="selectedBranchId" class="w-full px-3 py-2 border rounded-md bg-white text-slate-900">
              <option value="">-- Pilih Cabang --</option>
              <option v-for="branch in branches" :key="branch.id" :value="branch.id">
                {{ branch.name }}
              </option>
            </select>
          </div>
          <div class="grid gap-3 sm:grid-cols-2">
            <div class="space-y-1">
              <Label>Mulai</Label>
              <Input v-model="start" type="date" />
            </div>
            <div class="space-y-1">
              <Label>Selesai</Label>
              <Input v-model="end" type="date" />
            </div>
          </div>
          <Button type="button" @click="exportSalesByBranch">Export Per Cabang</Button>
          <p class="text-xs text-slate-400">
            1 sheet dengan data penjualan cabang terpilih, dikelompokkan per tanggal.
          </p>
        </div>
      </Card>

      <!-- Export global (semua cabang) -->
      <Card>
        <div class="p-4 space-y-3">
          <p class="text-sm font-bold">Laporan Penjualan Global (.xlsx)</p>
          <div class="grid gap-3 sm:grid-cols-2">
            <div class="space-y-1">
              <Label>Mulai</Label>
              <Input v-model="start" type="date" />
            </div>
            <div class="space-y-1">
              <Label>Selesai</Label>
              <Input v-model="end" type="date" />
            </div>
          </div>
          <Button type="button" @click="exportSalesGlobal">Export Global (Semua Cabang)</Button>
          <p class="text-xs text-slate-400">
            Multi-sheet Excel: 1 sheet per cabang, tiap sheet dikelompokkan per tanggal dengan kolom lebar rapi.
          </p>
        </div>
      </Card>

      <!-- Export penjualan (legacy) -->
      <Card>
        <div class="p-4 space-y-3">
          <p class="text-sm font-bold">Laporan Penjualan (Legacy) (.xlsx)</p>
          <div class="grid gap-3 sm:grid-cols-2">
            <div class="space-y-1">
              <Label>Mulai</Label>
              <Input v-model="start" type="date" />
            </div>
            <div class="space-y-1">
              <Label>Selesai</Label>
              <Input v-model="end" type="date" />
            </div>
          </div>
          <Button type="button" @click="exportSales">Export Penjualan</Button>
          <p class="text-xs text-slate-400">
            Backend Excelize menyisipkan header toko, judul periode, dan tabel ringkas (tanggal, nota, total, jumlah item).
          </p>
        </div>
      </Card>

      <!-- Export stock opname -->
      <Card>
        <div class="p-4 space-y-3">
          <p class="text-sm font-bold">Laporan Stock Opname (.xlsx)</p>
          <div class="space-y-2">
            <Label>ID Stock Opname</Label>
            <Input v-model="opnameId" placeholder="Masukkan ID" />
          </div>
          <Button type="button" @click="exportOpname">Export Opname</Button>
          <p class="text-xs text-slate-400">
            Format berisi daftar barang, stok sistem vs fisik, selisih, serta header cabang.
          </p>
        </div>
      </Card>
    </div>
  </section>
</template>
