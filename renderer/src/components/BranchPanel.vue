<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { api, type Branch } from '../api'
import Card from './ui/Card.vue'
import Button from './ui/Button.vue'
import Input from './ui/Input.vue'
import Label from './ui/Label.vue'

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
  if (form.id) {
    await api.updateBranch(form.id, form)
  } else {
    await api.createBranch(form)
  }
  Object.assign(form, { id: undefined, name: '', code: '', address: '', phone: '' })
  await load()
}

async function edit(item: Branch) {
  Object.assign(form, item)
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
          <p class="text-sm text-slate-300">Tambah / edit cabang</p>
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
        <div class="p-4">
          <div class="flex items-center justify-between">
            <p class="text-sm text-slate-300">Daftar Cabang</p>
            <span class="text-xs text-slate-500">{{ branches.length }} cabang</span>
          </div>
          <div v-if="loading" class="py-6 text-sm text-slate-400">Memuat...</div>
          <div v-else class="mt-3 grid gap-2 sm:grid-cols-2">
            <div
              v-for="branch in branches"
              :key="branch.id"
              class="flex flex-col gap-1 rounded-xl bg-slate-800/60 p-3 ring-1 ring-white/5"
            >
              <div class="flex items-center justify-between">
                <p class="text-sm font-semibold text-white">{{ branch.name }}</p>
                <span
                  class="rounded-full px-2 py-1 text-[10px] uppercase tracking-wide"
                  :class="syncedInfo[branch.id] ? 'bg-emerald-500/20 text-emerald-100' : 'bg-amber-500/20 text-amber-100'"
                >
                  {{ syncedInfo[branch.id] ? 'online (synced)' : 'offline (pending sync)' }}
                </span>
              </div>
              <p class="text-xs text-slate-400">{{ branch.code }} · {{ branch.phone }}</p>
              <p class="text-xs text-slate-500">{{ branch.address }}</p>
              <div class="flex items-center gap-3 text-xs">
                <Button variant="ghost" class="text-emerald-200" @click="edit(branch)">Edit</Button>
                <Button variant="ghost" class="text-rose-200" @click="remove(branch.id)">Hapus</Button>
              </div>
            </div>
            <p v-if="!branches.length" class="py-4 text-sm text-slate-400">Belum ada data.</p>
          </div>
        </div>
      </Card>
    </div>
  </section>
</template>
