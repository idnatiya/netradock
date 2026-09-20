<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import { IconBox, IconCheck, IconChevronDown, IconChevronRight, IconCopy, IconDotsVertical, IconFileText, IconPlayerPlay, IconPlayerStop, IconRefresh, IconSearch, IconTerminal2, IconTrash } from '@tabler/icons-vue'
import { api, notify, type Container, type Stats, type System } from '@/api'
import { containerAction } from '@/actions'
import { confirm } from '@/confirm'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
import PortLinks from '@/components/PortLinks.vue'
import StateBadge from '@/components/StateBadge.vue'
import Dropdown from '@/components/Dropdown.vue'
import { bytes, shortId } from '@/format'
import { useLiveStats } from '@/liveStats'
import { useLoad } from '@/useLoad'

const { data, error, loading, reload } = useLoad(() => api<Container[]>('GET', '/containers?all=true'), 5000)
const sys = useLoad(() => api<System>('GET', '/system'), 30000)
const query = ref('')
const onlyRunning = ref(false)
const busy = ref<string | null>(null)
const bulkBusy = ref(false)
const copied = ref<string | null>(null)
const collapsed = ref(new Set<string>())
const selected = ref(new Set<string>())
const live = useLiveStats()

// Compose projects become collapsible groups, like Docker Desktop; standalone containers go last.
const groups = computed(() => {
  const q = query.value.trim().toLowerCase()
  const items = (data.value ?? [])
    .filter((c) => !onlyRunning.value || c.state === 'running')
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
const allChecked = computed(() => shown.value.length > 0 && shown.value.every((c) => selected.value.has(c.id)))
const someChecked = computed(() => !allChecked.value && shown.value.some((c) => selected.value.has(c.id)))

const stats = (id: string): Stats | undefined => live.latest.value[id]
const memPct = (s: Stats) => (s.mem_limit ? (s.mem_usage / s.mem_limit) * 100 : 0)

// Docker Desktop's group row sums its children: usage and limit add up, percentages add up too.
function groupStats(items: Container[]) {
  let cpu = 0, usage = 0, limit = 0, pct = 0, n = 0
  for (const c of items) {
    const s = stats(c.id)
    if (!s) continue
    n++
    cpu += s.cpu_percent
    usage += s.mem_usage
    limit += s.mem_limit
    pct += memPct(s)
  }
  return n ? { cpu, usage, limit, pct } : null
}

// Host-wide readouts in the page header, like Docker Desktop's two usage lines.
const hostCpu = computed(() => live.total.cpu.at(-1) ?? 0)
const hostCpuMax = computed(() => (sys.data.value?.ncpu ?? 0) * 100)
const hostMem = computed(() => live.total.mem.at(-1) ?? 0)

async function copyId(c: Container) {
  try {
    await navigator.clipboard.writeText(c.id)
    copied.value = c.id
    setTimeout(() => (copied.value === c.id) && (copied.value = null), 1200)
  } catch {
    notify('error', 'Browser menolak akses clipboard.')
  }
}

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
      <div class="usage d-flex flex-wrap align-items-start gap-4 mt-2">
        <div>
          <div class="text-secondary small">Pemakaian CPU container</div>
          <div class="text-green">
            {{ hostCpu.toFixed(2) }}% / {{ hostCpuMax }}%
            <span v-if="sys.data.value" class="text-secondary small">({{ sys.data.value.ncpu }} CPU tersedia)</span>
          </div>
        </div>
        <div>
          <div class="text-secondary small">Pemakaian memori container</div>
          <div class="text-green">{{ bytes(hostMem) }}<span v-if="sys.data.value"> / {{ bytes(sys.data.value.mem_total) }}</span></div>
        </div>
        <RouterLink to="/" class="ms-md-auto align-self-center">Lihat grafik</RouterLink>
      </div>
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
            <input v-model="onlyRunning" class="form-check-input" type="checkbox">
            <span class="form-check-label">Hanya tampilkan container yang berjalan</span>
          </label>
        </template>
      </div>

      <div v-if="loading || error || shown.length === 0" class="card">
        <LoadState :loading="loading" :error="error" :empty="shown.length === 0" what="container" @retry="reload">
          <template #empty>
            <div class="empty-icon"><IconBox :size="40" /></div>
            <template v-if="!data?.length">
              <p class="empty-title">Belum ada container</p>
              <p class="empty-subtitle text-secondary">Jalankan container lewat <code>docker run</code> atau <code>docker compose up</code> di server, nanti muncul di sini.</p>
            </template>
            <template v-else>
              <p class="empty-title">Tidak ada yang cocok</p>
              <p class="empty-subtitle text-secondary">Ubah kata kunci atau matikan "Hanya tampilkan container yang berjalan".</p>
              <div class="empty-action">
                <button class="btn" type="button" @click="query = ''; onlyRunning = false">Hapus filter</button>
              </div>
            </template>
          </template>
        </LoadState>
      </div>

      <template v-else>
        <div class="table-responsive-md">
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
                <th class="dot"><span class="visually-hidden">Status</span></th>
                <th>Name</th>
                <th>Container ID</th>
                <th>Image</th>
                <th>Port(s)</th>
                <th>CPU (%)</th>
                <th>Memory usage</th>
                <th>Memory (%)</th>
                <th>Network I/O</th>
                <th class="actions-col">Actions</th>
              </tr>
            </thead>
            <tbody v-for="[project, items] in groups" :key="project">
              <tr v-if="project">
                <td class="pick">
                  <input
                    class="form-check-input m-0"
                    type="checkbox"
                    :aria-label="`Pilih semua di ${project}`"
                    :checked="items.every((c) => selected.has(c.id))"
                    @change="pickGroup(items, ($event.target as HTMLInputElement).checked)"
                  >
                </td>
                <td class="dot"></td>
                <td>
                  <button type="button" class="group-toggle" :aria-expanded="!collapsed.has(project)" @click="toggleGroup(project)">
                    <IconChevronDown v-if="!collapsed.has(project)" :size="16" />
                    <IconChevronRight v-else :size="16" />
                    <span class="fw-medium text-truncate" :title="project">{{ project }}</span>
                  </button>
                </td>
                <td class="text-secondary">-</td>
                <td class="text-secondary">-</td>
                <td class="text-secondary">-</td>
                <td class="tnum">{{ groupStats(items) ? `${groupStats(items)!.cpu.toFixed(2)}%` : '-' }}</td>
                <td class="tnum">{{ groupStats(items) ? `${bytes(groupStats(items)!.usage)} / ${bytes(groupStats(items)!.limit)}` : '-' }}</td>
                <td class="tnum">{{ groupStats(items) ? `${groupStats(items)!.pct.toFixed(2)}%` : '-' }}</td>
                <td class="text-secondary">-</td>
                <td class="actions-col"></td>
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
                <td class="dot" data-label="Status"><StateBadge dot :state="c.state" :label="c.status" /></td>
                <td :class="{ child: !!project }">
                  <RouterLink :to="`/containers/${c.id}`" class="d-block text-truncate" :title="c.name">{{ c.name }}</RouterLink>
                </td>
                <td data-label="Container ID">
                  <span class="d-inline-flex align-items-center gap-1 min-w-0">
                    <span class="font-monospace text-secondary text-truncate">{{ shortId(c.id) }}</span>
                    <button type="button" class="btn btn-icon btn-sm copy" :aria-label="`Salin ID ${c.name}`" title="Salin ID" @click="copyId(c)">
                      <IconCheck v-if="copied === c.id" :size="14" />
                      <IconCopy v-else :size="14" />
                    </button>
                  </span>
                </td>
                <td data-label="Image" class="font-monospace text-truncate" :title="c.image">{{ c.image }}</td>
                <td data-label="Port(s)"><PortLinks :ports="c.ports" /></td>
                <td data-label="CPU (%)" class="tnum">{{ stats(c.id) ? `${stats(c.id)!.cpu_percent.toFixed(2)}%` : '-' }}</td>
                <td data-label="Memory usage" class="tnum">{{ stats(c.id) ? `${bytes(stats(c.id)!.mem_usage)} / ${bytes(stats(c.id)!.mem_limit)}` : '-' }}</td>
                <td data-label="Memory (%)" class="tnum">{{ stats(c.id) ? `${memPct(stats(c.id)!).toFixed(2)}%` : '-' }}</td>
                <td data-label="Network I/O" class="tnum">{{ stats(c.id) ? `${bytes(stats(c.id)!.net_rx)} / ${bytes(stats(c.id)!.net_tx)}` : '-' }}</td>
                <td class="actions-col">
                  <div class="d-inline-flex gap-1">
                    <button v-if="c.state === 'running'" class="btn btn-icon btn-sm" type="button" :disabled="busy === c.id" :aria-label="`Hentikan ${c.name}`" title="Hentikan" @click="act(c, 'stop')"><IconPlayerStop :size="16" /></button>
                    <button v-else class="btn btn-icon btn-sm btn-primary" type="button" :disabled="busy === c.id" :aria-label="`Jalankan ${c.name}`" title="Jalankan" @click="act(c, 'start')"><IconPlayerPlay :size="16" /></button>
                    <Dropdown :label="`Aksi lain untuk ${c.name}`" end class="btn btn-icon btn-sm" :disabled="busy === c.id">
                      <template #toggle><IconDotsVertical :size="16" /></template>
                      <button type="button" class="dropdown-item" role="menuitem" @click="act(c, 'restart')"><IconRefresh :size="18" class="icon dropdown-item-icon" />Restart</button>
                      <RouterLink :to="`/containers/${c.id}?tab=logs`" class="dropdown-item" role="menuitem"><IconFileText :size="18" class="icon dropdown-item-icon" />Lihat log</RouterLink>
                      <RouterLink v-if="c.state === 'running'" :to="`/containers/${c.id}?tab=terminal`" class="dropdown-item" role="menuitem"><IconTerminal2 :size="18" class="icon dropdown-item-icon" />Terminal</RouterLink>
                    </Dropdown>
                    <button class="btn btn-icon btn-sm btn-ghost-danger" type="button" :disabled="busy === c.id" :aria-label="`Hapus ${c.name}`" title="Hapus" @click="act(c, 'remove')"><IconTrash :size="16" /></button>
                  </div>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
        <div class="text-secondary small text-end mt-2">Menampilkan {{ shown.length }} item</div>
      </template>
    </div>
  </div>
</template>

<style scoped>
.search { flex-grow: 1; max-width: 320px; }
@media (max-width: 575.98px) { .search { flex-basis: 100%; max-width: none; } }
.usage { column-gap: 3rem; }
.tnum { font-variant-numeric: tabular-nums; }
.copy { border: 0; background: transparent; color: var(--tblr-secondary); padding: 0; width: 1.25rem; height: 1.25rem; }
.group-toggle {
  display: inline-flex;
  align-items: center;
  gap: 0.375rem;
  max-width: 100%;
  padding: 0;
  border: 0;
  background: transparent;
  color: inherit;
}

@media (min-width: 768px) {
  .containers { table-layout: fixed; }
  .containers .pick { width: 2.25rem; }
  .containers .dot { width: 1.75rem; }
  .containers th:nth-child(3) { width: 14%; }
  .containers th:nth-child(4) { width: 11%; }
  .containers th:nth-child(5) { width: 13%; }
  .containers th:nth-child(6) { width: 10%; }
  .containers th:nth-child(7) { width: 7%; }
  .containers th:nth-child(8) { width: 14%; }
  .containers th:nth-child(9) { width: 8%; }
  .containers th:nth-child(10) { width: 11%; }
  .containers .actions-col { width: 7.5rem; border-left: var(--tblr-border-width) solid var(--tblr-border-color); }
  .containers td.child { padding-left: 1.75rem; }
}
</style>
