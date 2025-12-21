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
  const base = 'rounded-lg border px-4 py-3 text-sm shadow'
  const variants: Record<string, string> = {
    default: `${base} border-slate-200 bg-white text-slate-900`,
    success: `${base} border-emerald-200 bg-emerald-50 text-emerald-800`,
    error: `${base} border-red-200 bg-red-50 text-red-800`,
    warning: `${base} border-amber-200 bg-amber-50 text-amber-800`,
    info: `${base} border-blue-200 bg-blue-50 text-blue-800`,
  }
  return variants[props.variant]
})

const iconClass = computed(() => {
  const icons: Record<string, string> = {
    success: 'text-emerald-700',
    error: 'text-red-700',
    warning: 'text-amber-700',
    info: 'text-blue-700',
    default: 'text-slate-500',
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
