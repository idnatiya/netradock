<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import {
  IconBox,
  IconCheck,
  IconChevronRight,
  IconCpu,
  IconDatabase,
  IconDeviceSdCard,
  IconFlame,
  IconPlayerPlay,
  IconRefresh,
  IconServer,
  IconStack2,
} from '@tabler/icons-vue'
import { api, type Container, type Image, type System, type Volume } from '@/api'
import { containerAction } from '@/actions'
import { bytes } from '@/format'
import { useLiveStats } from '@/liveStats'
import LineChart from '@/components/LineChart.vue'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
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
const down = computed(() => containers.value.filter((c) => c.state !== 'running'))
const names = computed(() => Object.fromEntries(containers.value.map((c) => [c.id, c.name])))
const busiest = computed(() => live.busiest.value.slice(0, 6))
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

const barColor = (v: number) => (v >= 80 ? 'bg-red' : v >= 50 ? 'bg-yellow' : 'bg-primary')
const cpuColor = computed(() => (cpuNow.value >= 80 ? 'var(--tblr-danger)' : cpuNow.value >= 50 ? 'var(--tblr-warning)' : 'var(--tblr-primary)'))

async function start(c: Container) {
  if (await containerAction(c.name, c.id, 'start')) list.reload()
}
</script>

<template>
  <PageHeader pretitle="Ringkasan" :title="sys.data.value?.name ?? 'Host Docker'">
    <template #meta>
      <div v-if="sys.data.value" class="d-flex flex-wrap align-items-center gap-2 mt-2">
        <span class="badge bg-blue-lt d-inline-flex align-items-center gap-1">
          <IconServer :size="13" />
          Docker v{{ sys.data.value.server_version }}
        </span>
        <span class="badge bg-secondary-lt">
          {{ sys.data.value.operating_system }} ({{ sys.data.value.architecture }})
        </span>
        <span class="badge bg-secondary-lt font-monospace">
          {{ sys.data.value.ncpu }} Cores · {{ bytes(sys.data.value.mem_total) }} RAM
        </span>
        <span
          class="badge d-inline-flex align-items-center"
          :class="live.connected.value ? 'bg-green-lt text-green' : 'bg-secondary-lt text-secondary'"
        >
          <span class="status-dot me-1.5" :class="live.connected.value ? 'status-green status-dot-animated' : 'status-secondary'"></span>
          {{ live.connected.value ? 'Live Stream' : 'Menghubungkan...' }}
        </span>
      </div>
    </template>
    <template #actions>
      <button
        type="button"
        class="btn btn-outline-secondary btn-sm d-inline-flex align-items-center gap-1"
        :disabled="refreshing"
        @click="refreshAll"
      >
        <IconRefresh :size="15" :class="{ 'spin-animation': refreshing }" />
        Segarkan
      </button>
    </template>
  </PageHeader>

  <div class="page-body">
    <div class="container-xl">
      <div v-if="sys.loading.value || sys.error.value" class="card mb-3">
        <LoadState :loading="sys.loading.value" :error="sys.error.value" :empty="!sys.data.value" what="info Docker" @retry="sys.reload" />
      </div>

      <div class="row row-deck row-cards">
        <!-- 1. CPU container -->
        <div class="col-sm-6 col-lg-3">
          <div class="card card-kpi">
            <div class="card-body">
              <div class="d-flex align-items-center justify-content-between mb-2">
                <span class="subheader mb-0">CPU Container</span>
                <span class="avatar avatar-sm bg-primary-lt rounded-2"><IconCpu :size="18" /></span>
              </div>
              <div class="stat-value mb-1">{{ pct(cpuNow) }}</div>
              <div class="text-secondary small mb-3">dari {{ ncpu }} core host</div>
              <div class="chart-spark-wrapper">
                <Sparkline :values="cpuHost" :floor="5" :color="cpuColor" />
              </div>
            </div>
          </div>
        </div>

        <!-- 2. Memori container -->
        <div class="col-sm-6 col-lg-3">
          <div class="card card-kpi">
            <div class="card-body">
              <div class="d-flex align-items-center justify-content-between mb-2">
                <span class="subheader mb-0">Memori Container</span>
                <span class="avatar avatar-sm bg-azure-lt rounded-2"><IconDeviceSdCard :size="18" /></span>
              </div>
              <div class="stat-value mb-1">{{ bytes(memNow) }}</div>
              <div class="text-secondary small mb-2">{{ pct(memPct) }} dari {{ bytes(memTotal) }}</div>
              <div class="progress progress-sm">
                <div
                  class="progress-bar"
                  :class="barColor(memPct)"
                  :style="{ width: `${Math.min(100, memPct)}%` }"
                  role="progressbar"
                  :aria-valuenow="Math.round(memPct)"
                  aria-valuemin="0"
                  aria-valuemax="100"
                  aria-label="Memori terpakai"
                ></div>
              </div>
            </div>
          </div>
        </div>

        <!-- 3. Containers summary -->
        <div class="col-sm-6 col-lg-3">
          <RouterLink to="/containers" class="card card-kpi card-link">
            <div class="card-body">
              <div class="d-flex align-items-center justify-content-between mb-2">
                <span class="subheader mb-0">Container</span>
                <span class="avatar avatar-sm bg-green-lt rounded-2"><IconBox :size="18" /></span>
              </div>
              <div class="stat-value mb-1">
                {{ running }}<span class="fs-4 text-secondary font-weight-normal"> / {{ containers.length }}</span>
              </div>
              <div class="d-flex align-items-center gap-2 small text-secondary mb-2">
                <span class="text-green fw-medium d-inline-flex align-items-center gap-1">
                  <span class="badge-dot bg-green"></span> {{ running }} aktif
                </span>
                <span v-if="down.length" class="text-secondary d-inline-flex align-items-center gap-1">
                  <span class="badge-dot bg-secondary"></span> {{ down.length }} berhenti
                </span>
              </div>
              <div class="progress progress-sm">
                <div class="progress-bar bg-green" :style="{ width: `${runningPct}%` }"></div>
              </div>
            </div>
          </RouterLink>
        </div>

        <!-- 4. Images summary -->
        <div class="col-sm-6 col-lg-3">
          <RouterLink to="/images" class="card card-kpi card-link">
            <div class="card-body">
              <div class="d-flex align-items-center justify-content-between mb-2">
                <span class="subheader mb-0">Image & Storage</span>
                <span class="avatar avatar-sm bg-purple-lt rounded-2"><IconStack2 :size="18" /></span>
              </div>
              <div class="stat-value mb-1">
                {{ images.data.value?.length ?? sys.data.value?.images ?? '-' }}
                <span class="fs-4 text-secondary font-weight-normal">image</span>
              </div>
              <div class="text-secondary small mb-2">Total {{ bytes(imageSize) }} di disk</div>
              <div class="d-flex justify-content-end">
                <span class="badge bg-purple-lt">Kelola Image →</span>
              </div>
            </div>
          </RouterLink>
        </div>

        <!-- Usage history (Area charts) -->
        <div class="col-lg-8">
          <div class="card">
            <div class="card-header d-flex justify-content-between align-items-center">
              <div>
                <h3 class="card-title">Aktivitas 60 Detik Terakhir</h3>
                <div class="text-secondary small mt-0.5">Pemantauan konsumsi resource kontainer secara real-time</div>
              </div>
              <div class="card-actions">
                <span class="badge" :class="live.connected.value ? 'bg-green-lt text-green' : 'bg-secondary-lt text-secondary'">
                  <span class="status-dot me-1.5" :class="live.connected.value ? 'status-green status-dot-animated' : 'status-secondary'"></span>
                  {{ live.connected.value ? 'Live WebSocket' : 'Offline' }}
                </span>
              </div>
            </div>
            <div class="card-body">
              <div class="row g-4">
                <div class="col-md-6">
                  <LineChart
                    title="CPU (% host)"
                    :values="cpuHost"
                    :capacity="60"
                    :format="pct"
                    :floor="5"
                    color="var(--tblr-primary)"
                  />
                </div>
                <div class="col-md-6">
                  <LineChart
                    title="Memori"
                    :values="live.total.mem"
                    :capacity="60"
                    :format="bytes"
                    :floor="64 * 1024 * 1024"
                    color="#206bc4"
                  />
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Busiest containers -->
        <div class="col-lg-4">
          <div class="card">
            <div class="card-header">
              <h3 class="card-title d-flex align-items-center gap-1.5">
                <IconFlame :size="16" class="text-danger" />
                Container Paling Sibuk
              </h3>
            </div>
            <div v-if="busiest.length === 0" class="card-body text-secondary text-center py-4">
              <p class="mb-0">Menunggu data statistik kontainer...</p>
            </div>
            <div v-else class="list-group list-group-flush">
              <RouterLink
                v-for="[id, s] in busiest"
                :key="id"
                :to="`/containers/${id}`"
                class="list-group-item list-group-item-action d-flex align-items-center gap-3 py-2.5"
              >
                <div class="busiest-avatar">
                  <IconBox :size="16" class="text-primary" />
                </div>
                <div class="min-w-0 flex-grow-1">
                  <div class="d-flex justify-content-between align-items-baseline gap-2 mb-1">
                    <span class="text-truncate fw-medium">{{ names[id] ?? id.slice(0, 12) }}</span>
                    <span class="badge bg-secondary-lt font-monospace" style="font-size: 11px;">
                      {{ pct(s.cpu_percent / ncpu) }}
                    </span>
                  </div>
                  <div class="d-flex align-items-center gap-2">
                    <div class="progress progress-xs flex-grow-1">
                      <div
                        class="progress-bar"
                        :class="barColor(s.cpu_percent / ncpu)"
                        :style="{ width: `${Math.min(100, Math.max(2, s.cpu_percent / ncpu))}%` }"
                      ></div>
                    </div>
                    <span class="text-secondary small font-monospace" style="font-size: 11px;">{{ bytes(s.mem_usage) }}</span>
                  </div>
                </div>
                <IconChevronRight :size="16" class="text-secondary flex-shrink-0 opacity-50" />
              </RouterLink>
            </div>
          </div>
        </div>

        <!-- Docker Storage Breakdown -->
        <div class="col-lg-6">
          <div class="card h-100">
            <div class="card-header d-flex justify-content-between align-items-center">
              <div>
                <h3 class="card-title d-flex align-items-center gap-1.5">
                  <IconDatabase :size="16" class="text-primary" />
                  Alokasi Penyimpanan Docker
                </h3>
                <div class="text-secondary small mt-0.5">Estimasi kapasitas disk terpakai oleh objek Docker</div>
              </div>
              <span class="badge bg-primary-lt font-monospace">{{ bytes(imageSize) }}</span>
            </div>
            <div class="card-body d-flex flex-column justify-content-between">
              <div>
                <!-- Segmented Multi-Bar -->
                <div class="progress mb-3 rounded-2" style="background: var(--tblr-border-color-translucent); overflow: hidden; height: 10px;">
                  <div
                    class="progress-bar bg-primary"
                    role="progressbar"
                    :style="{ width: `${storageDistribution.imagesPct}%` }"
                    :title="`Images: ${bytes(imageSize)}`"
                  ></div>
                  <div
                    class="progress-bar bg-warning"
                    role="progressbar"
                    :style="{ width: `${storageDistribution.volumesPct}%` }"
                    :title="`Volumes: ${volumes.data.value?.length ?? 0} volume`"
                  ></div>
                  <div
                    class="progress-bar bg-green"
                    role="progressbar"
                    :style="{ width: `${storageDistribution.containersPct}%` }"
                    :title="`Containers: ${containers.length} total`"
                  ></div>
                </div>

                <!-- Storage breakdown legend & stats -->
                <div class="row g-3">
                  <div class="col-4">
                    <div class="d-flex align-items-center gap-1.5 mb-1">
                      <span class="badge-dot bg-primary"></span>
                      <span class="text-secondary small fw-medium">Images</span>
                    </div>
                    <div class="fw-bold font-monospace fs-4">{{ bytes(imageSize) }}</div>
                    <div class="text-secondary small">{{ images.data.value?.length ?? 0 }} image</div>
                  </div>
                  <div class="col-4">
                    <div class="d-flex align-items-center gap-1.5 mb-1">
                      <span class="badge-dot bg-warning"></span>
                      <span class="text-secondary small fw-medium">Volumes</span>
                    </div>
                    <div class="fw-bold font-monospace fs-4">{{ volumes.data.value?.length ?? 0 }}</div>
                    <div class="text-secondary small">Volume disk</div>
                  </div>
                  <div class="col-4">
                    <div class="d-flex align-items-center gap-1.5 mb-1">
                      <span class="badge-dot bg-green"></span>
                      <span class="text-secondary small fw-medium">Containers</span>
                    </div>
                    <div class="fw-bold font-monospace fs-4">{{ containers.length }}</div>
                    <div class="text-secondary small">{{ running }} aktif</div>
                  </div>
                </div>
              </div>

              <!-- Quick action links -->
              <div class="d-flex justify-content-between align-items-center pt-3 mt-3 border-top">
                <span class="text-secondary small">Kelola sumber daya penyimpanan</span>
                <div class="d-flex gap-2">
                  <RouterLink to="/images" class="btn btn-sm btn-ghost-primary">
                    Lihat Image →
                  </RouterLink>
                  <RouterLink to="/volumes" class="btn btn-sm btn-ghost-secondary">
                    Lihat Volume →
                  </RouterLink>
                </div>
              </div>
            </div>
          </div>
        </div>

        <!-- Stopped containers -->
        <div class="col-lg-6">
          <div class="card h-100">
            <div class="card-header d-flex justify-content-between align-items-center">
              <div>
                <h3 class="card-title">Container Berhenti</h3>
                <div class="text-secondary small mt-0.5">Daftar container yang tidak sedang berjalan</div>
              </div>
              <div class="card-actions">
                <RouterLink to="/containers" class="btn btn-sm btn-ghost-secondary">
                  Lihat Semua Container →
                </RouterLink>
              </div>
            </div>
            <LoadState :loading="list.loading.value" :error="list.error.value" :empty="down.length === 0" what="container" @retry="list.reload">
              <template #empty>
                <div class="text-center py-5">
                  <div class="avatar avatar-md bg-green-lt rounded-circle mx-auto mb-3">
                    <IconCheck :size="24" />
                  </div>
                  <h4 class="mb-1">Semua Container Berjalan</h4>
                  <p class="text-secondary small mb-0">Tidak ada container yang sedang berhenti atau error saat ini.</p>
                </div>
              </template>
              <div class="table-responsive-md">
                <table class="table card-table table-vcenter table-stack">
                  <thead>
                    <tr>
                      <th>Nama Container</th>
                      <th>Status</th>
                      <th>Image</th>
                      <th class="w-1"></th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr v-for="c in down" :key="c.id">
                      <td>
                        <RouterLink :to="`/containers/${c.id}`" class="text-break-all fw-medium">
                          {{ c.name }}
                        </RouterLink>
                      </td>
                      <td data-label="Status">
                        <StateBadge :state="c.state" :label="c.status" />
                      </td>
                      <td data-label="Image" class="text-secondary font-monospace text-break-all small">
                        {{ c.image }}
                      </td>
                      <td class="text-end">
                        <button class="btn btn-sm btn-outline-primary d-inline-flex align-items-center gap-1" type="button" @click="start(c)">
                          <IconPlayerPlay :size="14" class="icon" />
                          Jalankan
                        </button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </LoadState>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.card-kpi {
  border-radius: var(--tblr-border-radius);
  overflow: hidden;
}
.badge-dot {
  width: 6px;
  height: 6px;
  border-radius: 50%;
  display: inline-block;
}
.chart-spark-wrapper {
  margin-top: 0.5rem;
  margin-bottom: -0.5rem;
}
.busiest-avatar {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  border-radius: 6px;
  background: var(--tblr-bg-surface-secondary);
  flex-shrink: 0;
}
.spin-animation {
  animation: spin 1s linear infinite;
}
@keyframes spin {
  from { transform: rotate(0deg); }
  to { transform: rotate(360deg); }
}
</style>
