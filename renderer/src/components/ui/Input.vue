<script setup lang="ts">
const props = defineProps<{
  modelValue?: string | number
  type?: string
  placeholder?: string
}>()
const emit = defineEmits<{ (e: 'update:modelValue', value: string | number): void }>()

function onInput(event: Event) {
  const target = event.target as HTMLInputElement
  emit('update:modelValue', props.type === 'number' ? Number(target.value) : target.value)
}
</script>

<template>
  <input
    :type="props.type ?? 'text'"
    :value="props.modelValue"
    :placeholder="props.placeholder"
    class="w-full rounded-lg bg-slate-800/70 px-3 py-2 text-sm text-white ring-1 ring-white/10 focus:ring-emerald-400"
    @input="onInput"
  />
</template>
