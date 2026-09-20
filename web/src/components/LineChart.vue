<script setup lang="ts">
import { computed, ref, useId } from 'vue'

const props = withDefaults(
  defineProps<{
    title: string
    values: number[]
    capacity: number
    format: (v: number) => string
    floor: number
    color?: string
  }>(),
  {
    color: 'var(--tblr-primary)',
  },
)

const gradId = useId()
const glowId = useId()
const W = 600
const H = 140
const hover = ref<number | null>(null)

const top = computed(() => Math.max(props.floor, ...props.values) * 1.15)
const x = (i: number) => (i / (props.capacity - 1)) * W
const y = (v: number) => H - (v / top.value) * H
const offset = computed(() => props.capacity - props.values.length)

// Compute 2D coordinate points
const pts = computed<[number, number][]>(() =>
  props.values.map((v, i) => [x(i + offset.value), y(v)]),
)

// Generate smooth cubic Bézier spline
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

const path = computed(() => smoothSpline(pts.value))

const areaPath = computed(() => {
  if (pts.value.length === 0) return ''
  const startX = pts.value[0]![0].toFixed(1)
  const endX = pts.value[pts.value.length - 1]![0].toFixed(1)
  return `${path.value}L${endX},${H}L${startX},${H}Z`
})

function onMove(e: PointerEvent) {
  const box = (e.currentTarget as SVGElement).getBoundingClientRect()
  const i = Math.round(((e.clientX - box.left) / box.width) * (props.capacity - 1)) - offset.value
  hover.value = i >= 0 && i < props.values.length ? i : null
}

const latest = computed(() => props.values.at(-1))
</script>

<template>
  <figure class="line-chart">
    <figcaption class="chart-header">
      <div class="d-flex align-items-center gap-2">
        <span class="chart-dot" :style="{ backgroundColor: color }"></span>
        <span class="subheader mb-0">{{ title }}</span>
      </div>
      <strong class="h2 mb-0 font-monospace">{{ latest === undefined ? 'Menunggu...' : format(latest) }}</strong>
    </figcaption>

    <div class="plot">
      <svg
        :viewBox="`0 0 ${W} ${H}`"
        preserveAspectRatio="none"
        role="img"
        :aria-label="`${title}, 60 detik terakhir, nilai terbaru ${latest === undefined ? 'belum ada' : format(latest)}`"
        @pointermove="onMove"
        @pointerleave="hover = null"
      >
        <defs>
          <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="color" stop-opacity="0.32" />
            <stop offset="100%" :stop-color="color" stop-opacity="0.0" />
          </linearGradient>
          <filter :id="glowId" x="-10%" y="-10%" width="120%" height="120%">
            <feDropShadow dx="0" dy="2" stdDeviation="3" :flood-color="color" flood-opacity="0.35" />
          </filter>
        </defs>

        <line :x1="0" :x2="W" :y1="H - 0.5" :y2="H - 0.5" class="axis" />
        <line :x1="0" :x2="W" :y1="y(top / 1.15)" :y2="y(top / 1.15)" class="grid" />

        <path :d="areaPath" :fill="`url(#${gradId})`" />
        <path :d="path" class="line" :stroke="color" :filter="`url(#${glowId})`" vector-effect="non-scaling-stroke" />

        <g v-if="hover !== null && values[hover] !== undefined">
          <line
            :x1="x(hover + offset)"
            :x2="x(hover + offset)"
            :y1="0"
            :y2="H"
            class="cross"
            vector-effect="non-scaling-stroke"
          />
          <circle
            :cx="x(hover + offset)"
            :cy="y(values[hover]!)"
            r="4.5"
            :fill="color"
            stroke="var(--tblr-bg-surface)"
            stroke-width="2.5"
          />
        </g>
      </svg>

      <span class="max">Puncak: {{ format(top / 1.15) }}</span>
      <div
        v-if="hover !== null && values[hover] !== undefined"
        class="tip"
        :style="{ left: `${(x(hover + offset) / W) * 100}%` }"
      >
        <span class="tip-val">{{ format(values[hover]!) }}</span>
        <span class="tip-time">{{ values.length - 1 - hover }}s lalu</span>
      </div>
    </div>

    <div class="ticks">
      <span>60 detik lalu</span>
      <span>sekarang</span>
    </div>
  </figure>
</template>

<style scoped>
.line-chart {
  margin: 0;
}
.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
}
.chart-dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  display: inline-block;
  box-shadow: 0 0 6px rgba(var(--tblr-primary-rgb), 0.4);
}
.plot {
  position: relative;
  border-radius: var(--tblr-border-radius);
  padding: 4px 0;
}
svg {
  display: block;
  width: 100%;
  height: 130px;
  overflow: visible;
  touch-action: none;
}
.line {
  fill: none;
  stroke-width: 2.25;
  stroke-linejoin: round;
  stroke-linecap: round;
}
.axis {
  stroke: var(--tblr-border-color);
  stroke-width: 1;
}
.grid {
  stroke: var(--tblr-border-color);
  stroke-width: 1;
  stroke-dasharray: 4 4;
}
.cross {
  stroke: var(--tblr-secondary);
  stroke-width: 1;
  stroke-dasharray: 2 2;
}
.max {
  position: absolute;
  top: -6px;
  right: 0;
  font-size: 11px;
  color: var(--tblr-secondary);
  font-variant-numeric: tabular-nums;
}
.tip {
  position: absolute;
  top: 4px;
  transform: translateX(-50%);
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 3px 8px;
  background: var(--tblr-bg-surface-secondary);
  color: var(--tblr-body-color);
  border: 1px solid var(--tblr-border-color);
  border-radius: 6px;
  box-shadow: 0 6px 16px rgba(0, 0, 0, 0.2);
  backdrop-filter: blur(4px);
  font-size: 11px;
  white-space: nowrap;
  pointer-events: none;
  z-index: 10;
}
.tip-val {
  font-weight: 600;
  font-variant-numeric: tabular-nums;
}
.tip-time {
  color: var(--tblr-secondary);
}
.ticks {
  display: flex;
  justify-content: space-between;
  margin-top: 6px;
  font-size: 11px;
  color: var(--tblr-secondary);
}
</style>
