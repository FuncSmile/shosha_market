<script setup lang="ts">
import { cn } from '@/lib/utils'

defineProps<{
  class?: string
}>()
</script>

<template>
  <tr
    :class="cn(
      'border-b border-slate-700 transition-colors hover:bg-slate-800/50 data-[state=selected]:bg-slate-800',
      $props.class
    )"
  >
    <slot />
  </tr>
</template>
