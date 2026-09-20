<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  Activity,
  ArrowRight,
  Boxes,
  Check,
  ChevronRight,
  Cpu,
  Flame,
  HardDrive,
  Layers,
  Pause,
  Play,
  RotateCcw,
  Server,
  StopCircle,
} from 'lucide-vue-next'
import { api, type Container, type Image, type System, type Volume } from '@/api'
import { containerAction } from '@/actions'
import { bytes } from '@/format'
import { useLiveStats } from '@/liveStats'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import LineChart from '@/components/LineChart.vue'
import LoadState from '@/components/LoadState.vue'
import Sparkline from '@/components/Sparkline.vue'
import StateBadge from '@/components/StateBadge.vue'
import { useLoad } from '@/useLoad'

const sys = useLoad(() => api<System>('GET', '/system'), 10000)
const list = useLoad(() => api<Container[]>('GET', '/containers?all=true'), 5000)
const images = useLoad(() => api<Image[]>('GET', '/images'), 30000)
const volumes = useLoad(() => api<Volume[]>('GET', '/volumes'), 30000)
const live = useLiveStats()

const refreshing = ref(false)
async function refreshAll() {
  refreshing.value = true
  await Promise.allSettled([sys.reload(), list.reload(), images.reload(), volumes.reload()])
  refreshing.value = false
}

const ncpu = computed(() => sys.data.value?.ncpu || 1)
const memTotal = computed(() => sys.data.value?.mem_total || 1)
// Docker reports CPU per core (100% = one core); divide by core count for share of the host.
const cpuHost = computed(() => live.total.cpu.map((v) => v / ncpu.value))
const cpuNow = computed(() => cpuHost.value.at(-1) ?? 0)
const memNow = computed(() => live.total.mem.at(-1) ?? 0)

const containers = computed(() => list.data.value ?? [])
const running = computed(() => containers.value.filter((c) => c.state === 'running').length)
const paused = computed(() => containers.value.filter((c) => c.state === 'paused').length)
const stopped = computed(() => containers.value.filter((c) => c.state !== 'running' && c.state !== 'paused').length)
const down = computed(() => containers.value.filter((c) => c.state !== 'running'))
const names = computed(() => Object.fromEntries(containers.value.map((c) => [c.id, c.name])))
const busiest = computed(() => live.busiest.value.slice(0, 5))
const imageSize = computed(() => (images.data.value ?? []).reduce((n, i) => n + i.size, 0))

const pct = (v: number) => `${v.toFixed(1)}%`
const memPct = computed(() => (memNow.value / memTotal.value) * 100)
const runningPct = computed(() => (containers.value.length ? (running.value / containers.value.length) * 100 : 0))

const storageDistribution = computed(() => {
  const imgBytes = imageSize.value
  const volEstimate = (volumes.data.value?.length ?? 0) * 100 * 1024 * 1024
  const contEstimate = containers.value.length * 20 * 1024 * 1024
  const total = Math.max(1, imgBytes + volEstimate + contEstimate)
  return {
    imagesPct: Math.max(10, Math.round((imgBytes / total) * 100)),
    volumesPct: Math.max(5, Math.round((volEstimate / total) * 100)),
    containersPct: Math.max(5, Math.round((contEstimate / total) * 100)),
  }
})

const cpuColor = computed(() => (cpuNow.value >= 80 ? '#ff453a' : cpuNow.value >= 50 ? '#ff9f0a' : '#30d158'))

async function start(c: Container) {
  if (await containerAction(c.name, c.id, 'start')) list.reload()
}
</script>

<template>
  <div class="p-6 w-full space-y-6">
    <!-- Page Header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-4 pb-3 border-b border-border">
      <div class="space-y-1">
        <div class="flex items-center gap-2">
          <h1 class="text-2xl font-bold tracking-tight text-foreground">Dashboard</h1>
          <Badge v-if="live.connected.value" variant="success" class="gap-1.5 font-mono text-xs px-2 py-0.5">
            <span class="size-1.5 rounded-full bg-emerald-500 animate-pulse-dot" />
            LIVE
          </Badge>
        </div>
        <p class="text-sm text-muted-foreground">
          Real-time container infrastructure telemetry and host diagnostics.
        </p>
      </div>

      <div class="flex items-center gap-2.5">
        <Button
          variant="outline"
          size="sm"
          :disabled="refreshing"
          class="gap-1.5 text-xs font-medium h-8.5 px-3"
          @click="refreshAll"
        >
          <RotateCcw class="size-3.5" :class="{ 'animate-spin': refreshing }" />
          <span>Refresh</span>
        </Button>
        <RouterLink to="/containers">
          <Button
            size="sm"
            class="gap-2 text-xs font-semibold h-8.5 px-3.5 bg-primary text-primary-foreground shadow-sm hover:bg-primary/90 border border-primary/20 cursor-pointer"
          >
            <Boxes class="size-4" />
            <span>Manage Containers</span>
            <ArrowRight class="size-3.5 opacity-80" />
          </Button>
        </RouterLink>
      </div>
    </div>

    <!-- Error Banner if Docker daemon unreachable -->
    <div v-if="sys.loading.value || sys.error.value">
      <LoadState :loading="sys.loading.value" :error="sys.error.value" :empty="!sys.data.value" what="Docker info" @retry="sys.reload" />
    </div>

    <!-- 1. Top Stat KPI Cards -->
    <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Running -->
      <Card class="hover:border-emerald-500/40 transition-colors">
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Running
          </CardTitle>
          <div class="size-8 rounded-lg bg-emerald-500/10 text-emerald-600 dark:text-emerald-400 flex items-center justify-center">
            <Play class="size-4 fill-current" />
          </div>
        </CardHeader>
        <CardContent class="space-y-1">
          <div class="text-3xl font-bold tracking-tight text-foreground">{{ running }}</div>
          <p class="text-xs text-muted-foreground font-mono">
            {{ pct(runningPct) }} of total containers
          </p>
        </CardContent>
      </Card>

      <!-- Paused -->
      <Card class="hover:border-amber-500/40 transition-colors">
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Paused
          </CardTitle>
          <div class="size-8 rounded-lg bg-amber-500/10 text-amber-600 dark:text-amber-400 flex items-center justify-center">
            <Pause class="size-4" />
          </div>
        </CardHeader>
        <CardContent class="space-y-1">
          <div class="text-3xl font-bold tracking-tight text-foreground">{{ paused }}</div>
          <p class="text-xs text-muted-foreground font-mono">
            Temporarily suspended
          </p>
        </CardContent>
      </Card>

      <!-- Stopped -->
      <Card class="hover:border-rose-500/40 transition-colors">
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Stopped
          </CardTitle>
          <div class="size-8 rounded-lg bg-rose-500/10 text-rose-600 dark:text-rose-400 flex items-center justify-center">
            <StopCircle class="size-4" />
          </div>
        </CardHeader>
        <CardContent class="space-y-1">
          <div class="text-3xl font-bold tracking-tight text-foreground">{{ stopped }}</div>
          <p class="text-xs text-muted-foreground font-mono">
            {{ containers.length ? pct((stopped / containers.length) * 100) : '0%' }} inactive
          </p>
        </CardContent>
      </Card>

      <!-- Total Containers -->
      <Card class="hover:border-primary/40 transition-colors">
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <CardTitle class="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            Total
          </CardTitle>
          <div class="size-8 rounded-lg bg-primary/10 text-primary flex items-center justify-center">
            <Boxes class="size-4" />
          </div>
        </CardHeader>
        <CardContent class="space-y-1">
          <div class="text-3xl font-bold tracking-tight text-foreground">{{ containers.length }}</div>
          <p class="text-xs text-muted-foreground font-mono">
            {{ images.data.value?.length ?? 0 }} images on disk
          </p>
        </CardContent>
      </Card>
    </div>

    <!-- 2. Host Telemetry Gauges -->
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <!-- CPU Load -->
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <div>
            <CardTitle class="text-sm font-semibold">Host CPU Load</CardTitle>
            <CardDescription class="text-xs">Distributed across {{ ncpu }} physical cores</CardDescription>
          </div>
          <div class="text-right">
            <div class="text-2xl font-bold font-mono text-foreground">{{ pct(cpuNow) }}</div>
            <div class="text-xs font-mono text-muted-foreground">Utilization</div>
          </div>
        </CardHeader>
        <CardContent class="pt-2">
          <div class="h-10 w-full overflow-hidden">
            <Sparkline :values="cpuHost" :floor="5" :color="cpuColor" :height="40" />
          </div>
        </CardContent>
      </Card>

      <!-- RAM Usage -->
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-2">
          <div>
            <CardTitle class="text-sm font-semibold">Host Memory Usage</CardTitle>
            <CardDescription class="text-xs">{{ bytes(memNow) }} of {{ bytes(memTotal) }} consumed</CardDescription>
          </div>
          <div class="text-right">
            <div class="text-2xl font-bold font-mono text-foreground">{{ pct(memPct) }}</div>
            <div class="text-xs font-mono text-muted-foreground">RAM Allocated</div>
          </div>
        </CardHeader>
        <CardContent class="space-y-2 pt-2">
          <div class="h-2 w-full rounded-full bg-muted overflow-hidden">
            <div
              class="h-full rounded-full transition-all duration-300"
              :class="memPct >= 80 ? 'bg-rose-500' : memPct >= 60 ? 'bg-amber-500' : 'bg-sky-500'"
              :style="{ width: `${Math.min(100, memPct)}%` }"
            />
          </div>
          <div class="flex justify-between text-xs font-mono text-muted-foreground">
            <span>Free: {{ bytes(Math.max(0, memTotal - memNow)) }}</span>
            <span>Total: {{ bytes(memTotal) }}</span>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- 3. 60s Historical Charts & Busiest Containers -->
    <div class="grid grid-cols-1 lg:grid-cols-3 gap-6">
      <!-- 60-Second Realtime Charts (2 cols) -->
      <Card class="lg:col-span-2">
        <CardHeader class="flex flex-row items-center justify-between pb-4">
          <div>
            <CardTitle class="text-base font-semibold">Last 60 Seconds Activity</CardTitle>
            <CardDescription class="text-xs">Real-time resource telemetry streaming via WebSocket</CardDescription>
          </div>
          <Badge variant="outline" class="font-mono text-xs gap-1.5">
            <Activity class="size-3 text-emerald-500" />
            <span>Telemetry</span>
          </Badge>
        </CardHeader>
        <CardContent class="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
          <div class="p-4 rounded-lg border border-border bg-muted/30">
            <LineChart
              title="CPU Load"
              :values="cpuHost"
              :capacity="60"
              :format="pct"
              :floor="5"
              color="#30d158"
            />
          </div>
          <div class="p-4 rounded-lg border border-border bg-muted/30">
            <LineChart
              title="Memory Usage"
              :values="live.total.mem"
              :capacity="60"
              :format="bytes"
              :floor="64 * 1024 * 1024"
              color="#38bdf8"
            />
          </div>
        </CardContent>
      </Card>

      <!-- Busiest Containers (1 col) -->
      <Card>
        <CardHeader class="pb-3">
          <div class="flex items-center gap-2">
            <Flame class="size-4 text-rose-500" />
            <CardTitle class="text-base font-semibold">Busiest Containers</CardTitle>
          </div>
          <CardDescription class="text-xs">Highest active CPU consumers</CardDescription>
        </CardHeader>
        <CardContent class="p-0">
          <div v-if="busiest.length === 0" class="p-6 text-center text-xs text-muted-foreground font-mono">
            Awaiting container statistics...
          </div>
          <div v-else class="divide-y divide-border">
            <RouterLink
              v-for="[id, s] in busiest"
              :key="id"
              :to="`/containers/${id}`"
              class="flex items-center justify-between gap-3 px-5 py-3 transition-colors hover:bg-muted/50"
            >
              <div class="min-w-0 flex-1 space-y-1">
                <div class="flex items-center justify-between text-xs">
                  <span class="font-medium text-foreground truncate">{{ names[id] ?? id.slice(0, 12) }}</span>
                  <span class="font-mono text-muted-foreground">{{ pct(s.cpu_percent / ncpu) }}</span>
                </div>
                <div class="h-1.5 w-full rounded-full bg-muted overflow-hidden">
                  <div
                    class="h-full rounded-full bg-rose-500 transition-all"
                    :style="{ width: `${Math.min(100, Math.max(4, s.cpu_percent / ncpu))}%` }"
                  />
                </div>
                <div class="text-xs font-mono text-muted-foreground">
                  {{ bytes(s.mem_usage) }} RAM
                </div>
              </div>
              <ChevronRight class="size-4 text-muted-foreground shrink-0" />
            </RouterLink>
          </div>
        </CardContent>
      </Card>
    </div>

    <!-- 4. Storage Breakdown & Stopped Containers Table -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Storage Breakdown -->
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-3">
          <div>
            <div class="flex items-center gap-2">
              <HardDrive class="size-4 text-primary" />
              <CardTitle class="text-base font-semibold">Docker Storage Allocation</CardTitle>
            </div>
            <CardDescription class="text-xs">Disk capacity consumed by Docker objects</CardDescription>
          </div>
          <Badge variant="secondary" class="font-mono text-xs">
            {{ bytes(imageSize) }}
          </Badge>
        </CardHeader>
        <CardContent class="space-y-4">
          <!-- Segmented Multi-Bar -->
          <div class="h-2.5 w-full rounded-full bg-muted overflow-hidden flex">
            <div
              class="h-full bg-primary"
              :style="{ width: `${storageDistribution.imagesPct}%` }"
              :title="`Images: ${bytes(imageSize)}`"
            />
            <div
              class="h-full bg-amber-500"
              :style="{ width: `${storageDistribution.volumesPct}%` }"
              :title="`Volumes: ${volumes.data.value?.length ?? 0}`"
            />
            <div
              class="h-full bg-emerald-500"
              :style="{ width: `${storageDistribution.containersPct}%` }"
              :title="`Containers: ${containers.length}`"
            />
          </div>

          <!-- Legend -->
          <div class="grid grid-cols-3 gap-4 pt-1">
            <div class="space-y-1">
              <div class="flex items-center gap-1.5 text-xs text-muted-foreground font-medium">
                <span class="size-2 rounded-full bg-primary" />
                <span>Images</span>
              </div>
              <div class="text-base font-bold font-mono text-foreground">{{ bytes(imageSize) }}</div>
              <div class="text-xs text-muted-foreground font-mono">{{ images.data.value?.length ?? 0 }} items</div>
            </div>

            <div class="space-y-1">
              <div class="flex items-center gap-1.5 text-xs text-muted-foreground font-medium">
                <span class="size-2 rounded-full bg-amber-500" />
                <span>Volumes</span>
              </div>
              <div class="text-base font-bold font-mono text-foreground">{{ volumes.data.value?.length ?? 0 }}</div>
              <div class="text-xs text-muted-foreground font-mono">mounted disks</div>
            </div>

            <div class="space-y-1">
              <div class="flex items-center gap-1.5 text-xs text-muted-foreground font-medium">
                <span class="size-2 rounded-full bg-emerald-500" />
                <span>Containers</span>
              </div>
              <div class="text-base font-bold font-mono text-foreground">{{ containers.length }}</div>
              <div class="text-xs text-muted-foreground font-mono">{{ running }} active</div>
            </div>
          </div>

          <!-- Action buttons -->
          <div class="flex items-center justify-between pt-3 border-t border-border">
            <span class="text-xs text-muted-foreground">Manage storage resources</span>
            <div class="flex items-center gap-2">
              <RouterLink to="/images">
                <Button variant="outline" size="sm" class="h-8 text-xs gap-1">
                  <span>Images</span>
                  <ArrowRight class="size-3" />
                </Button>
              </RouterLink>
              <RouterLink to="/volumes">
                <Button variant="outline" size="sm" class="h-8 text-xs gap-1">
                  <span>Volumes</span>
                  <ArrowRight class="size-3" />
                </Button>
              </RouterLink>
            </div>
          </div>
        </CardContent>
      </Card>

      <!-- Stopped Containers Table -->
      <Card>
        <CardHeader class="flex flex-row items-center justify-between pb-3">
          <div>
            <CardTitle class="text-base font-semibold">Stopped Containers</CardTitle>
            <CardDescription class="text-xs">Containers currently not running on this host</CardDescription>
          </div>
          <RouterLink to="/containers">
            <Button variant="ghost" size="sm" class="h-8 text-xs gap-1 text-muted-foreground hover:text-foreground">
              <span>View all</span>
              <ChevronRight class="size-3.5" />
            </Button>
          </RouterLink>
        </CardHeader>
        <CardContent class="p-0">
          <LoadState :loading="list.loading.value" :error="list.error.value" :empty="down.length === 0" what="containers" @retry="list.reload">
            <template #empty>
              <div class="p-8 text-center space-y-2">
                <div class="size-10 rounded-full bg-emerald-500/10 text-emerald-500 flex items-center justify-center mx-auto">
                  <Check class="size-5" />
                </div>
                <h4 class="text-sm font-semibold text-foreground">All Containers Running</h4>
                <p class="text-xs text-muted-foreground max-w-xs mx-auto">
                  Every container registered on this host is active and running healthy.
                </p>
              </div>
            </template>

            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead class="text-sm">Container</TableHead>
                  <TableHead class="text-sm">Status</TableHead>
                  <TableHead class="text-sm text-right">Action</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                <TableRow v-for="c in down.slice(0, 5)" :key="c.id">
                  <TableCell>
                    <RouterLink :to="`/containers/${c.id}`" class="font-medium text-foreground hover:underline truncate block max-w-[180px]">
                      {{ c.name }}
                    </RouterLink>
                    <div class="text-xs font-mono text-muted-foreground truncate max-w-[180px]">
                      {{ c.image }}
                    </div>
                  </TableCell>
                  <TableCell>
                    <StateBadge :state="c.state" :label="c.status" />
                  </TableCell>
                  <TableCell class="text-right">
                    <Button
                      variant="outline"
                      size="sm"
                      class="h-7 text-xs gap-1"
                      @click="start(c)"
                    >
                      <Play class="size-3 fill-current" />
                      <span>Start</span>
                    </Button>
                  </TableCell>
                </TableRow>
              </TableBody>
            </Table>
          </LoadState>
        </CardContent>
      </Card>
    </div>
  </div>
</template>
