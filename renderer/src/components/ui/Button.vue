<script setup lang="ts">
const props = defineProps<{
  variant?: 'default' | 'outline' | 'ghost'
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
}>()

const base = 'inline-flex items-center justify-center rounded-lg px-4 py-2 text-sm font-semibold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2'
const variants: Record<string, string> = {
  default: 'bg-emerald-600 text-white shadow-sm hover:bg-emerald-700 disabled:opacity-60',
  outline: 'border border-slate-200 text-slate-900 bg-white hover:bg-slate-50 disabled:opacity-60',
  ghost: 'text-slate-700 hover:bg-slate-50 disabled:opacity-60',
}
</script>

<template>
  <button :type="props.type ?? 'button'" :disabled="props.disabled" :class="[base, variants[props.variant ?? 'default']]">
    <slot />
  </button>
</template>
