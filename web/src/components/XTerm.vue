<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { Terminal } from '@xterm/xterm'
import { FitAddon } from '@xterm/addon-fit'
import '@xterm/xterm/css/xterm.css'
import { wsURL } from '@/api'

// interactive: stdin goes out as binary frames and size changes as JSON text frames (see ws_controller.go).
const props = defineProps<{ path: string; interactive?: boolean; label: string }>()
const emit = defineEmits<{ closed: [reason: string] }>()

const el = ref<HTMLDivElement>()
let term: Terminal | undefined
let ws: WebSocket | undefined
let observer: ResizeObserver | undefined

function sendSize() {
  if (props.interactive && term && ws?.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ type: 'resize', cols: term.cols, rows: term.rows }))
  }
}

onMounted(() => {
  term = new Terminal({
    convertEol: !props.interactive,
    disableStdin: !props.interactive,
    cursorBlink: props.interactive,
    scrollback: 5000,
    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Consolas, monospace',
    fontSize: 13,
    theme: { background: '#181d26', foreground: '#e6e8eb', cursor: '#e6e8eb', selectionBackground: '#41454d' },
  })
  const fit = new FitAddon()
  term.loadAddon(fit)
  term.open(el.value!)
  fit.fit()
  term.textarea?.setAttribute('aria-label', props.label)

  ws = new WebSocket(wsURL(props.path))
  ws.binaryType = 'arraybuffer'
  ws.onopen = () => {
    sendSize()
    if (props.interactive) term?.focus()
  }
  ws.onmessage = (e) => term?.write(new Uint8Array(e.data as ArrayBuffer))
  // Normal closure means the stream or shell ended; the parent picks the wording.
  ws.onclose = (e) => emit('closed', e.code === 1000 ? '' : e.reason || `koneksi terputus (${e.code})`)

  if (props.interactive) {
    const enc = new TextEncoder()
    term.onData((d) => ws?.readyState === WebSocket.OPEN && ws.send(enc.encode(d)))
    term.onResize(sendSize)
  }
  observer = new ResizeObserver(() => fit.fit())
  observer.observe(el.value!)
})

onUnmounted(() => {
  observer?.disconnect()
  ws?.close()
  term?.dispose()
})

defineExpose({ clear: () => term?.clear() })
</script>

<template>
  <div ref="el" class="xterm-host"></div>
</template>

<style scoped>
.xterm-host {
  height: min(60vh, 560px);
  min-height: 280px;
  padding: 12px;
  background: #181d26;
  border-radius: var(--tblr-border-radius);
  overflow: hidden;
}
</style>
