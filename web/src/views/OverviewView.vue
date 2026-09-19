<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { IconBox, IconCpu, IconDeviceSdCard, IconPlayerPlay, IconStack2 } from '@tabler/icons-vue'
import { api, type Container, type Image, type System } from '@/api'
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
const live = useLiveStats()

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
const barColor = (v: number) => (v >= 80 ? 'bg-red' : v >= 50 ? 'bg-yellow' : 'bg-primary')

async function start(c: Container) {
  if (await containerAction(c.name, c.id, 'start')) list.reload()
}
</script>

<template>
  <PageHeader pretitle="Ringkasan" :title="sys.data.value?.name ?? 'Host Docker'">
    <template #meta>
      <div v-if="sys.data.value" class="text-secondary mt-1">
        Docker {{ sys.data.value.server_version }} · {{ sys.data.value.operating_system }} · {{ sys.data.value.ncpu }} core · {{ bytes(sys.data.value.mem_total) }}
      </div>
    </template>
  </PageHeader>

  <div class="page-body">
    <div class="container-xl">
      <div v-if="sys.loading.value || sys.error.value" class="card mb-3">
        <LoadState :loading="sys.loading.value" :error="sys.error.value" :empty="!sys.data.value" what="info Docker" @retry="sys.reload" />
      </div>

      <div class="row row-deck row-cards">
        <!-- Live usage tiles -->
        <div class="col-sm-6 col-lg-3">
          <div class="card">
            <div class="card-body">
              <div class="d-flex align-items-center">
                <div class="subheader">CPU container</div>
                <span class="ms-auto avatar avatar-sm bg-primary-lt"><IconCpu :size="18" /></span>
              </div>
              <div class="h1 mb-1">{{ pct(cpuNow) }}</div>
              <div class="text-secondary small mb-2">dari {{ ncpu }} core host</div>
              <Sparkline :values="cpuHost" :floor="5" />
            </div>
          </div>
        </div>
        <div class="col-sm-6 col-lg-3">
          <div class="card">
            <div class="card-body">
              <div class="d-flex align-items-center">
                <div class="subheader">Memori container</div>
                <span class="ms-auto avatar avatar-sm bg-azure-lt"><IconDeviceSdCard :size="18" /></span>
              </div>
              <div class="h1 mb-1">{{ bytes(memNow) }}</div>
              <div class="text-secondary small mb-2">{{ pct((memNow / memTotal) * 100) }} dari {{ bytes(memTotal) }}</div>
              <div class="progress progress-sm mb-2">
                <div class="progress-bar" :class="barColor((memNow / memTotal) * 100)" :style="{ width: `${Math.min(100, (memNow / memTotal) * 100)}%` }" role="progressbar" :aria-valuenow="Math.round((memNow / memTotal) * 100)" aria-valuemin="0" aria-valuemax="100" aria-label="Memori terpakai"></div>
              </div>
            </div>
          </div>
        </div>
        <div class="col-sm-6 col-lg-3">
          <RouterLink to="/containers" class="card card-link">
            <div class="card-body">
              <div class="d-flex align-items-center">
                <div class="subheader">Container</div>
                <span class="ms-auto avatar avatar-sm bg-green-lt"><IconBox :size="18" /></span>
              </div>
              <div class="h1 mb-1">{{ running }}<span class="fs-3 text-secondary"> / {{ containers.length }}</span></div>
              <div class="text-secondary small mb-2">berjalan, {{ down.length }} tidak berjalan</div>
              <div class="progress progress-sm">
                <div class="progress-bar bg-green" :style="{ width: `${containers.length ? (running / containers.length) * 100 : 0}%` }"></div>
              </div>
            </div>
          </RouterLink>
        </div>
        <div class="col-sm-6 col-lg-3">
          <RouterLink to="/images" class="card card-link">
            <div class="card-body">
              <div class="d-flex align-items-center">
                <div class="subheader">Image</div>
                <span class="ms-auto avatar avatar-sm bg-purple-lt"><IconStack2 :size="18" /></span>
              </div>
              <div class="h1 mb-1">{{ images.data.value?.length ?? sys.data.value?.images ?? '-' }}</div>
              <div class="text-secondary small">total {{ bytes(imageSize) }} di disk</div>
            </div>
          </RouterLink>
        </div>

        <!-- Usage history -->
        <div class="col-lg-8">
          <div class="card">
            <div class="card-header">
              <h3 class="card-title">Pemakaian 60 detik terakhir</h3>
              <div class="card-actions">
                <span class="badge" :class="live.connected.value ? 'bg-green-lt' : 'bg-secondary-lt'">
                  <span class="status-dot me-1" :class="live.connected.value ? 'status-green status-dot-animated' : 'status-secondary'"></span>
                  {{ live.connected.value ? 'Live' : 'Menghubungkan' }}
                </span>
              </div>
            </div>
            <div class="card-body">
              <div class="row g-4">
                <div class="col-md-6"><LineChart title="CPU (% host)" :values="cpuHost" :capacity="60" :format="pct" :floor="5" /></div>
                <div class="col-md-6"><LineChart title="Memori" :values="live.total.mem" :capacity="60" :format="bytes" :floor="64 * 1024 * 1024" /></div>
              </div>
            </div>
          </div>
        </div>

        <!-- Busiest containers -->
        <div class="col-lg-4">
          <div class="card">
            <div class="card-header"><h3 class="card-title">Paling sibuk</h3></div>
            <div v-if="busiest.length === 0" class="card-body text-secondary">Menunggu data statistik...</div>
            <div v-else class="list-group list-group-flush">
              <RouterLink v-for="[id, s] in busiest" :key="id" :to="`/containers/${id}`" class="list-group-item list-group-item-action">
                <div class="d-flex justify-content-between gap-2 mb-1">
                  <span class="text-truncate">{{ names[id] ?? id.slice(0, 12) }}</span>
                  <span class="text-secondary text-nowrap">{{ pct(s.cpu_percent / ncpu) }} · {{ bytes(s.mem_usage) }}</span>
                </div>
                <div class="progress progress-xs">
                  <div class="progress-bar" :class="barColor(s.cpu_percent / ncpu)" :style="{ width: `${Math.min(100, Math.max(1, s.cpu_percent / ncpu))}%` }"></div>
                </div>
              </RouterLink>
            </div>
          </div>
        </div>

        <!-- Not running -->
        <div class="col-12">
          <div class="card">
            <div class="card-header">
              <h3 class="card-title">Tidak berjalan</h3>
              <div class="card-actions"><RouterLink to="/containers" class="btn btn-sm">Semua container</RouterLink></div>
            </div>
            <LoadState :loading="list.loading.value" :error="list.error.value" :empty="down.length === 0" what="container" @retry="list.reload">
              <template #empty>
                <p class="empty-title">Semua container berjalan</p>
                <p class="empty-subtitle text-secondary">Container yang berhenti akan muncul di sini.</p>
              </template>
              <div class="table-responsive-md">
                <table class="table card-table table-vcenter table-stack">
                  <thead><tr><th>Nama</th><th>Status</th><th>Image</th><th class="w-1"></th></tr></thead>
                  <tbody>
                    <tr v-for="c in down" :key="c.id">
                      <td><RouterLink :to="`/containers/${c.id}`" class="text-break-all">{{ c.name }}</RouterLink></td>
                      <td data-label="Status"><StateBadge :state="c.state" :label="c.status" /></td>
                      <td data-label="Image" class="text-secondary font-monospace text-break-all">{{ c.image }}</td>
                      <td class="text-end"><button class="btn btn-sm btn-primary" type="button" @click="start(c)"><IconPlayerPlay :size="16" class="icon" />Jalankan</button></td>
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
