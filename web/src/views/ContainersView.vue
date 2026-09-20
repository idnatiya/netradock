<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { IconBox, IconChevronDown, IconChevronRight, IconDotsVertical, IconFileText, IconPlayerPlay, IconPlayerStop, IconRefresh, IconSearch, IconTerminal2, IconTrash } from '@tabler/icons-vue'
import { api, notify, type Container } from '@/api'
import { containerAction } from '@/actions'
import { confirm } from '@/confirm'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
import PortLinks from '@/components/PortLinks.vue'
import StateBadge from '@/components/StateBadge.vue'
import Dropdown from '@/components/Dropdown.vue'
import Sparkline from '@/components/Sparkline.vue'
import { bytes } from '@/format'
import { useLiveStats } from '@/liveStats'
import { useLoad } from '@/useLoad'

const { data, error, loading, reload } = useLoad(() => api<Container[]>('GET', '/containers?all=true'), 5000)
const query = ref('')
const showStopped = ref(true)
const busy = ref<string | null>(null)
const bulkBusy = ref(false)
const collapsed = ref(new Set<string>())
const selected = ref(new Set<string>())
const live = useLiveStats()
const stateColor = (state: string) => (state === 'running' ? 'bg-green-lt' : state === 'exited' || state === 'dead' ? 'bg-red-lt' : 'bg-yellow-lt')

// Compose projects become collapsible groups, like Docker Desktop; standalone containers go last.
const groups = computed(() => {
  const q = query.value.trim().toLowerCase()
  const items = (data.value ?? [])
    .filter((c) => showStopped.value || c.state === 'running')
    .filter((c) => !q || c.name.toLowerCase().includes(q) || c.image.toLowerCase().includes(q))
    .sort((a, b) => a.name.localeCompare(b.name))
  const map = new Map<string, Container[]>()
  for (const c of items) {
    const key = c.project ?? ''
    map.set(key, [...(map.get(key) ?? []), c])
  }
  return [...map.entries()].sort(([a], [b]) => (a === '' ? 1 : b === '' ? -1 : a.localeCompare(b)))
})

const shown = computed(() => groups.value.flatMap(([, items]) => items))
const total = computed(() => data.value?.length ?? 0)
const running = computed(() => data.value?.filter((c) => c.state === 'running').length ?? 0)
const runningIn = (items: Container[]) => items.filter((c) => c.state === 'running').length

const allChecked = computed(() => shown.value.length > 0 && shown.value.every((c) => selected.value.has(c.id)))
const someChecked = computed(() => !allChecked.value && shown.value.some((c) => selected.value.has(c.id)))

function toggleGroup(project: string) {
  collapsed.value.has(project) ? collapsed.value.delete(project) : collapsed.value.add(project)
}
function pick(id: string, on: boolean) {
  on ? selected.value.add(id) : selected.value.delete(id)
}
function pickAll(on: boolean) {
  for (const c of shown.value) pick(c.id, on)
}
function pickGroup(items: Container[], on: boolean) {
  for (const c of items) pick(c.id, on)
}

// Containers can disappear between polls; never keep a stale id selected.
watch(data, (list) => {
  const ids = new Set((list ?? []).map((c) => c.id))
  for (const id of [...selected.value]) if (!ids.has(id)) selected.value.delete(id)
})

async function act(c: Container, action: 'start' | 'stop' | 'restart' | 'remove') {
  busy.value = c.id
  if (await containerAction(c.name, c.id, action, c.state === 'running')) await reload()
  busy.value = null
}

const bulkLabels = { start: 'Jalankan', stop: 'Hentikan', restart: 'Restart', remove: 'Hapus' } as const
const bulkVerbs = { start: 'dijalankan', stop: 'dihentikan', restart: 'di-restart', remove: 'dihapus' } as const

// One confirmation for the whole batch, so containerAction()'s per-item prompt is bypassed on purpose.
async function bulk(action: keyof typeof bulkLabels) {
  const items = (data.value ?? []).filter((c) => selected.value.has(c.id))
  if (!items.length) return
  const names = items.map((c) => c.name).join(', ')
  const body = action === 'remove'
    ? `${names}. Yang sedang berjalan dihentikan paksa, data di luar volume ikut hilang.`
    : names
  if (!(await confirm(`${bulkLabels[action]} ${items.length} container?`, body, bulkLabels[action]))) return

  bulkBusy.value = true
  const results = await Promise.allSettled(items.map((c) =>
    action === 'remove'
      ? api('DELETE', `/containers/${c.id}?force=${c.state === 'running'}`)
      : api('POST', `/containers/${c.id}/${action}`)))
  const ok = results.filter((r) => r.status === 'fulfilled').length
  const failed = results.length - ok
  notify(failed ? 'error' : 'ok', failed ? `${ok} berhasil, ${failed} gagal.` : `${ok} container ${bulkVerbs[action]}.`)
  selected.value.clear()
  bulkBusy.value = false
  await reload()
}
</script>

<template>
  <PageHeader title="Containers">
    <template #meta>
      <div v-if="data" class="text-secondary mt-1">{{ running }} berjalan, {{ total - running }} tidak berjalan</div>
    </template>
  </PageHeader>

  <div class="page-body">
    <div class="container-xl">
      <div class="toolbar d-flex flex-wrap align-items-center gap-3 mb-2">
        <template v-if="selected.size">
          <span class="fw-medium">{{ selected.size }} dipilih</span>
          <div class="btn-list">
            <button class="btn btn-sm" type="button" :disabled="bulkBusy" @click="bulk('start')"><IconPlayerPlay :size="16" class="icon" />Jalankan</button>
            <button class="btn btn-sm" type="button" :disabled="bulkBusy" @click="bulk('stop')"><IconPlayerStop :size="16" class="icon" />Hentikan</button>
            <button class="btn btn-sm" type="button" :disabled="bulkBusy" @click="bulk('restart')"><IconRefresh :size="16" class="icon" />Restart</button>
            <button class="btn btn-sm btn-ghost-danger" type="button" :disabled="bulkBusy" @click="bulk('remove')"><IconTrash :size="16" class="icon" />Hapus</button>
          </div>
          <button class="btn btn-sm btn-ghost-secondary ms-auto" type="button" @click="selected.clear()">Batal pilih</button>
        </template>
        <template v-else>
          <div class="input-icon search">
            <span class="input-icon-addon"><IconSearch :size="18" /></span>
            <input v-model="query" type="search" class="form-control" placeholder="Cari nama atau image" aria-label="Cari container">
          </div>
          <label class="form-check form-switch mb-0">
            <input v-model="showStopped" class="form-check-input" type="checkbox">
            <span class="form-check-label">Tampilkan yang berhenti</span>
          </label>
        </template>
      </div>

      <div v-if="loading || error || shown.length === 0" class="card">
        <LoadState :loading="loading" :error="error" :empty="shown.length === 0" what="container" @retry="reload">
          <template #empty>
            <div class="empty-icon"><IconBox :size="40" /></div>
            <template v-if="total === 0">
              <p class="empty-title">Belum ada container</p>
              <p class="empty-subtitle text-secondary">Jalankan container lewat <code>docker run</code> atau <code>docker compose up</code> di server, nanti muncul di sini.</p>
            </template>
            <template v-else>
              <p class="empty-title">Tidak ada yang cocok</p>
              <p class="empty-subtitle text-secondary">Ubah kata kunci atau nyalakan "Tampilkan yang berhenti".</p>
              <div class="empty-action">
                <button class="btn" type="button" @click="query = ''; showStopped = true">Hapus filter</button>
              </div>
            </template>
          </template>
        </LoadState>
      </div>

      <div v-else class="table-responsive-md">
        <table class="table table-vcenter table-stack dd-table containers">
          <thead>
            <tr>
              <th class="pick">
                <input
                  class="form-check-input m-0"
                  type="checkbox"
                  aria-label="Pilih semua container"
                  :checked="allChecked"
                  :indeterminate="someChecked"
                  @change="pickAll(($event.target as HTMLInputElement).checked)"
                >
              </th>
              <th>Container</th><th>Status</th><th>Port</th><th>CPU</th><th>Memori</th><th><span class="visually-hidden">Aksi</span></th>
            </tr>
          </thead>
          <tbody v-for="[project, items] in groups" :key="project">
            <tr v-if="project" class="group-row">
              <td class="pick">
                <input
                  class="form-check-input m-0"
                  type="checkbox"
                  :aria-label="`Pilih semua di ${project}`"
                  :checked="items.every((c) => selected.has(c.id))"
                  @change="pickGroup(items, ($event.target as HTMLInputElement).checked)"
                >
              </td>
              <td colspan="6">
                <button type="button" class="group-toggle" :aria-expanded="!collapsed.has(project)" @click="toggleGroup(project)">
                  <IconChevronDown v-if="!collapsed.has(project)" :size="16" />
                  <IconChevronRight v-else :size="16" />
                  <span class="fw-medium text-break-all">{{ project }}</span>
                </button>
                <span class="text-secondary small ms-2">{{ runningIn(items) }}/{{ items.length }} berjalan</span>
              </td>
            </tr>
            <tr v-for="c in (project && collapsed.has(project) ? [] : items)" :key="c.id" :aria-busy="busy === c.id">
              <td class="pick">
                <input
                  class="form-check-input m-0"
                  type="checkbox"
                  :aria-label="`Pilih ${c.name}`"
                  :checked="selected.has(c.id)"
                  @change="pick(c.id, ($event.target as HTMLInputElement).checked)"
                >
              </td>
              <td :class="{ child: !!project }">
                <div class="d-flex align-items-center gap-2 min-w-0">
                  <span class="avatar avatar-sm flex-shrink-0" :class="stateColor(c.state)"><IconBox :size="18" /></span>
                  <div class="min-w-0">
                    <RouterLink :to="`/containers/${c.id}`" class="fw-medium d-block text-truncate">{{ c.name }}</RouterLink>
                    <div class="text-secondary small font-monospace text-truncate" :title="c.image">{{ c.image }}</div>
                  </div>
                </div>
              </td>
              <td data-label="Status"><StateBadge :state="c.state" :label="c.status" /></td>
              <td data-label="Port"><PortLinks :ports="c.ports" /></td>
              <td data-label="CPU">
                <template v-if="live.latest.value[c.id]">
                  <span class="small d-md-block">{{ live.latest.value[c.id]!.cpu_percent.toFixed(1) }}%</span>
                  <Sparkline class="d-none d-md-block" :values="live.history[c.id]?.cpu ?? []" :height="20" :floor="5" />
                </template>
                <span v-else class="text-secondary">-</span>
              </td>
              <td data-label="Memori">
                <template v-if="live.latest.value[c.id]">
                  <span class="small d-md-block">{{ bytes(live.latest.value[c.id]!.mem_usage) }}</span>
                  <div class="progress progress-xs mt-1 d-none d-md-flex">
                    <div class="progress-bar" :style="{ width: `${Math.min(100, (live.latest.value[c.id]!.mem_usage / live.latest.value[c.id]!.mem_limit) * 100)}%` }"></div>
                  </div>
                </template>
                <span v-else class="text-secondary">-</span>
              </td>
              <td class="text-end">
                <div class="d-inline-flex gap-1 row-actions">
                  <button v-if="c.state === 'running'" class="btn btn-icon" type="button" :disabled="busy === c.id" :aria-label="`Hentikan ${c.name}`" title="Hentikan" @click="act(c, 'stop')"><IconPlayerStop :size="18" /></button>
                  <button v-else class="btn btn-icon btn-primary" type="button" :disabled="busy === c.id" :aria-label="`Jalankan ${c.name}`" title="Jalankan" @click="act(c, 'start')"><IconPlayerPlay :size="18" /></button>
                  <Dropdown :label="`Aksi lain untuk ${c.name}`" end class="btn btn-icon" :disabled="busy === c.id">
                    <template #toggle><IconDotsVertical :size="18" /></template>
                    <button type="button" class="dropdown-item" role="menuitem" @click="act(c, 'restart')"><IconRefresh :size="18" class="icon dropdown-item-icon" />Restart</button>
                    <RouterLink :to="`/containers/${c.id}?tab=logs`" class="dropdown-item" role="menuitem"><IconFileText :size="18" class="icon dropdown-item-icon" />Lihat log</RouterLink>
                    <RouterLink v-if="c.state === 'running'" :to="`/containers/${c.id}?tab=terminal`" class="dropdown-item" role="menuitem"><IconTerminal2 :size="18" class="icon dropdown-item-icon" />Terminal</RouterLink>
                    <div class="dropdown-divider"></div>
                    <button type="button" class="dropdown-item text-danger" role="menuitem" @click="act(c, 'remove')"><IconTrash :size="18" class="icon dropdown-item-icon" />Hapus</button>
                  </Dropdown>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.search { flex-grow: 1; max-width: 320px; }
@media (max-width: 575.98px) { .search { flex-basis: 100%; max-width: none; } }
.group-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
}
.group-row td { background: var(--tblr-bg-surface-secondary); }

@media (min-width: 768px) {
  .containers { table-layout: fixed; }
  .containers .pick { width: 2.5rem; }
  .containers th:nth-child(2) { width: 28%; }
  .containers th:nth-child(3) { width: 18%; }
  .containers th:nth-child(4) { width: 13%; }
  .containers th:nth-child(5) { width: 11%; }
  .containers th:nth-child(6) { width: 11%; }
  .containers th:nth-child(7) { width: 11%; }
  .containers td.child { padding-left: 1.75rem; }
}
</style>
