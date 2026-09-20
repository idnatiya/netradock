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
    color: '#38bdf8',
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
  <figure class="m-0 select-none">
    <figcaption class="flex items-center justify-between gap-3 mb-3">
      <div class="flex items-center gap-2">
        <span class="size-2 rounded-full" :style="{ backgroundColor: color }" />
        <span class="text-xs font-medium text-muted-foreground uppercase tracking-wider">{{ title }}</span>
      </div>
      <strong class="text-lg font-mono font-semibold tracking-tight text-foreground">
        {{ latest === undefined ? 'Waiting...' : format(latest) }}
      </strong>
    </figcaption>

    <div class="relative py-1">
      <svg
        :viewBox="`0 0 ${W} ${H}`"
        preserveAspectRatio="none"
        role="img"
        class="block w-full h-[120px] overflow-visible touch-none"
        :aria-label="`${title}, last 60 seconds, latest ${latest === undefined ? 'none' : format(latest)}`"
        @pointermove="onMove"
        @pointerleave="hover = null"
      >
        <defs>
          <linearGradient :id="gradId" x1="0" y1="0" x2="0" y2="1">
            <stop offset="0%" :stop-color="color" stop-opacity="0.25" />
            <stop offset="100%" :stop-color="color" stop-opacity="0.0" />
          </linearGradient>
          <filter :id="glowId" x="-10%" y="-10%" width="120%" height="120%">
            <feDropShadow dx="0" dy="2" stdDeviation="3" :flood-color="color" flood-opacity="0.3" />
          </filter>
        </defs>

        <!-- Horizontal grid lines -->
        <g stroke="currentColor" stroke-opacity="0.1" stroke-dasharray="3 3">
          <line :x1="0" :y1="H * 0.25" :x2="W" :y2="H * 0.25" />
          <line :x1="0" :y1="H * 0.5" :x2="W" :y2="H * 0.5" />
          <line :x1="0" :y1="H * 0.75" :x2="W" :y2="H * 0.75" />
        </g>

        <!-- Base axis -->
        <line :x1="0" :y1="H" :x2="W" :y2="H" stroke="currentColor" stroke-opacity="0.15" />

        <!-- Gradient fill under curve -->
        <path v-if="areaPath" :d="areaPath" :fill="`url(#${gradId})`" />

        <!-- Smooth Bézier Spline Line -->
        <path
          v-if="path"
          :d="path"
          :stroke="color"
          fill="none"
          stroke-width="2"
          stroke-linecap="round"
          stroke-linejoin="round"
          :filter="`url(#${glowId})`"
        />

        <!-- Hover crosshair & point indicator -->
        <g v-if="hover !== null && values[hover] !== undefined">
          <line
            :x1="x(hover + offset)"
            :y1="0"
            :x2="x(hover + offset)"
            :y2="H"
            stroke="currentColor"
            stroke-opacity="0.3"
            stroke-dasharray="2 2"
          />
          <circle
            :cx="x(hover + offset)"
            :cy="y(values[hover]!)"
            r="4"
            :fill="color"
            stroke="hsl(var(--background))"
            stroke-width="2"
          />
        </g>
      </svg>

      <span class="absolute -top-1.5 right-0 text-[10px] font-mono text-muted-foreground">
        Peak: {{ format(top / 1.15) }}
      </span>

      <!-- Tooltip -->
      <div
        v-if="hover !== null && values[hover] !== undefined"
        class="absolute top-1 -translate-x-1/2 flex items-center gap-1.5 px-2 py-1 rounded-md border border-border bg-popover text-popover-foreground shadow-md text-xs font-mono pointer-events-none z-10"
        :style="{ left: `${(x(hover + offset) / W) * 100}%` }"
      >
        <span class="font-semibold">{{ format(values[hover]!) }}</span>
        <span class="text-muted-foreground text-[10px]">{{ values.length - 1 - hover }}s ago</span>
      </div>
    </div>

    <div class="flex justify-between mt-1 text-[11px] font-mono text-muted-foreground">
      <span>60s ago</span>
      <span>now</span>
    </div>
  </figure>
</template>
