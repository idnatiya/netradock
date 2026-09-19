<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import { IconCopy, IconPlayerPlay, IconPlayerStop, IconRefresh, IconTerminal2, IconTrash } from '@tabler/icons-vue'
import { api, notify } from '@/api'
import { containerAction } from '@/actions'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
import StateBadge from '@/components/StateBadge.vue'
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
  <PageHeader pretitle="Container" :title="name || id.slice(0, 12)">
    <template #meta>
      <div v-if="data" class="d-flex flex-wrap align-items-center gap-2 mt-1 text-secondary">
        <StateBadge :state="state" :label="state === 'exited' ? `exited, kode ${data.State.ExitCode}` : state" />
        <span v-if="data.State.Health" class="badge bg-secondary-lt">health: {{ data.State.Health.Status }}</span>
        <span v-if="data.RestartCount" class="badge bg-yellow-lt">restart {{ data.RestartCount }}x</span>
        <span class="font-monospace text-break-all">{{ data.Config.Image }}</span>
      </div>
    </template>
    <template v-if="data" #actions>
      <div class="btn-list">
        <button v-if="running" class="btn" type="button" :disabled="busy" @click="act('stop')"><IconPlayerStop :size="18" class="icon" />Hentikan</button>
        <button v-else class="btn btn-primary" type="button" :disabled="busy" @click="act('start')"><IconPlayerPlay :size="18" class="icon" />Jalankan</button>
        <button class="btn" type="button" :disabled="busy" @click="act('restart')"><IconRefresh :size="18" class="icon" />Restart</button>
        <button class="btn btn-ghost-danger" type="button" :disabled="busy" @click="act('remove')"><IconTrash :size="18" class="icon" />Hapus</button>
      </div>
    </template>
  </PageHeader>

  <div class="page-body">
    <div class="container-xl">
      <div v-if="loading || !data" class="card">
        <LoadState :loading="loading" :error="error" :empty="!data" what="container" @retry="reload">
          <template #empty>
            <p class="empty-title">Container tidak ditemukan</p>
            <div class="empty-action"><RouterLink to="/containers" class="btn">Kembali ke daftar container</RouterLink></div>
          </template>
        </LoadState>
      </div>

      <div v-else class="card">
        <div v-if="error" class="alert alert-warning m-3" role="alert">Data mungkin sudah lama: {{ error }}</div>
        <div class="card-header">
          <ul class="nav nav-tabs card-header-tabs flex-nowrap overflow-auto" role="tablist" aria-label="Detail container" @keydown="onTabKey">
            <li v-for="t in tabs" :key="t.key" class="nav-item" role="presentation">
              <button
                :id="`tab-${t.key}`"
                class="nav-link"
                :class="{ active: tab === t.key }"
                role="tab"
                type="button"
                :aria-selected="tab === t.key"
                :aria-controls="`panel-${t.key}`"
                :tabindex="tab === t.key ? 0 : -1"
                @click="select(t.key)"
              >{{ t.label }}</button>
            </li>
          </ul>
        </div>

        <div v-if="tab === 'logs'" id="panel-logs" class="card-body" role="tabpanel" aria-labelledby="tab-logs">
          <div class="d-flex flex-wrap align-items-center gap-2 mb-3">
            <label class="d-flex align-items-center gap-2 mb-0" for="tail">Baris awal</label>
            <select id="tail" v-model="tail" class="form-select w-auto" @change="logKey++">
              <option value="100">100</option>
              <option value="200">200</option>
              <option value="1000">1000</option>
              <option value="5000">5000</option>
            </select>
            <button class="btn" type="button" @click="logEnded = ''; logKey++"><IconRefresh :size="18" class="icon" />Muat ulang log</button>
          </div>
          <XTerm :key="`${id}-${logKey}`" :path="`/containers/${id}/logs?tail=${tail}`" label="Log container" @closed="(r) => (logEnded = r || 'Stream log selesai.')" />
          <div v-if="logEnded" class="text-secondary mt-2" role="status">{{ logEnded }}</div>
        </div>

        <div v-else-if="tab === 'stats'" id="panel-stats" class="card-body" role="tabpanel" aria-labelledby="tab-stats">
          <StatsPanel v-if="running" :key="id" :id="id" />
          <div v-else class="empty">
            <p class="empty-title">Container tidak berjalan</p>
            <p class="empty-subtitle text-secondary">Statistik hanya tersedia saat container berjalan.</p>
            <div class="empty-action"><button class="btn btn-primary" type="button" :disabled="busy" @click="act('start')"><IconPlayerPlay :size="18" class="icon" />Jalankan</button></div>
          </div>
        </div>

        <div v-else-if="tab === 'terminal'" id="panel-terminal" class="card-body" role="tabpanel" aria-labelledby="tab-terminal">
          <div v-if="!running" class="empty">
            <p class="empty-title">Container tidak berjalan</p>
            <p class="empty-subtitle text-secondary">Jalankan container dulu untuk membuka shell.</p>
            <div class="empty-action"><button class="btn btn-primary" type="button" :disabled="busy" @click="act('start')"><IconPlayerPlay :size="18" class="icon" />Jalankan</button></div>
          </div>
          <template v-else>
            <div class="d-flex flex-wrap align-items-center gap-2 mb-3">
              <label class="mb-0" for="shell">Shell</label>
              <select id="shell" v-model="shell" class="form-select w-auto" :disabled="execOn">
                <option>/bin/sh</option>
                <option>/bin/bash</option>
                <option>/bin/ash</option>
              </select>
              <button v-if="!execOn" class="btn btn-primary" type="button" @click="connect"><IconTerminal2 :size="18" class="icon" />Buka shell</button>
              <button v-else class="btn" type="button" @click="execOn = false">Putuskan</button>
            </div>
            <XTerm
              v-if="execOn"
              :key="`${id}-${shell}`"
              :path="`/containers/${id}/exec?cmd=${encodeURIComponent(shell)}`"
              interactive
              label="Terminal container"
              @closed="(r) => { execOn = false; execEnded = r || 'Shell ditutup.' }"
            />
            <div v-if="execEnded" class="text-secondary" role="status">{{ execEnded }}</div>
            <div v-if="!execOn && !execEnded" class="alert alert-warning mb-0">
              Perintah berjalan di dalam container dengan user default image, tanpa batasan perintah.
            </div>
          </template>
        </div>

        <div v-else id="panel-inspect" class="card-body" role="tabpanel" aria-labelledby="tab-inspect">
          <div class="mb-3"><button class="btn" type="button" @click="copyInspect"><IconCopy :size="18" class="icon" />Salin JSON</button></div>
          <pre class="json-view" tabindex="0">{{ JSON.stringify(data, null, 2) }}</pre>
        </div>
      </div>
    </div>
  </div>
</template>
