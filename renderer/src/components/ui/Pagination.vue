<script setup lang="ts">
import { computed } from 'vue'
import Button from './Button.vue'

const props = defineProps<{
  currentPage: number
  totalPages: number
  pageSize: number
  totalItems: number
}>()

const emit = defineEmits<{
  'update:currentPage': [page: number]
}>()

const startItem = computed(() => (props.currentPage - 1) * props.pageSize + 1)
const endItem = computed(() => Math.min(props.currentPage * props.pageSize, props.totalItems))

function goToPage(page: number) {
  if (page >= 1 && page <= props.totalPages) {
    emit('update:currentPage', page)
  }
}
</script>

<template>
  <div class="flex items-center justify-between px-2">
    <div class="text-sm text-slate-400">
      Menampilkan {{ startItem }} - {{ endItem }} dari {{ totalItems }} cabang
    </div>
    <div class="flex items-center gap-2">
      <Button
        variant="ghost"
        size="sm"
        :disabled="currentPage === 1"
        @click="goToPage(currentPage - 1)"
      >
        Sebelumnya
      </Button>
      <div class="flex items-center gap-1">
        <button
          v-for="page in totalPages"
          :key="page"
          :class="[
            'h-8 w-8 rounded text-sm transition-colors',
            page === currentPage
              ? 'bg-emerald-500 text-white'
              : 'text-slate-400 hover:bg-slate-800 hover:text-white'
          ]"
          @click="goToPage(page)"
        >
          {{ page }}
        </button>
      </div>
      <Button
        variant="ghost"
        size="sm"
        :disabled="currentPage === totalPages"
        @click="goToPage(currentPage + 1)"
      >
        Berikutnya
      </Button>
    </div>
  </div>
</template>
