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
    <div v-if="status" class="text-secondary mb-3" role="status">{{ status }}</div>
    <div class="row row-cards">
      <div class="col-md-6">
        <div class="card"><div class="card-body"><LineChart title="CPU" :values="cpu" :capacity="CAPACITY" :format="pct" :floor="5" /></div></div>
      </div>
      <div class="col-md-6">
        <div class="card"><div class="card-body"><LineChart title="Memori" :values="mem" :capacity="CAPACITY" :format="bytes" :floor="16 * 1024 * 1024" /></div></div>
      </div>
    </div>
    <div v-if="last" class="datagrid mt-4">
      <div class="datagrid-item"><div class="datagrid-title">Batas memori</div><div class="datagrid-content">{{ bytes(last.mem_limit) }}</div></div>
      <div class="datagrid-item"><div class="datagrid-title">Memori terpakai</div><div class="datagrid-content">{{ ((last.mem_usage / last.mem_limit) * 100).toFixed(1) }}% dari batas</div></div>
      <div class="datagrid-item"><div class="datagrid-title">Jaringan masuk</div><div class="datagrid-content">{{ bytes(rate.rx) }}/dtk · total {{ bytes(last.net_rx) }}</div></div>
      <div class="datagrid-item"><div class="datagrid-title">Jaringan keluar</div><div class="datagrid-content">{{ bytes(rate.tx) }}/dtk · total {{ bytes(last.net_tx) }}</div></div>
    </div>
  </div>
</template>
