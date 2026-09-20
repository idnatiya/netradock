<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ state: string; label?: string; dot?: boolean }>()

// Tabler status colors; the text always names the state, so color is never the only signal.
const color = computed(() => ({ running: 'green', exited: 'red', dead: 'red', paused: 'yellow', restarting: 'yellow' })[props.state] ?? 'secondary')
const text = computed(() => props.label ?? props.state)
</script>

<template>
  <!-- Docker Desktop's list shows the dot alone; the words stay available as a tooltip. -->
  <span v-if="dot" class="status status-dot-only" :class="`status-${color}`" :title="text">
    <span class="status-dot" :class="{ 'status-dot-animated': state === 'restarting' }"></span>
    <span class="visually-hidden">{{ text }}</span>
  </span>
  <span v-else class="status" :class="`status-${color}`">
    <span class="status-dot" :class="{ 'status-dot-animated': state === 'restarting' }"></span>
    {{ text }}
  </span>
</template>

<style scoped>
.status { white-space: normal; height: auto; min-height: var(--tblr-status-height); text-align: left; }
.status-dot-only { min-height: 0; padding: 0; background: transparent; }
</style>
