<script setup lang="ts">
import { Toaster as Sonner } from 'vue-sonner'

interface ToasterProps {
  position?: 'top-left' | 'top-right' | 'bottom-left' | 'bottom-right' | 'top-center' | 'bottom-center'
  theme?: 'light' | 'dark' | 'system'
  richColors?: boolean
  expand?: boolean
  duration?: number
  gap?: number
  visibleToasts?: number
  closeButton?: boolean
  toastOptions?: {
    class?: string
    descriptionClass?: string
    style?: Record<string, string>
  }
}

const props = withDefaults(defineProps<ToasterProps>(), {
  position: 'top-right',
  theme: 'light',
  richColors: false,
  expand: true,
  duration: 3000,
  gap: 8,
  visibleToasts: 5,
  closeButton: true,
})
</script>

<template>
  <Sonner
    :position="props.position"
    :theme="props.theme"
    :rich-colors="props.richColors"
    :expand="props.expand"
    :duration="props.duration"
    :gap="props.gap"
    :visible-toasts="props.visibleToasts"
    :close-button="props.closeButton"
  />
</template>
