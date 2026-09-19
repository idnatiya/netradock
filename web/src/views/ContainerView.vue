<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { api, notify } from '@/api'
import { containerAction } from '@/actions'
import LoadState from '@/components/LoadState.vue'
import StatsPanel from '@/components/StatsPanel.vue'
import XTerm from '@/components/XTerm.vue'
import { useLoad } from '@/useLoad'

interface Inspect {
  Id: string
  Name: string
  Created: string
  State: { Status: string; StartedAt: string; ExitCode: number; Health?: { Status: string } }
  Config: { Image: string; Cmd: string[] | null; Labels: Record<string, string> | null }
  RestartCount: number
}

const route = useRoute()
const router = useRouter()
const id = computed(() => route.params.id as string)
const { data, error, loading, reload } = useLoad(() => api<Inspect>('GET', `/containers/${id.value}`), 5000)

const tabs = [
  { key: 'logs', label: 'Log' },
  { key: 'stats', label: 'Statistik' },
  { key: 'terminal', label: 'Terminal' },
  { key: 'inspect', label: 'Inspect' },
] as const
type Tab = (typeof tabs)[number]['key']
const tab = computed<Tab>(() => (tabs.some((t) => t.key === route.query.tab) ? (route.query.tab as Tab) : 'logs'))

function select(t: Tab) {
  router.replace({ query: { ...route.query, tab: t } })
}
function onTabKey(e: KeyboardEvent) {
  const i = tabs.findIndex((t) => t.key === tab.value)
  const next = e.key === 'ArrowRight' ? i + 1 : e.key === 'ArrowLeft' ? i - 1 : null
  if (next === null) return
  const t = tabs[(next + tabs.length) % tabs.length]!
  select(t.key)
  document.getElementById(`tab-${t.key}`)?.focus()
}

const name = computed(() => data.value?.Name.replace(/^\//, '') ?? '')
const state = computed(() => data.value?.State.Status ?? '')
const running = computed(() => state.value === 'running')

const busy = ref(false)
async function act(action: 'start' | 'stop' | 'restart' | 'remove') {
  busy.value = true
  const ok = await containerAction(name.value, id.value, action, running.value)
  busy.value = false
  if (ok && action === 'remove') router.push('/containers')
  else if (ok) {
    await reload()
    logKey.value++
  }
}

// Logs: remount the terminal to restart the stream with a new tail size.
const tail = ref('200')
const logKey = ref(0)
const logEnded = ref('')

// Exec: connect on demand, never automatically.
const shell = ref('/bin/sh')
const execOn = ref(false)
const execEnded = ref('')
function connect() {
  execEnded.value = ''
  execOn.value = true
}

async function copyInspect() {
  try {
    await navigator.clipboard.writeText(JSON.stringify(data.value, null, 2))
    notify('ok', 'JSON inspect disalin.')
  } catch {
    notify('error', 'Clipboard tidak tersedia. Pilih teks lalu salin manual.')
  }
}
</script>

<template>
  <main class="page">
    <p class="crumb"><RouterLink to="/containers">Container</RouterLink> / {{ name || id.slice(0, 12) }}</p>
    <LoadState :loading="loading" :error="error" :empty="!data" what="container" @retry="reload">
      <template v-if="data">
        <div class="page-head">
          <div>
            <h1>{{ name }}</h1>
            <p>
              <span class="state" :class="state">{{ state }}</span>
              <span v-if="state === 'exited'"> dengan kode {{ data.State.ExitCode }}</span>
              <span v-if="data.State.Health"> · health: {{ data.State.Health.Status }}</span>
              <span v-if="data.RestartCount"> · restart {{ data.RestartCount }}x</span>
            </p>
            <p class="mono image">{{ data.Config.Image }}</p>
          </div>
          <div class="head-actions">
            <button v-if="running" class="btn" type="button" :disabled="busy" @click="act('stop')">Hentikan</button>
            <button v-else class="btn btn-primary" type="button" :disabled="busy" @click="act('start')">Jalankan</button>
            <button class="btn" type="button" :disabled="busy" @click="act('restart')">Restart</button>
            <button class="btn btn-danger" type="button" :disabled="busy" @click="act('remove')">Hapus</button>
          </div>
        </div>

        <div class="tabs" role="tablist" aria-label="Detail container" @keydown="onTabKey">
          <button
            v-for="t in tabs"
            :id="`tab-${t.key}`"
            :key="t.key"
            role="tab"
            type="button"
            :aria-selected="tab === t.key"
            :aria-controls="`panel-${t.key}`"
            :tabindex="tab === t.key ? 0 : -1"
            @click="select(t.key)"
          >{{ t.label }}</button>
        </div>

        <section v-if="tab === 'logs'" id="panel-logs" role="tabpanel" aria-labelledby="tab-logs">
          <div class="toolbar">
            <label>Baris awal
              <select v-model="tail" @change="logKey++">
                <option value="100">100</option>
                <option value="200">200</option>
                <option value="1000">1000</option>
                <option value="5000">5000</option>
              </select>
            </label>
            <button class="btn btn-sm" type="button" @click="logEnded = ''; logKey++">Muat ulang log</button>
          </div>
          <XTerm :key="`${id}-${logKey}`" :path="`/containers/${id}/logs?tail=${tail}`" label="Log container" @closed="(r) => (logEnded = r || 'Stream log selesai.')" />
          <p v-if="logEnded" class="ended" role="status">{{ logEnded }}</p>
        </section>

        <section v-else-if="tab === 'stats'" id="panel-stats" role="tabpanel" aria-labelledby="tab-stats">
          <StatsPanel v-if="running" :key="id" :id="id" />
          <div v-else class="empty">
            <strong>Container tidak berjalan.</strong>
            Statistik hanya tersedia saat container berjalan.
          </div>
        </section>

        <section v-else-if="tab === 'terminal'" id="panel-terminal" role="tabpanel" aria-labelledby="tab-terminal">
          <div v-if="!running" class="empty">
            <strong>Container tidak berjalan.</strong>
            Jalankan container dulu untuk membuka shell.
          </div>
          <template v-else>
            <div class="toolbar">
              <label>Shell
                <select v-model="shell" :disabled="execOn">
                  <option>/bin/sh</option>
                  <option>/bin/bash</option>
                  <option>/bin/ash</option>
                </select>
              </label>
              <button v-if="!execOn" class="btn btn-primary btn-sm" type="button" @click="connect">Buka shell</button>
              <button v-else class="btn btn-sm" type="button" @click="execOn = false">Putuskan</button>
            </div>
            <XTerm
              v-if="execOn"
              :key="`${id}-${shell}`"
              :path="`/containers/${id}/exec?cmd=${encodeURIComponent(shell)}`"
              interactive
              label="Terminal container"
              @closed="(r) => { execOn = false; execEnded = r || 'Shell ditutup.' }"
            />
            <p v-if="execEnded" class="ended" role="status">{{ execEnded }}</p>
            <p v-if="!execOn && !execEnded" class="hint">Perintah berjalan di dalam container dengan user default image. Hati-hati: tidak ada batasan perintah.</p>
          </template>
        </section>

        <section v-else id="panel-inspect" role="tabpanel" aria-labelledby="tab-inspect">
          <div class="toolbar"><button class="btn btn-sm" type="button" @click="copyInspect">Salin JSON</button></div>
          <pre class="json" tabindex="0">{{ JSON.stringify(data, null, 2) }}</pre>
        </section>
      </template>
    </LoadState>
  </main>
</template>

<style scoped>
.crumb { margin: 0 0 12px; color: var(--muted); overflow-wrap: anywhere; }
h1 { overflow-wrap: anywhere; }
.image { margin-top: 4px; overflow-wrap: anywhere; }
.head-actions { display: flex; flex-wrap: wrap; gap: 8px; }
.tabs { display: flex; gap: 4px; border-bottom: 1px solid var(--hairline); margin-bottom: 24px; overflow-x: auto; }
.tabs button {
  min-height: 44px;
  padding: 0 16px;
  border: 0;
  background: none;
  color: var(--body);
  font: 500 14px var(--font);
  cursor: pointer;
  white-space: nowrap;
}
.tabs button[aria-selected='true'] { color: var(--ink); box-shadow: inset 0 -2px 0 var(--ink); }
.toolbar { display: flex; flex-wrap: wrap; align-items: center; gap: 12px; margin-bottom: 12px; }
.toolbar label { display: inline-flex; align-items: center; gap: 8px; color: var(--ink); }
select {
  height: 44px;
  padding: 0 12px;
  border: 1px solid var(--hairline);
  border-radius: var(--r-sm);
  background: var(--canvas);
  color: var(--ink);
  font: inherit;
}
.ended, .hint { color: var(--muted); margin: 12px 0 0; }
.json {
  margin: 0;
  max-height: 70vh;
  overflow: auto;
  padding: 16px;
  background: var(--surface-soft);
  border: 1px solid var(--hairline);
  border-radius: var(--r-md);
  font: 12px/1.5 var(--mono);
  color: var(--ink);
}
</style>
