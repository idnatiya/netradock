<script setup lang="ts">
import { computed, useId } from 'vue'

const props = withDefaults(
  defineProps<{
    values: number[]
    capacity?: number
    floor?: number
    height?: number
    color?: string
  }>(),
  {
    capacity: 60,
    floor: 1,
    height: 36,
    color: 'var(--tblr-primary)',
  },
)

const gradId = useId()
const W = 140

function smoothSpline(points: [number, number][]): string {
  if (points.length === 0) return ''
  if (points.length === 1) return `M${points[0]![0].toFixed(1)},${points[0]![1].toFixed(1)}`
  if (points.length === 2) {
    return `M${points[0]![0].toFixed(1)},${points[0]![1].toFixed(1)}L${points[1]![0].toFixed(1)},${points[1]![1].toFixed(1)}`
  }

  let d = `M${points[0]![0].toFixed(1)},${points[0]![1].toFixed(1)}`
  const tension = 0.2

  for (let i = 0; i < points.length - 1; i++) {
    const p0 = points[i === 0 ? i : i - 1]!
    const p1 = points[i]!
    const p2 = points[i + 1]!
    const p3 = points[i + 2 < points.length ? i + 2 : i + 1]!

    const cp1x = p1[0] + (p2[0] - p0[0]) * tension
    const cp1y = p1[1] + (p2[1] - p0[1]) * tension
    const cp2x = p2[0] - (p3[0] - p1[0]) * tension
    const cp2y = p2[1] - (p3[1] - p1[1]) * tension

    d += ` C${cp1x.toFixed(1)},${cp1y.toFixed(1)} ${cp2x.toFixed(1)},${cp2y.toFixed(1)} ${p2[0].toFixed(1)},${p2[1].toFixed(1)}`
  }
  return d
}

const area = computed(() => {
  const top = Math.max(props.floor, ...props.values) * 1.15
  const offset = props.capacity - props.values.length
  const pts: [number, number][] = props.values.map((v, i) => [
    ((i + offset) / (props.capacity - 1)) * W,
    props.height - (v / top) * props.height,
  ])
  if (pts.length === 0) return { line: '', fill: '' }
  const line = smoothSpline(pts)
  const startX = pts[0]![0].toFixed(1)
  const endX = pts[pts.length - 1]![0].toFixed(1)
  return {
    line,
    fill: `${line}L${endX},${props.height}L${startX},${props.height}Z`,
  }
})
</script>

<template>
  <svg class="sparkline" :viewBox="`0 0 ${W} ${height}`" :height="height" preserveAspectRatio="none" aria-hidden="true">
    <defs>
      <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" :stop-color="color" stop-opacity="0.32" />
        <stop offset="100%" :stop-color="color" stop-opacity="0.0" />
      </linearGradient>
    </defs>
    <path :d="area.fill" :fill="`url(#${gradId})`" />
    <path :d="area.line" class="line" :stroke="color" vector-effect="non-scaling-stroke" />
  </svg>
</template>

<style scoped>
.sparkline {
  display: block;
  width: 100%;
  overflow: visible;
}
.line {
  fill: none;
  stroke-width: 1.75;
  stroke-linejoin: round;
  stroke-linecap: round;
}
</style>
