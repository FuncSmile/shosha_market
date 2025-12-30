<script setup lang="ts">
import { AccordionHeader, AccordionTrigger } from 'radix-vue'
import { ChevronDown } from 'lucide-vue-next'

const props = defineProps<{
  class?: string
}>()
</script>

<template>
  <AccordionHeader class="flex">
    <AccordionTrigger
      :class="['flex flex-1 items-center justify-between py-4 px-4 font-medium transition-all hover:bg-slate-50 [&[data-state=open]>svg]:rotate-180', props.class]"
    >
      <slot />
      <ChevronDown class="h-4 w-4 shrink-0 text-slate-500 transition-transform duration-200" />
    </AccordionTrigger>
  </AccordionHeader>
</template>
