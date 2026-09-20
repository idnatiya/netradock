<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink, useRoute, useRouter } from 'vue-router'
import {
  Activity,
  Boxes,
  Check,
  Code2,
  Copy,
  FileText,
  Play,
  RotateCcw,
  Square,
  Terminal,
  Trash2,
} from 'lucide-vue-next'
import Sparkline from '@/components/Sparkline.vue'
import { ago, bytes } from '@/format'
import { useLiveStats } from '@/liveStats'
import { api, notify } from '@/api'
import { containerAction } from '@/actions'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
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

const tabs = ['logs', 'stats', 'terminal', 'inspect'] as const
type Tab = (typeof tabs)[number]
const activeTab = computed<Tab>({
  get: () => (tabs.includes(route.query.tab as Tab) ? (route.query.tab as Tab) : 'logs'),
  set: (val: Tab) => router.replace({ query: { ...route.query, tab: val } }),
})

const name = computed(() => data.value?.Name.replace(/^\//, '') ?? '')
const state = computed(() => data.value?.State.Status ?? '')
const running = computed(() => state.value === 'running')

const live = useLiveStats()
const now = computed(() => live.latest.value[data.value?.Id ?? ''])
const startedAt = computed(() => (data.value && running.value ? Date.parse(data.value.State.StartedAt) / 1000 : 0))

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

// Logs: remount terminal to restart stream with new tail size
const tail = ref('200')
const logKey = ref(0)
const logEnded = ref('')

// Exec: connect on demand
const shell = ref('/bin/sh')
const execOn = ref(false)
const execEnded = ref('')
function connect() {
  execEnded.value = ''
  execOn.value = true
}

const copied = ref(false)
async function copyInspect() {
  try {
    await navigator.clipboard.writeText(JSON.stringify(data.value, null, 2))
    copied.value = true
    setTimeout(() => (copied.value = false), 1500)
    notify('ok', 'Inspect JSON copied to clipboard.')
  } catch {
    notify('error', 'Clipboard denied. Please select text manually.')
  }
}
</script>

<template>
  <div>
    <!-- Page Header -->
    <PageHeader pretitle="Container" :title="name || id.slice(0, 12)">
      <template #meta>
        <div v-if="data" class="flex flex-wrap items-center gap-2 mt-1">
          <StateBadge :state="state" :label="state === 'exited' ? `exited (${data.State.ExitCode})` : state" />
          <Badge v-if="data.State.Health" variant="secondary" class="font-mono text-xs">
            health: {{ data.State.Health.Status }}
          </Badge>
          <Badge v-if="data.RestartCount" variant="warning" class="font-mono text-xs">
            restart {{ data.RestartCount }}x
          </Badge>
          <span class="font-mono text-xs text-muted-foreground break-all">
            {{ data.Config.Image }}
          </span>
        </div>
      </template>

      <template v-if="data" #actions>
        <div class="flex items-center gap-1.5">
          <Button
            v-if="running"
            variant="outline"
            size="sm"
            class="h-8 text-xs gap-1.5"
            :disabled="busy"
            @click="act('stop')"
          >
            <Square class="size-3 fill-current text-muted-foreground" />
            <span>Stop</span>
          </Button>
          <Button
            v-else
            size="sm"
            class="h-8 text-xs gap-1.5"
            :disabled="busy"
            @click="act('start')"
          >
            <Play class="size-3 fill-current" />
            <span>Start</span>
          </Button>

          <Button
            variant="outline"
            size="sm"
            class="h-8 text-xs gap-1.5"
            :disabled="busy"
            @click="act('restart')"
          >
            <RotateCcw class="size-3" />
            <span>Restart</span>
          </Button>

          <Button
            variant="destructive"
            size="sm"
            class="h-8 text-xs gap-1.5"
            :disabled="busy"
            @click="act('remove')"
          >
            <Trash2 class="size-3" />
            <span>Remove</span>
          </Button>
        </div>
      </template>
    </PageHeader>

    <div class="p-6 max-w-7xl mx-auto space-y-6">
      <div v-if="loading || !data">
        <LoadState :loading="loading" :error="error" :empty="!data" what="container" @retry="reload">
          <template #empty>
            <div class="p-12 text-center space-y-3">
              <h3 class="text-base font-semibold">Container Not Found</h3>
              <RouterLink to="/containers">
                <Button variant="outline" size="sm">Back to containers</Button>
              </RouterLink>
            </div>
          </template>
        </LoadState>
      </div>

      <template v-if="data">
        <!-- 4 Stat Summary Cards -->
        <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
          <Card>
            <CardHeader class="pb-1 p-4">
              <CardTitle class="text-xs uppercase tracking-wider text-muted-foreground font-mono">CPU</CardTitle>
            </CardHeader>
            <CardContent class="p-4 pt-0 space-y-1">
              <div class="text-xl font-bold font-mono text-foreground">
                {{ now ? `${now.cpu_percent.toFixed(1)}%` : '-' }}
              </div>
              <div class="h-6 w-full overflow-hidden">
                <Sparkline :values="live.history[data.Id]?.cpu ?? []" :height="24" :floor="5" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader class="pb-1 p-4">
              <CardTitle class="text-xs uppercase tracking-wider text-muted-foreground font-mono">Memory</CardTitle>
            </CardHeader>
            <CardContent class="p-4 pt-0 space-y-1">
              <div class="text-xl font-bold font-mono text-foreground">
                {{ now ? bytes(now.mem_usage) : '-' }}
              </div>
              <div class="h-6 w-full overflow-hidden">
                <Sparkline :values="live.history[data.Id]?.mem ?? []" :height="24" :floor="16 * 1024 * 1024" color="#0ea5e9" />
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader class="pb-1 p-4">
              <CardTitle class="text-xs uppercase tracking-wider text-muted-foreground font-mono">Uptime</CardTitle>
            </CardHeader>
            <CardContent class="p-4 pt-0">
              <div class="text-xl font-bold font-mono text-foreground">
                {{ startedAt ? ago(startedAt) : 'Inactive' }}
              </div>
              <p class="text-[11px] text-muted-foreground font-mono mt-0.5">
                created {{ ago(Date.parse(data.Created) / 1000) }}
              </p>
            </CardContent>
          </Card>

          <Card>
            <CardHeader class="pb-1 p-4">
              <CardTitle class="text-xs uppercase tracking-wider text-muted-foreground font-mono">Restarts</CardTitle>
            </CardHeader>
            <CardContent class="p-4 pt-0">
              <div class="text-xl font-bold font-mono text-foreground">
                {{ data.RestartCount }}x
              </div>
              <p class="text-[11px] text-muted-foreground font-mono mt-0.5">
                last exit code {{ data.State.ExitCode }}
              </p>
            </CardContent>
          </Card>
        </div>

        <!-- Tabbed Container Workspace -->
        <Card class="overflow-hidden">
          <Tabs v-model="activeTab" class="w-full">
            <div class="border-b border-border px-5 pt-3 bg-muted/20">
              <TabsList class="h-9 bg-muted/60">
                <TabsTrigger value="logs" class="gap-1.5 text-xs">
                  <FileText class="size-3.5" />
                  <span>Logs</span>
                </TabsTrigger>
                <TabsTrigger value="stats" class="gap-1.5 text-xs">
                  <Activity class="size-3.5" />
                  <span>Stats</span>
                </TabsTrigger>
                <TabsTrigger value="terminal" class="gap-1.5 text-xs">
                  <Terminal class="size-3.5" />
                  <span>Terminal</span>
                </TabsTrigger>
                <TabsTrigger value="inspect" class="gap-1.5 text-xs">
                  <Code2 class="size-3.5" />
                  <span>Inspect</span>
                </TabsTrigger>
              </TabsList>
            </div>

            <!-- Tab 1: Logs -->
            <TabsContent value="logs" class="p-5 mt-0 space-y-4">
              <div class="flex flex-wrap items-center justify-between gap-3">
                <div class="flex items-center gap-2 text-xs">
                  <span class="text-muted-foreground">Tail lines:</span>
                  <select
                    id="tail"
                    v-model="tail"
                    class="h-8 rounded-md border border-input bg-background px-2 text-xs font-mono"
                    @change="logKey++"
                  >
                    <option value="100">100</option>
                    <option value="200">200</option>
                    <option value="1000">1000</option>
                    <option value="5000">5000</option>
                  </select>
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  class="h-8 text-xs gap-1.5"
                  @click="logEnded = ''; logKey++"
                >
                  <RotateCcw class="size-3" />
                  <span>Reload logs</span>
                </Button>
              </div>

              <div class="rounded-lg border border-border overflow-hidden bg-[#181d26]">
                <XTerm
                  :key="`${id}-${logKey}`"
                  :path="`/containers/${id}/logs?tail=${tail}`"
                  label="Container log stream"
                  @closed="(r) => (logEnded = r || 'Log stream closed.')"
                />
              </div>
              <div v-if="logEnded" class="text-xs text-muted-foreground font-mono" role="status">
                {{ logEnded }}
              </div>
            </TabsContent>

            <!-- Tab 2: Stats -->
            <TabsContent value="stats" class="p-5 mt-0">
              <StatsPanel v-if="running" :key="id" :id="id" />
              <div v-else class="p-12 text-center space-y-3">
                <h4 class="text-sm font-semibold">Container Not Running</h4>
                <p class="text-xs text-muted-foreground">Real-time statistics are available only while the container is active.</p>
                <Button size="sm" class="gap-1.5" :disabled="busy" @click="act('start')">
                  <Play class="size-3 fill-current" />
                  <span>Start Container</span>
                </Button>
              </div>
            </TabsContent>

            <!-- Tab 3: Terminal Exec -->
            <TabsContent value="terminal" class="p-5 mt-0 space-y-4">
              <div v-if="!running" class="p-12 text-center space-y-3">
                <h4 class="text-sm font-semibold">Container Not Running</h4>
                <p class="text-xs text-muted-foreground">Start the container to attach an interactive terminal session.</p>
                <Button size="sm" class="gap-1.5" :disabled="busy" @click="act('start')">
                  <Play class="size-3 fill-current" />
                  <span>Start Container</span>
                </Button>
              </div>
              <template v-else>
                <div class="flex flex-wrap items-center justify-between gap-3">
                  <div class="flex items-center gap-2 text-xs">
                    <span class="text-muted-foreground">Shell binary:</span>
                    <select
                      id="shell"
                      v-model="shell"
                      class="h-8 rounded-md border border-input bg-background px-2 text-xs font-mono"
                      :disabled="execOn"
                    >
                      <option>/bin/sh</option>
                      <option>/bin/bash</option>
                      <option>/bin/ash</option>
                    </select>
                  </div>
                  <Button
                    v-if="!execOn"
                    size="sm"
                    class="h-8 text-xs gap-1.5"
                    @click="connect"
                  >
                    <Terminal class="size-3.5" />
                    <span>Connect Shell</span>
                  </Button>
                  <Button
                    v-else
                    variant="destructive"
                    size="sm"
                    class="h-8 text-xs"
                    @click="execOn = false"
                  >
                    Disconnect
                  </Button>
                </div>

                <div v-if="execOn" class="rounded-lg border border-border overflow-hidden bg-[#181d26]">
                  <XTerm
                    :key="`${id}-${shell}`"
                    :path="`/containers/${id}/exec?cmd=${encodeURIComponent(shell)}`"
                    interactive
                    label="Container interactive terminal"
                    @closed="(r) => { execOn = false; execEnded = r || 'Terminal session disconnected.' }"
                  />
                </div>

                <div v-if="execEnded" class="text-xs text-muted-foreground font-mono" role="status">
                  {{ execEnded }}
                </div>

                <div v-if="!execOn && !execEnded" class="p-3 rounded-lg border border-border bg-muted/20 text-xs text-muted-foreground">
                  Commands execute directly within the target container process namespace with root/container privileges.
                </div>
              </template>
            </TabsContent>

            <!-- Tab 4: Inspect JSON -->
            <TabsContent value="inspect" class="p-5 mt-0 space-y-4">
              <div class="flex items-center justify-between">
                <span class="text-xs text-muted-foreground">Raw Docker Engine inspect object</span>
                <Button variant="outline" size="sm" class="h-8 text-xs gap-1.5" @click="copyInspect">
                  <Check v-if="copied" class="size-3.5 text-emerald-500" />
                  <Copy v-else class="size-3.5" />
                  <span>{{ copied ? 'Copied' : 'Copy JSON' }}</span>
                </Button>
              </div>
              <pre class="p-4 rounded-lg border border-border bg-muted/40 font-mono text-xs text-foreground overflow-auto max-h-[60vh] select-text">{{ JSON.stringify(data, null, 2) }}</pre>
            </TabsContent>
          </Tabs>
        </Card>
      </template>
    </div>
  </div>
</template>
