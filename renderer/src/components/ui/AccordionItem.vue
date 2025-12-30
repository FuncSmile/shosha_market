<script setup lang="ts">
import { AccordionItem } from 'radix-vue'
import type { AccordionItemProps } from 'radix-vue'

const props = defineProps<AccordionItemProps>()
</script>

<template>
  <AccordionItem v-bind="props" class="border-b border-slate-200">
    <slot />
  </AccordionItem>
</template>
    