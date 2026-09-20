<script setup lang="ts">
import type { HTMLAttributes } from 'vue'
import { cn } from '@/lib/utils'

const props = defineProps<{
  class?: HTMLAttributes['class']
}>()
</script>

<template>
  <thead :class="cn('[&_tr]:border-b border-border bg-muted/40', props.class)">
    <slot />
  </thead>
</template>
