<script setup lang="ts">
import { cn } from '@/lib/utils'

defineProps<{
  class?: string
}>()
</script>

<template>
  <div class="relative w-full overflow-auto bg-white">
    <table :class="cn('w-full caption-bottom text-sm divide-y divide-slate-200', $props.class)">
      <slot />
    </table>
  </div>
</template>
