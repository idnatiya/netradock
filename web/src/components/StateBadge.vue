<script setup lang="ts">
import { computed } from 'vue'

const props = defineProps<{ state: string; label?: string }>()

// Tabler status colors; the text always names the state, so color is never the only signal.
const color = computed(() => ({ running: 'green', exited: 'red', dead: 'red', paused: 'yellow', restarting: 'yellow' })[props.state] ?? 'secondary')
</script>

<template>
  <span class="status" :class="`status-${color}`">
    <span class="status-dot" :class="{ 'status-dot-animated': state === 'restarting' }"></span>
    {{ label ?? state }}
  </span>
</template>

<style scoped>
.status { white-space: normal; height: auto; min-height: var(--tblr-status-height); text-align: left; }
</style>
