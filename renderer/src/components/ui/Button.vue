<script setup lang="ts">
const props = defineProps<{
  variant?: 'default' | 'outline' | 'ghost'
  type?: 'button' | 'submit' | 'reset'
  disabled?: boolean
}>()

const base = 'inline-flex items-center justify-center rounded-lg px-4 py-2 text-sm font-semibold transition focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-emerald-400'
const variants: Record<string, string> = {
  default: 'bg-emerald-500 text-emerald-50 shadow-md shadow-emerald-500/30 hover:bg-emerald-400 disabled:opacity-60',
  outline: 'border border-white/10 text-slate-100 hover:bg-slate-800 disabled:opacity-60',
  ghost: 'text-slate-200 hover:bg-slate-800/60 disabled:opacity-60',
}
</script>

<template>
  <button :type="props.type ?? 'button'" :disabled="props.disabled" :class="[base, variants[props.variant ?? 'default']]">
    <slot />
  </button>
</template>
