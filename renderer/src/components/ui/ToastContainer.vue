<script setup lang="ts">
import { ref } from 'vue'
import Toast from './Toast.vue'

interface ToastMessage {
  id: string
  title?: string
  description: string
  variant: 'default' | 'success' | 'error' | 'warning' | 'info'
  duration: number
  open: boolean
}

// Global state
const toasts = ref<ToastMessage[]>([])

const addToast = (message: ToastMessage) => {
  toasts.value.push(message)
}

const removeToast = (id: string) => {
  toasts.value = toasts.value.filter(t => t.id !== id)
}

const handleToastOpenChange = (id: string, open: boolean) => {
  if (!open) {
    removeToast(id)
  }
}

defineExpose({
  toasts,
  addToast,
  removeToast,
})
</script>

<template>
  <div class="fixed top-4 right-4 z-[9999] flex flex-col gap-2 pointer-events-none">
    <Toast
      v-for="toast in toasts"
      :key="toast.id"
      :id="toast.id"
      :title="toast.title"
      :description="toast.description"
      :variant="toast.variant"
      :duration="toast.duration"
      :open="toast.open"
      @open-change="handleToastOpenChange(toast.id, $event)"
      class="pointer-events-auto"
    />
  </div>
</template>
