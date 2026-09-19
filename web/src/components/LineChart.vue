<script setup lang="ts">
import { computed, ref } from 'vue'

// One series over the last N one-second samples. Title names the series, so no legend.
const props = defineProps<{ title: string; values: number[]; capacity: number; format: (v: number) => string; floor: number }>()

const W = 600
const H = 140
const hover = ref<number | null>(null)

const top = computed(() => Math.max(props.floor, ...props.values) * 1.15)
const x = (i: number) => (i / (props.capacity - 1)) * W
const y = (v: number) => H - (v / top.value) * H
// Right-align so the newest sample always sits at the right edge.
const offset = computed(() => props.capacity - props.values.length)
const path = computed(() => props.values.map((v, i) => `${i ? 'L' : 'M'}${x(i + offset.value).toFixed(1)},${y(v).toFixed(1)}`).join(''))

function onMove(e: PointerEvent) {
  const box = (e.currentTarget as SVGElement).getBoundingClientRect()
  const i = Math.round(((e.clientX - box.left) / box.width) * (props.capacity - 1)) - offset.value
  hover.value = i >= 0 && i < props.values.length ? i : null
}

const latest = computed(() => props.values.at(-1))
</script>

<template>
  <figure>
    <figcaption>
      <span class="subheader">{{ title }}</span>
      <strong class="h1 mb-0">{{ latest === undefined ? 'menunggu sampel' : format(latest) }}</strong>
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
        <line :x1="0" :x2="W" :y1="H - 0.5" :y2="H - 0.5" class="axis" />
        <line :x1="0" :x2="W" :y1="y(top / 1.15)" :y2="y(top / 1.15)" class="grid" />
        <path :d="path" class="line" vector-effect="non-scaling-stroke" />
        <line v-if="hover !== null" :x1="x(hover + offset)" :x2="x(hover + offset)" :y1="0" :y2="H" class="cross" vector-effect="non-scaling-stroke" />
      </svg>
      <span class="max">garis putus: {{ format(top / 1.15) }}</span>
      <span
        v-if="hover !== null"
        class="tip"
        :style="{ left: `${(x(hover + offset) / W) * 100}%` }"
      >{{ format(values[hover]!) }}, {{ values.length - 1 - hover }} dtk lalu</span>
    </div>
    <div class="ticks"><span>60 dtk lalu</span><span>sekarang</span></div>
  </figure>
</template>

<style scoped>
figure { margin: 0; }
figcaption { display: flex; justify-content: space-between; align-items: baseline; gap: 12px; margin-bottom: 24px; }
figcaption strong { font-weight: 600; font-variant-numeric: tabular-nums; }
.plot { position: relative; }
svg { display: block; width: 100%; height: 140px; overflow: visible; touch-action: none; }
.line { fill: none; stroke: var(--tblr-primary); stroke-width: 2; stroke-linejoin: round; }
.axis { stroke: var(--tblr-border-color-dark, var(--tblr-border-color)); stroke-width: 1; }
.grid { stroke: var(--tblr-border-color); stroke-width: 1; stroke-dasharray: 4 4; }
.cross { stroke: var(--tblr-secondary); stroke-width: 1; }
.max { position: absolute; top: 0; right: 0; transform: translateY(-110%); font-size: 12px; color: var(--tblr-secondary); }
.tip {
  position: absolute;
  top: 8px;
  transform: translateX(-50%);
  padding: 4px 8px;
  background: var(--tblr-body-color);
  color: var(--tblr-bg-surface);
  border-radius: var(--tblr-border-radius);
  font-size: 12px;
  white-space: nowrap;
  pointer-events: none;
}
.ticks { display: flex; justify-content: space-between; margin-top: 6px; font-size: 12px; color: var(--tblr-secondary); }
</style>
