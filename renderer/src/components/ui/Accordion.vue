<script setup lang="ts">
import { AccordionRoot } from 'radix-vue'
import type { AccordionRootProps } from 'radix-vue'

const props = defineProps<AccordionRootProps>()
</script>

<template>
  <AccordionRoot v-bind="props">
    <slot />
  </AccordionRoot>
</template>
