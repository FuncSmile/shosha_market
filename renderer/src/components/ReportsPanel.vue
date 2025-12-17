<script setup lang="ts">
import { ref } from 'vue'
import { api } from '../api'
import { useToast } from '../composables/useToast'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'

const { success, error } = useToast()
const start = ref('')
const end = ref('')
const opnameId = ref('')

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
        <p class="text-sm uppercase tracking-[0.2em] text-emerald-200/80">Laporan Excel</p>
        <h2 class="text-2xl font-semibold text-white">Export penjualan & stock opname</h2>
      </div>
    </header>

    <div class="grid gap-4 lg:grid-cols-2">
      <Card>
        <div class="p-4 space-y-3">
          <p class="text-sm text-slate-300">Laporan Penjualan (.xlsx)</p>
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

      <Card>
        <div class="p-4 space-y-3">
          <p class="text-sm text-slate-300">Laporan Stock Opname (.xlsx)</p>
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
