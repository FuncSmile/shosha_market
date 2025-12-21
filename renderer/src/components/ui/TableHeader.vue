<script setup lang="ts">
import { cn } from '@/lib/utils'

defineProps<{
  class?: string
}>()
</script>

<template>
  <thead :class="cn('[&_tr]:border-b border-slate-200 bg-white', $props.class)">
    <slot />
  </thead>
</template>
