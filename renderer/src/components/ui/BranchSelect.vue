<script setup lang="ts">
import { computed, ref } from 'vue'
import { type Branch } from '../../api'

const props = defineProps<{
  modelValue: string
  branches: Branch[]
  disabled?: boolean
}>()
const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

const searchQuery = ref('')
const isOpen = ref(false)

const filteredBranches = computed(() => {
  if (!searchQuery.value) return props.branches
  const query = searchQuery.value.toLowerCase()
  return props.branches.filter(
    (b) =>
      b.name?.toLowerCase().includes(query) ||
      b.code?.toLowerCase().includes(query) ||
      b.address?.toLowerCase().includes(query) ||
      b.phone?.toLowerCase().includes(query)
  )
})

const selectedBranch = computed(() => {
  return props.branches.find((b) => b.id === props.modelValue)
})

function selectBranch(branch: Branch) {
  emit('update:modelValue', branch.id)
  searchQuery.value = ''
  isOpen.value = false
}

function handleInputFocus() {
  isOpen.value = true
}

function handleInputBlur() {
  setTimeout(() => {
    isOpen.value = false
  }, 200)
}
</script>

<template>
  <div class="relative">
    <input
      v-model="searchQuery"
      type="text"
      :placeholder="selectedBranch ? `${selectedBranch.name} (${selectedBranch.code})` : 'Ketik untuk cari cabang...'"
      :disabled="disabled"
      class="w-full rounded-lg bg-slate-800/70 px-3 py-2 text-sm text-white ring-1 ring-white/10 focus:ring-emerald-400 focus:outline-none disabled:opacity-60"
      @focus="handleInputFocus"
      @blur="handleInputBlur"
    />
    <div
      v-if="isOpen && filteredBranches.length > 0"
      class="absolute top-full left-0 right-0 mt-1 z-50 max-h-64 overflow-y-auto rounded-lg bg-slate-800 ring-1 ring-white/10 shadow-lg"
    >
      <div
        v-for="branch in filteredBranches"
        :key="branch.id"
        class="px-3 py-2 cursor-pointer hover:bg-slate-700/60 transition-colors"
        :class="branch.id === modelValue ? 'bg-emerald-500/20' : ''"
        @mousedown.prevent="selectBranch(branch)"
      >
        <p class="text-sm font-semibold text-white">{{ branch.name }}</p>
        <p class="text-xs text-slate-400">{{ branch.code }} • {{ branch.phone || '-' }}</p>
      </div>
    </div>
    <div v-else-if="isOpen && searchQuery && filteredBranches.length === 0" class="absolute top-full left-0 right-0 mt-1 z-50 rounded-lg bg-slate-800 ring-1 ring-white/10 p-3 text-xs text-slate-400">
      Tidak ada cabang yang cocok
    </div>
  </div>
</template>
