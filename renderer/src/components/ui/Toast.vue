<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

interface ToastProps {
  id?: string
  title?: string
  description?: string
  variant?: 'default' | 'success' | 'error' | 'warning' | 'info'
  duration?: number
  open?: boolean
}

const props = withDefaults(defineProps<ToastProps>(), {
  variant: 'default',
  duration: 3000,
  open: true,
})

const emit = defineEmits<{
  openChange: [value: boolean]
}>()

const isOpen = ref(props.open)

const variantClasses = computed(() => {
  const base = 'rounded-lg border px-4 py-3 text-sm shadow-lg'
  const variants: Record<string, string> = {
    default: `${base} border-slate-600/50 bg-slate-900/95 text-slate-100`,
    success: `${base} border-emerald-500/30 bg-gradient-to-r from-slate-900/95 to-emerald-900/20 text-emerald-100`,
    error: `${base} border-red-500/30 bg-gradient-to-r from-slate-900/95 to-red-900/20 text-red-100`,
    warning: `${base} border-amber-500/30 bg-gradient-to-r from-slate-900/95 to-amber-900/20 text-amber-100`,
    info: `${base} border-blue-500/30 bg-gradient-to-r from-slate-900/95 to-blue-900/20 text-blue-100`,
  }
  return variants[props.variant]
})

const iconClass = computed(() => {
  const icons: Record<string, string> = {
    success: 'text-emerald-400',
    error: 'text-red-400',
    warning: 'text-amber-400',
    info: 'text-blue-400',
    default: 'text-slate-400',
  }
  return icons[props.variant]
})

const getIcon = () => {
  const icons: Record<string, string> = {
    success: '✓',
    error: '✕',
    warning: '⚠',
    info: 'ℹ',
    default: '●',
  }
  return icons[props.variant]
}

onMounted(() => {
  if (props.duration && props.duration > 0) {
    setTimeout(() => {
      isOpen.value = false
      emit('openChange', false)
    }, props.duration)
  }
})

const handleClose = () => {
  isOpen.value = false
  emit('openChange', false)
}
</script>

<template>
  <transition
    name="toast-fade"
    @enter="(el: any) => el.offsetHeight"
    @leave="(el: any) => el.offsetHeight"
  >
    <div v-if="isOpen" :class="variantClasses" role="alert" class="flex items-start gap-3 max-w-sm">
      <span :class="iconClass" class="mt-0.5 flex-shrink-0 text-lg font-bold">
        {{ getIcon() }}
      </span>
      <div class="flex-1">
        <p v-if="title" class="font-semibold">{{ title }}</p>
        <p v-if="description" class="text-xs opacity-90">{{ description }}</p>
      </div>
      <button
        @click="handleClose"
        class="flex-shrink-0 ml-2 text-slate-400 hover:text-slate-100 transition-colors"
        aria-label="Close"
      >
        ✕
      </button>
    </div>
  </transition>
</template>

<style scoped>
.toast-fade-enter-active,
.toast-fade-leave-active {
  transition: all 0.3s ease;
}

.toast-fade-enter-from {
  opacity: 0;
  transform: translateX(20px);
}

.toast-fade-leave-to {
  opacity: 0;
  transform: translateX(20px);
}
</style>
