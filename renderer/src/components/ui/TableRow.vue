<script setup lang="ts">
import { cn } from '@/lib/utils'

defineProps<{
  class?: string
}>()
</script>

<template>
  <tr
    :class="cn(
      'border-b border-slate-200 transition-colors hover:bg-slate-50 data-[state=selected]:bg-emerald-50',
      $props.class
    )"
  >
    <slot />
  </tr>
</template>
