import { computed, onMounted, onUnmounted, reactive, ref } from 'vue'
import { wsURL, type Stats } from '@/api'

const HISTORY = 60

// One shared /ws/stats connection, opened while at least one component uses it.
const latest = ref<Record<string, Stats>>({})
const history = reactive<Record<string, { cpu: number[]; mem: number[] }>>({})
const total = reactive({ cpu: [] as number[], mem: [] as number[] })
const connected = ref(false)

let ws: WebSocket | undefined
let users = 0
let retry: number | undefined

function push(list: number[], v: number) {
  list.push(v)
  if (list.length > HISTORY) list.shift()
}

function open() {
  ws = new WebSocket(wsURL('/stats'))
  ws.onopen = () => (connected.value = true)
  ws.onmessage = (e) => {
    const snap = JSON.parse(e.data) as Record<string, Stats>
    latest.value = snap
    let cpu = 0
    let mem = 0
    for (const [id, s] of Object.entries(snap)) {
      const h = (history[id] ??= { cpu: [], mem: [] })
      push(h.cpu, s.cpu_percent)
      push(h.mem, s.mem_usage)
      cpu += s.cpu_percent
      mem += s.mem_usage
    }
    for (const id of Object.keys(history)) if (!snap[id]) delete history[id]
    push(total.cpu, cpu)
    push(total.mem, mem)
  }
  ws.onclose = () => {
    connected.value = false
    ws = undefined
    if (users > 0) retry = window.setTimeout(open, 3000)
  }
}

export function useLiveStats() {
  onMounted(() => {
    if (users++ === 0 && !ws) open()
  })
  onUnmounted(() => {
    if (--users === 0) {
      clearTimeout(retry)
      ws?.close()
    }
  })
  return { latest, history, total, connected, busiest: computed(() => Object.entries(latest.value).sort(([, a], [, b]) => b.cpu_percent - a.cpu_percent)) }
}
