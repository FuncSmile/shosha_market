<script setup lang="ts">
import { onMounted, reactive, ref, computed } from 'vue'
import { api, type Branch } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'
import Table from './ui/Table.vue'
import TableHeader from './ui/TableHeader.vue'
import TableBody from './ui/TableBody.vue'
import TableHead from './ui/TableHead.vue'
import TableRow from './ui/TableRow.vue'
import TableCell from './ui/TableCell.vue'
import Pagination from './ui/Pagination.vue'
import Select from './ui/Select.vue'

const branches = ref<Branch[]>([])
const loading = ref(false)
const message = ref('')
const syncedInfo = ref<Record<string, boolean>>({})
const form = reactive<Partial<Branch>>({
  name: '',
  code: '',
  address: '',
  phone: '',
})

// Search and sorting
const searchQuery = ref('')
const codeFilter = ref('')
const sortOrder = ref<'asc' | 'desc'>('asc')
const currentPage = ref(1)
const pageSize = 10

// Unique codes for filter dropdown
const uniqueCodes = computed(() => {
  const codes = [...new Set(branches.value.map(b => b.code).filter(Boolean))]
  return codes.sort()
})

// Computed filtered and sorted branches
const filteredBranches = computed(() => {
  let result = [...branches.value]
  
  // Code filter
  if (codeFilter.value) {
    result = result.filter(b => b.code === codeFilter.value)
  }
  
  // Search filter
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(b => 
      b.name?.toLowerCase().includes(query) ||
      b.code?.toLowerCase().includes(query) ||
      b.address?.toLowerCase().includes(query) ||
      b.phone?.toLowerCase().includes(query)
    )
  }
  
  // Sort by code
  result.sort((a, b) => {
    const codeA = (a.code || '').toLowerCase()
    const codeB = (b.code || '').toLowerCase()
    return sortOrder.value === 'asc' 
      ? codeA.localeCompare(codeB)
      : codeB.localeCompare(codeA)
  })
  
  return result
})

// Pagination
const totalPages = computed(() => Math.ceil(filteredBranches.value.length / pageSize))
const paginatedBranches = computed(() => {
  const start = (currentPage.value - 1) * pageSize
  const end = start + pageSize
  return filteredBranches.value.slice(start, end)
})

function toggleSort() {
  sortOrder.value = sortOrder.value === 'asc' ? 'desc' : 'asc'
}

async function load() {
  loading.value = true
  try {
    branches.value = await api.listBranches()
    syncedInfo.value = branches.value.reduce((acc, b) => {
      acc[b.id] = Boolean(b.synced)
      return acc
    }, {} as Record<string, boolean>)
  } catch (err) {
    message.value = (err as Error).message
  } finally {
    loading.value = false
  }
}

async function save() {
  message.value = ''
  try {
    if (form.id) {
      await api.updateBranch(form.id, form)
      message.value = 'Berhasil memperbarui cabang'
    } else {
      await api.createBranch(form)
      message.value = 'Berhasil menambah cabang'
    }
    Object.assign(form, { id: undefined, name: '', code: '', address: '', phone: '' })
    await load()
  } catch (err) {
    message.value = (err as Error).message
  }
}

function edit(item: Branch) {
  Object.assign(form, item)
  // Scroll to form
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function remove(id: string) {
  await api.deleteBranch(id)
  await load()
}

onMounted(load)
</script>

<template>
  <section class="space-y-4">
    <header class="flex flex-col justify-between gap-2 sm:flex-row sm:items-center">
      <div>
        <p class="text-sm uppercase tracking-[0.2em] text-emerald-200/80">Cabang</p>
        <h2 class="text-2xl font-semibold text-white">CRUD cabang & metadata sinkronisasi</h2>
      </div>
      <span v-if="message" class="rounded-lg bg-rose-500/20 px-3 py-1 text-sm text-rose-100">{{ message }}</span>
    </header>

    <div class="grid gap-4 lg:grid-cols-[360px_1fr]">
      <Card>
        <div class="p-4">
          <div class="flex items-center justify-between">
            <p class="text-sm text-slate-300">
              {{ form.id ? '✏️ Edit cabang' : 'Tambah cabang' }}
            </p>
            <Button
              v-if="form.id"
              variant="ghost"
              size="sm"
              @click="Object.assign(form, { id: undefined, name: '', code: '', address: '', phone: '' })"
              class="text-xs text-slate-400"
            >
              Batal Edit
            </Button>
          </div>
          <form class="mt-3 space-y-3" @submit.prevent="save">
            <div class="space-y-1">
              <Label>Kode</Label>
              <Input v-model="form.code" required />
            </div>
            <div class="space-y-1">
              <Label>Nama</Label>
              <Input v-model="form.name" required />
            </div>
            <div class="space-y-1">
              <Label>Alamat</Label>
              <textarea
                v-model="form.address"
                rows="2"
                class="w-full rounded-lg bg-slate-800/70 px-3 py-2 text-sm text-white ring-1 ring-white/10 focus:ring-emerald-400"
              />
            </div>
            <div class="space-y-1">
              <Label>Kontak</Label>
              <Input v-model="form.phone" />
            </div>
            <div class="flex items-center gap-2">
              <Button type="submit">{{ form.id ? 'Simpan Perubahan' : 'Tambah Cabang' }}</Button>
              <Button
                type="button"
                variant="ghost"
                @click="Object.assign(form, { id: undefined, name: '', code: '', address: '', phone: '' })"
              >
                Reset
              </Button>
            </div>
          </form>
        </div>
      </Card>

      <Card>
        <div class="p-4 space-y-4">
          <div class="flex items-center justify-between gap-4">
            <div class="flex items-center gap-3 flex-1">
              <Input 
                v-model="searchQuery" 
                placeholder="Cari nama, kode, alamat, atau kontak..."
                class="max-w-sm"
              />
              <div class="flex items-center gap-2">
                <Select v-model="codeFilter" class="w-40">
                  <option value="">Semua Kode</option>
                  <option v-for="code in uniqueCodes" :key="code" :value="code">
                    {{ code }}
                  </option>
                </Select>
                <Button 
                  v-if="codeFilter" 
                  variant="ghost" 
                  size="sm" 
                  @click="codeFilter = ''"
                  class="text-xs text-slate-400 hover:text-white"
                >
                  ✕
                </Button>
              </div>
            </div>
            <div class="flex items-center gap-2">
              <Button variant="ghost" size="sm" @click="toggleSort">
                <span class="text-xs">Kode {{ sortOrder === 'asc' ? '↑' : '↓' }}</span>
              </Button>
              <span class="text-xs text-slate-500">{{ filteredBranches.length }} cabang</span>
            </div>
          </div>

          <div v-if="loading" class="py-6 text-sm text-slate-400 text-center">Memuat...</div>
          <div v-else-if="!filteredBranches.length" class="py-6 text-sm text-slate-400 text-center">
            {{ searchQuery ? 'Tidak ada hasil pencarian' : 'Belum ada data.' }}
          </div>
          <div v-else class="space-y-4">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Kode</TableHead>
                  <TableHead>Nama</TableHead>
                  <TableHead>Alamat</TableHead>
                  <TableHead>Kontak</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead class="text-right">Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="branch in paginatedBranches" :key="branch.id">
                  <TableCell class="font-medium">{{ branch.code }}</TableCell>
                  <TableCell>{{ branch.name }}</TableCell>
                  <TableCell class="text-slate-400 text-xs">{{ branch.address || '-' }}</TableCell>
                  <TableCell class="text-slate-400 text-xs">{{ branch.phone || '-' }}</TableCell>
                  <TableCell>
                    <span
                      class="inline-flex rounded-full px-2 py-1 text-[10px] uppercase tracking-wide"
                      :class="syncedInfo[branch.id] ? 'bg-emerald-500/20 text-emerald-100' : 'bg-amber-500/20 text-amber-100'"
                    >
                      {{ syncedInfo[branch.id] ? 'online (synced)' : 'offline (pending sync)' }}
                    </span>
                  </TableCell>
                  <TableCell class="text-right">
                    <div class="flex items-center justify-end gap-2">
                      <Button variant="ghost" size="sm" @click="edit(branch)">Edit</Button>
                      <Button variant="ghost" size="sm" class="text-rose-200" @click="remove(branch.id)">Hapus</Button>
                    </div>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>

            <Pagination
              v-if="totalPages > 1"
              :current-page="currentPage"
              :total-pages="totalPages"
              :page-size="pageSize"
              :total-items="filteredBranches.length"
              @update:current-page="currentPage = $event"
            />
          </div>
        </div>
      </Card>
    </div>
  </section>
</template>
