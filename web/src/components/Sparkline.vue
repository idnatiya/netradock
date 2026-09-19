<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{ values: number[]; capacity?: number; floor?: number; height?: number }>(), {
  capacity: 60,
  floor: 1,
  height: 32,
})

const W = 120
const area = computed(() => {
  const top = Math.max(props.floor, ...props.values) * 1.1
  const offset = props.capacity - props.values.length
  const pts = props.values.map((v, i) => [((i + offset) / (props.capacity - 1)) * W, props.height - (v / top) * props.height])
  if (pts.length === 0) return { line: '', fill: '' }
  const line = pts.map(([x, y], i) => `${i ? 'L' : 'M'}${x!.toFixed(1)},${y!.toFixed(1)}`).join('')
  return { line, fill: `${line}L${pts.at(-1)![0]!.toFixed(1)},${props.height}L${pts[0]![0]!.toFixed(1)},${props.height}Z` }
})
</script>

<template>
  <svg class="sparkline" :viewBox="`0 0 ${W} ${height}`" :height="height" preserveAspectRatio="none" aria-hidden="true">
    <path :d="area.fill" class="fill" />
    <path :d="area.line" class="line" vector-effect="non-scaling-stroke" />
  </svg>
</template>

<style scoped>
.sparkline { display: block; width: 100%; overflow: visible; }
.line { fill: none; stroke: var(--tblr-primary); stroke-width: 1.5; stroke-linejoin: round; }
.fill { fill: rgba(var(--tblr-primary-rgb), 0.12); }
</style>
