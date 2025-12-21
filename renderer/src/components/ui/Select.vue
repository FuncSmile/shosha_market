<script setup lang="ts">
const props = defineProps<{
  modelValue?: string
  disabled?: boolean
}>()
const emit = defineEmits<{ (e: 'update:modelValue', value: string): void }>()

function onChange(event: Event) {
  const target = event.target as HTMLSelectElement
  emit('update:modelValue', target.value)
}
</script>

<template>
  <select
    :value="props.modelValue"
    :disabled="props.disabled"
    class="w-full rounded-lg bg-white px-3 py-2 text-sm text-slate-900 ring-1 ring-slate-200 focus:ring-emerald-600 disabled:opacity-60"
    @change="onChange"
  >
    <slot />
  </select>
</template>
