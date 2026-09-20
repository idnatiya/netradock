<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink } from 'vue-router'
import {
  Boxes,
  Check,
  ChevronDown,
  ChevronRight,
  Copy,
  FileText,
  Play,
  RotateCcw,
  Search,
  Square,
  Terminal,
  Trash2,
  X,
} from 'lucide-vue-next'
import { api, notify, type Container, type Stats, type System } from '@/api'
import { containerAction } from '@/actions'
import { confirm } from '@/confirm'
import { bytes, shortId } from '@/format'
import { useLiveStats } from '@/liveStats'
import { useLoad } from '@/useLoad'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
import PortLinks from '@/components/PortLinks.vue'
import StateBadge from '@/components/StateBadge.vue'

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

// Compose projects become collapsible groups, standalone containers go last
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

const hostCpu = computed(() => live.total.cpu.at(-1) ?? 0)
const hostCpuMax = computed(() => (sys.data.value?.ncpu ?? 0) * 100)
const hostMem = computed(() => live.total.mem.at(-1) ?? 0)

async function copyId(c: Container) {
  try {
    await navigator.clipboard.writeText(c.id)
    copied.value = c.id
    setTimeout(() => (copied.value === c.id) && (copied.value = null), 1200)
  } catch {
    notify('error', 'Browser denied clipboard access.')
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

watch(data, (list) => {
  const ids = new Set((list ?? []).map((c) => c.id))
  for (const id of [...selected.value]) if (!ids.has(id)) selected.value.delete(id)
})

async function act(c: Container, action: 'start' | 'stop' | 'restart' | 'remove') {
  busy.value = c.id
  if (await containerAction(c.name, c.id, action, c.state === 'running')) await reload()
  busy.value = null
}

const bulkLabels = { start: 'Start', stop: 'Stop', restart: 'Restart', remove: 'Remove' } as const
const bulkVerbs = { start: 'started', stop: 'stopped', restart: 'restarted', remove: 'removed' } as const

async function bulk(action: keyof typeof bulkLabels) {
  const items = (data.value ?? []).filter((c) => selected.value.has(c.id))
  if (!items.length) return
  const names = items.map((c) => c.name).join(', ')
  const body = action === 'remove'
    ? `${names}. Running containers will be forcefully stopped.`
    : names
  if (!(await confirm(`${bulkLabels[action]} ${items.length} containers?`, body, bulkLabels[action]))) return

  bulkBusy.value = true
  const results = await Promise.allSettled(items.map((c) =>
    action === 'remove'
      ? api('DELETE', `/containers/${c.id}?force=${c.state === 'running'}`)
      : api('POST', `/containers/${c.id}/${action}`)))
  const ok = results.filter((r) => r.status === 'fulfilled').length
  const failed = results.length - ok
  notify(failed ? 'error' : 'ok', failed ? `${ok} succeeded, ${failed} failed.` : `${ok} containers ${bulkVerbs[action]}.`)
  selected.value.clear()
  bulkBusy.value = false
  await reload()
}
</script>

<template>
  <div>
    <!-- Page Header -->
    <PageHeader title="Containers">
      <template #meta>
        <div class="flex flex-wrap items-center gap-4 text-xs font-mono text-muted-foreground mt-1">
          <div>
            <span>CPU: </span>
            <span class="text-emerald-600 dark:text-emerald-400 font-semibold">{{ hostCpu.toFixed(2) }}%</span>
            <span> / {{ hostCpuMax }}%</span>
          </div>
          <span class="text-border">|</span>
          <div>
            <span>RAM: </span>
            <span class="text-emerald-600 dark:text-emerald-400 font-semibold">{{ bytes(hostMem) }}</span>
            <span v-if="sys.data.value"> / {{ bytes(sys.data.value.mem_total) }}</span>
          </div>
        </div>
      </template>
    </PageHeader>

    <div class="p-6 max-w-7xl mx-auto space-y-4">
      <!-- Toolbar -->
      <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
        <!-- Batch selection bar -->
        <div v-if="selected.size" class="flex flex-wrap items-center gap-2">
          <Badge variant="secondary" class="font-mono text-xs px-2.5 py-1">
            {{ selected.size }} selected
          </Badge>
          <div class="flex items-center gap-1.5">
            <Button size="sm" variant="outline" class="h-8 text-xs gap-1" :disabled="bulkBusy" @click="bulk('start')">
              <Play class="size-3 fill-current" />
              <span>Start</span>
            </Button>
            <Button size="sm" variant="outline" class="h-8 text-xs gap-1" :disabled="bulkBusy" @click="bulk('stop')">
              <Square class="size-3 fill-current" />
              <span>Stop</span>
            </Button>
            <Button size="sm" variant="outline" class="h-8 text-xs gap-1" :disabled="bulkBusy" @click="bulk('restart')">
              <RotateCcw class="size-3" />
              <span>Restart</span>
            </Button>
            <Button size="sm" variant="destructive" class="h-8 text-xs gap-1" :disabled="bulkBusy" @click="bulk('remove')">
              <Trash2 class="size-3" />
              <span>Remove</span>
            </Button>
          </div>
          <Button size="sm" variant="ghost" class="h-8 text-xs" @click="selected.clear()">
            Clear selection
          </Button>
        </div>

        <!-- Filter bar -->
        <div v-else class="flex flex-wrap items-center gap-3 w-full sm:w-auto">
          <div class="relative w-full sm:w-72">
            <Search class="absolute left-2.5 top-1/2 -translate-y-1/2 size-3.5 text-muted-foreground" />
            <Input
              v-model="query"
              type="search"
              placeholder="Filter by name or image..."
              class="h-8 pl-8 text-xs"
            />
          </div>

          <label class="flex items-center gap-2 text-xs text-muted-foreground select-none cursor-pointer">
            <input
              v-model="onlyRunning"
              type="checkbox"
              class="rounded border-input text-primary focus:ring-ring size-3.5"
            />
            <span>Running only</span>
          </label>
        </div>

        <div class="text-xs text-muted-foreground font-mono">
          {{ shown.length }} container{{ shown.length === 1 ? '' : 's' }}
        </div>
      </div>

      <!-- Container Table Card -->
      <div class="rounded-xl border border-border bg-card overflow-hidden shadow-xs">
        <LoadState :loading="loading" :error="error" :empty="shown.length === 0" what="containers" @retry="reload">
          <template #empty>
            <div class="p-12 text-center space-y-3">
              <div class="size-12 rounded-xl bg-muted text-muted-foreground flex items-center justify-center mx-auto">
                <Boxes class="size-6" />
              </div>
              <template v-if="!data?.length">
                <h3 class="text-base font-semibold text-foreground">No Containers Found</h3>
                <p class="text-xs text-muted-foreground max-w-sm mx-auto">
                  Launch a container via <code>docker run</code> or <code>docker compose up</code> and it will appear here in real-time.
                </p>
              </template>
              <template v-else>
                <h3 class="text-base font-semibold text-foreground">No Matching Containers</h3>
                <p class="text-xs text-muted-foreground">Adjust your search query or disable the "Running only" filter.</p>
                <Button size="sm" variant="outline" class="mt-2 text-xs" @click="query = ''; onlyRunning = false">
                  Reset filters
                </Button>
              </template>
            </div>
          </template>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="w-10">
                  <input
                    type="checkbox"
                    class="rounded border-input text-primary focus:ring-ring size-3.5"
                    aria-label="Select all containers"
                    :checked="allChecked"
                    :indeterminate="someChecked"
                    @change="pickAll(($event.target as HTMLInputElement).checked)"
                  />
                </TableHead>
                <TableHead class="w-8" />
                <TableHead>Name</TableHead>
                <TableHead>Container ID</TableHead>
                <TableHead>Image</TableHead>
                <TableHead>Port(s)</TableHead>
                <TableHead class="text-right">CPU</TableHead>
                <TableHead class="text-right">Memory</TableHead>
                <TableHead class="text-right w-28">Actions</TableHead>
              </TableRow>
            </TableHeader>

            <TableBody v-for="[project, items] in groups" :key="project">
              <!-- Docker Compose Project Header Row -->
              <TableRow v-if="project" class="bg-muted/40 font-medium">
                <TableCell class="w-10">
                  <input
                    type="checkbox"
                    class="rounded border-input text-primary focus:ring-ring size-3.5"
                    :aria-label="`Select all in ${project}`"
                    :checked="items.every((c) => selected.has(c.id))"
                    @change="pickGroup(items, ($event.target as HTMLInputElement).checked)"
                  />
                </TableCell>
                <TableCell class="w-8" />
                <TableCell colspan="4">
                  <button
                    type="button"
                    class="flex items-center gap-1.5 text-xs font-semibold text-foreground hover:text-primary transition-colors cursor-pointer"
                    @click="toggleGroup(project)"
                  >
                    <ChevronDown v-if="!collapsed.has(project)" class="size-3.5" />
                    <ChevronRight v-else class="size-3.5" />
                    <span>{{ project }}</span>
                    <Badge variant="outline" class="font-mono text-[10px] py-0 px-1 ml-1">
                      {{ items.length }}
                    </Badge>
                  </button>
                </TableCell>
                <TableCell class="text-right font-mono text-xs text-muted-foreground">
                  {{ groupStats(items) ? `${groupStats(items)!.cpu.toFixed(1)}%` : '-' }}
                </TableCell>
                <TableCell class="text-right font-mono text-xs text-muted-foreground">
                  {{ groupStats(items) ? bytes(groupStats(items)!.usage) : '-' }}
                </TableCell>
                <TableCell class="text-right" />
              </TableRow>

              <!-- Individual Container Row -->
              <TableRow
                v-for="c in (project && collapsed.has(project) ? [] : items)"
                :key="c.id"
                :class="{ 'bg-muted/20': selected.has(c.id) }"
              >
                <TableCell class="w-10">
                  <input
                    type="checkbox"
                    class="rounded border-input text-primary focus:ring-ring size-3.5"
                    :aria-label="`Select ${c.name}`"
                    :checked="selected.has(c.id)"
                    @change="pick(c.id, ($event.target as HTMLInputElement).checked)"
                  />
                </TableCell>
                <TableCell class="w-8">
                  <StateBadge dot :state="c.state" :label="c.status" />
                </TableCell>
                <TableCell>
                  <RouterLink
                    :to="`/containers/${c.id}`"
                    class="font-medium text-foreground hover:underline text-xs block truncate max-w-[200px]"
                    :class="{ 'pl-3': !!project }"
                    :title="c.name"
                  >
                    {{ c.name }}
                  </RouterLink>
                </TableCell>
                <TableCell>
                  <div class="flex items-center gap-1 font-mono text-xs text-muted-foreground">
                    <span>{{ shortId(c.id) }}</span>
                    <button
                      type="button"
                      class="text-muted-foreground hover:text-foreground cursor-pointer p-0.5"
                      :title="`Copy ID ${c.name}`"
                      @click="copyId(c)"
                    >
                      <Check v-if="copied === c.id" class="size-3 text-emerald-500" />
                      <Copy v-else class="size-3" />
                    </button>
                  </div>
                </TableCell>
                <TableCell>
                  <span class="font-mono text-xs text-muted-foreground truncate block max-w-[180px]" :title="c.image">
                    {{ c.image }}
                  </span>
                </TableCell>
                <TableCell>
                  <PortLinks :ports="c.ports" />
                </TableCell>
                <TableCell class="text-right font-mono text-xs">
                  {{ stats(c.id) ? `${stats(c.id)!.cpu_percent.toFixed(1)}%` : '-' }}
                </TableCell>
                <TableCell class="text-right font-mono text-xs">
                  {{ stats(c.id) ? bytes(stats(c.id)!.mem_usage) : '-' }}
                </TableCell>
                <TableCell class="text-right">
                  <div class="flex items-center justify-end gap-1">
                    <Button
                      v-if="c.state === 'running'"
                      variant="ghost"
                      size="icon-sm"
                      class="size-7"
                      :disabled="busy === c.id"
                      title="Stop container"
                      @click="act(c, 'stop')"
                    >
                      <Square class="size-3 fill-current text-muted-foreground" />
                    </Button>
                    <Button
                      v-else
                      variant="ghost"
                      size="icon-sm"
                      class="size-7 text-emerald-600 hover:text-emerald-700"
                      :disabled="busy === c.id"
                      title="Start container"
                      @click="act(c, 'start')"
                    >
                      <Play class="size-3 fill-current" />
                    </Button>

                    <Button
                      variant="ghost"
                      size="icon-sm"
                      class="size-7"
                      :disabled="busy === c.id"
                      title="Restart container"
                      @click="act(c, 'restart')"
                    >
                      <RotateCcw class="size-3 text-muted-foreground" />
                    </Button>

                    <RouterLink :to="`/containers/${c.id}?tab=logs`">
                      <Button variant="ghost" size="icon-sm" class="size-7" title="View Logs">
                        <FileText class="size-3 text-muted-foreground" />
                      </Button>
                    </RouterLink>

                    <RouterLink v-if="c.state === 'running'" :to="`/containers/${c.id}?tab=terminal`">
                      <Button variant="ghost" size="icon-sm" class="size-7" title="Terminal Exec">
                        <Terminal class="size-3 text-muted-foreground" />
                      </Button>
                    </RouterLink>

                    <Button
                      variant="ghost"
                      size="icon-sm"
                      class="size-7 text-muted-foreground hover:text-destructive"
                      :disabled="busy === c.id"
                      title="Remove container"
                      @click="act(c, 'remove')"
                    >
                      <Trash2 class="size-3" />
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </LoadState>
      </div>
    </div>
  </div>
</template>
