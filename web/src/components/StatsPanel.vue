<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { wsURL, type Stats } from '@/api'
import { bytes } from '@/format'
import { Card, CardContent } from '@/components/ui/card'
import LineChart from '@/components/LineChart.vue'

const props = defineProps<{ id: string }>()
const CAPACITY = 60

const cpu = ref<number[]>([])
const mem = ref<number[]>([])
const last = ref<Stats>()
const status = ref('Connecting...')
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
  ws.onclose = (e) => (status.value = e.reason ? `Stream closed: ${e.reason}` : 'Stream stopped. Container may have terminated.')
})
onUnmounted(() => ws?.close())

const pct = (v: number) => `${v.toFixed(1)}%`
</script>

<template>
  <div class="space-y-4">
    <div v-if="status" class="text-xs font-mono text-muted-foreground" role="status">
      {{ status }}
    </div>

    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <Card class="p-4 bg-card/50">
        <LineChart title="CPU" :values="cpu" :capacity="CAPACITY" :format="pct" :floor="5" color="#10b981" />
      </Card>
      <Card class="p-4 bg-card/50">
        <LineChart title="Memory" :values="mem" :capacity="CAPACITY" :format="bytes" :floor="16 * 1024 * 1024" color="#0ea5e9" />
      </Card>
    </div>

    <div v-if="last" class="grid grid-cols-2 sm:grid-cols-4 gap-4 pt-2">
      <Card class="p-3">
        <div class="text-[11px] font-mono text-muted-foreground uppercase">Memory Limit</div>
        <div class="text-base font-bold font-mono text-foreground mt-0.5">{{ bytes(last.mem_limit) }}</div>
      </Card>

      <Card class="p-3">
        <div class="text-[11px] font-mono text-muted-foreground uppercase">Memory Used</div>
        <div class="text-base font-bold font-mono text-foreground mt-0.5">
          {{ ((last.mem_usage / last.mem_limit) * 100).toFixed(1) }}%
        </div>
      </Card>

      <Card class="p-3">
        <div class="text-[11px] font-mono text-muted-foreground uppercase">Network RX</div>
        <div class="text-base font-bold font-mono text-foreground mt-0.5">{{ bytes(rate.rx) }}/s</div>
        <div class="text-[10px] text-muted-foreground font-mono mt-0.5">total {{ bytes(last.net_rx) }}</div>
      </Card>

      <Card class="p-3">
        <div class="text-[11px] font-mono text-muted-foreground uppercase">Network TX</div>
        <div class="text-base font-bold font-mono text-foreground mt-0.5">{{ bytes(rate.tx) }}/s</div>
        <div class="text-[10px] text-muted-foreground font-mono mt-0.5">total {{ bytes(last.net_tx) }}</div>
      </Card>
    </div>
  </div>
</template>
