<script setup lang="ts">
import { computed } from 'vue'
import { Badge } from '@/components/ui/badge'

const props = defineProps<{ state: string; label?: string; dot?: boolean }>()

const variant = computed(() => {
  switch (props.state) {
    case 'running':
      return 'success'
    case 'exited':
    case 'dead':
      return 'destructive'
    case 'paused':
    case 'restarting':
      return 'warning'
    default:
      return 'secondary'
  }
})

const dotColor = computed(() => {
  switch (props.state) {
    case 'running':
      return 'bg-emerald-500'
    case 'exited':
    case 'dead':
      return 'bg-rose-500'
    case 'paused':
    case 'restarting':
      return 'bg-amber-500'
    default:
      return 'bg-muted-foreground'
  }
})

const text = computed(() => props.label ?? props.state)
</script>

<template>
  <span v-if="dot" class="inline-flex items-center" :title="text">
    <span
      class="inline-block size-2 rounded-full"
      :class="[dotColor, { 'animate-pulse-dot': state === 'restarting' }]"
    />
    <span class="sr-only">{{ text }}</span>
  </span>
  <Badge v-else :variant="variant" class="font-mono text-xs py-0.5 px-2 font-normal tracking-tight">
    <span
      class="inline-block size-1.5 rounded-full"
      :class="[dotColor, { 'animate-pulse-dot': state === 'restarting' }]"
    />
    <span>{{ text }}</span>
  </Badge>
</template>
