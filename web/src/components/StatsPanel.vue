<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { wsURL, type Stats } from '@/api'
import { bytes } from '@/format'
import LineChart from '@/components/LineChart.vue'

const props = defineProps<{ id: string }>()
const CAPACITY = 60

const cpu = ref<number[]>([])
const mem = ref<number[]>([])
const last = ref<Stats>()
const status = ref('Menghubungkan...')
let ws: WebSocket | undefined
let prevNet: { rx: number; tx: number } | undefined
const rate = ref({ rx: 0, tx: 0 })

function push(list: number[], v: number) {
  list.push(v)
  if (list.length > CAPACITY) list.shift()
}

onMounted(() => {
  ws = new WebSocket(wsURL(`/containers/${props.id}/stats`))
  ws.onmessage = (e) => {
    const s = JSON.parse(e.data) as Stats
    push(cpu.value, s.cpu_percent)
    push(mem.value, s.mem_usage)
    if (prevNet) rate.value = { rx: Math.max(0, s.net_rx - prevNet.rx), tx: Math.max(0, s.net_tx - prevNet.tx) }
    prevNet = { rx: s.net_rx, tx: s.net_tx }
    last.value = s
    status.value = ''
  }
  ws.onclose = (e) => (status.value = e.reason ? `Stream berhenti: ${e.reason}` : 'Stream berhenti. Container mungkin sudah tidak berjalan.')
})
onUnmounted(() => ws?.close())

const pct = (v: number) => `${v.toFixed(1)}%`
</script>

<template>
  <div>
    <p v-if="status" class="status" role="status">{{ status }}</p>
    <div class="charts">
      <LineChart title="CPU" :values="cpu" :capacity="CAPACITY" :format="pct" :floor="5" />
      <LineChart title="Memori" :values="mem" :capacity="CAPACITY" :format="bytes" :floor="16 * 1024 * 1024" />
    </div>
    <dl v-if="last" class="facts">
      <div><dt>Batas memori</dt><dd>{{ bytes(last.mem_limit) }}</dd></div>
      <div><dt>Memori terpakai</dt><dd>{{ ((last.mem_usage / last.mem_limit) * 100).toFixed(1) }}% dari batas</dd></div>
      <div><dt>Jaringan masuk</dt><dd>{{ bytes(rate.rx) }}/dtk · total {{ bytes(last.net_rx) }}</dd></div>
      <div><dt>Jaringan keluar</dt><dd>{{ bytes(rate.tx) }}/dtk · total {{ bytes(last.net_tx) }}</dd></div>
    </dl>
  </div>
</template>

<style scoped>
.status { color: var(--muted); margin: 0 0 16px; }
.charts { display: grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap: 40px 48px; padding-top: 16px; }
.facts { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 16px 32px; margin: 32px 0 0; padding-top: 24px; border-top: 1px solid var(--hairline); }
dt { color: var(--muted); }
dd { margin: 2px 0 0; color: var(--ink); font-variant-numeric: tabular-nums; }
</style>
