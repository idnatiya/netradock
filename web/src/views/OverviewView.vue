<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { IconBox, IconCpu, IconPlayerPlay, IconPlayerStop, IconStack2 } from '@tabler/icons-vue'
import { api, type Container, type System } from '@/api'
import { containerAction } from '@/actions'
import { bytes } from '@/format'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
import StateBadge from '@/components/StateBadge.vue'
import { useLoad } from '@/useLoad'

const sys = useLoad(() => api<System>('GET', '/system'), 10000)
const list = useLoad(() => api<Container[]>('GET', '/containers?all=true'), 5000)

const down = computed(() => (list.data.value ?? []).filter((c) => c.state !== 'running'))

async function start(c: Container) {
  if (await containerAction(c.name, c.id, 'start')) list.reload()
}
</script>

<template>
  <PageHeader pretitle="Ringkasan" :title="sys.data.value?.name ?? 'Host Docker'" />
  <div class="page-body">
    <div class="container-xl">
      <LoadState :loading="sys.loading.value" :error="sys.error.value" :empty="!sys.data.value" what="info Docker" @retry="sys.reload">
        <div v-if="sys.data.value" class="row row-cards">
          <div class="col-sm-6 col-lg-3">
            <RouterLink to="/containers" class="card card-sm card-link">
              <div class="card-body d-flex align-items-center gap-3">
                <span class="avatar bg-green-lt"><IconPlayerPlay :size="20" /></span>
                <div>
                  <div class="h3 mb-0">{{ sys.data.value.containers_running }}</div>
                  <div class="text-secondary">Container berjalan</div>
                </div>
              </div>
            </RouterLink>
          </div>
          <div class="col-sm-6 col-lg-3">
            <RouterLink to="/containers" class="card card-sm card-link">
              <div class="card-body d-flex align-items-center gap-3">
                <span class="avatar" :class="sys.data.value.containers_stopped ? 'bg-red-lt' : 'bg-secondary-lt'"><IconPlayerStop :size="20" /></span>
                <div>
                  <div class="h3 mb-0">{{ sys.data.value.containers_stopped + sys.data.value.containers_paused }}</div>
                  <div class="text-secondary">Berhenti atau jeda</div>
                </div>
              </div>
            </RouterLink>
          </div>
          <div class="col-sm-6 col-lg-3">
            <RouterLink to="/images" class="card card-sm card-link">
              <div class="card-body d-flex align-items-center gap-3">
                <span class="avatar bg-blue-lt"><IconStack2 :size="20" /></span>
                <div>
                  <div class="h3 mb-0">{{ sys.data.value.images }}</div>
                  <div class="text-secondary">Image</div>
                </div>
              </div>
            </RouterLink>
          </div>
          <div class="col-sm-6 col-lg-3">
            <div class="card card-sm">
              <div class="card-body d-flex align-items-center gap-3">
                <span class="avatar bg-secondary-lt"><IconCpu :size="20" /></span>
                <div>
                  <div class="h3 mb-0">{{ sys.data.value.ncpu }} core · {{ bytes(sys.data.value.mem_total) }}</div>
                  <div class="text-secondary">Sumber daya host</div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </LoadState>

      <div class="row row-cards mt-0">
        <div class="col-lg-8">
          <div class="card">
            <div class="card-header">
              <div>
                <h3 class="card-title">Tidak berjalan</h3>
                <p class="card-subtitle">Container yang berhenti, keluar, atau sedang restart.</p>
              </div>
              <div class="card-actions">
                <RouterLink to="/containers" class="btn btn-sm">Semua container</RouterLink>
              </div>
            </div>
            <LoadState :loading="list.loading.value" :error="list.error.value" :empty="down.length === 0" what="container" @retry="list.reload">
              <template #empty>
                <div class="empty-icon"><IconBox :size="40" /></div>
                <p class="empty-title">Semua container berjalan</p>
                <p class="empty-subtitle text-secondary">Container yang berhenti akan muncul di sini.</p>
              </template>
              <div class="table-responsive-md">
                <table class="table card-table table-vcenter table-stack">
                  <thead><tr><th>Nama</th><th>Status</th><th>Image</th><th class="w-1"><span class="visually-hidden">Aksi</span></th></tr></thead>
                  <tbody>
                    <tr v-for="c in down" :key="c.id">
                      <td><RouterLink :to="`/containers/${c.id}`" class="text-break-all">{{ c.name }}</RouterLink></td>
                      <td data-label="Status"><StateBadge :state="c.state" :label="c.status" /></td>
                      <td data-label="Image" class="text-secondary font-monospace text-break-all">{{ c.image }}</td>
                      <td class="text-end">
                        <button class="btn btn-sm" type="button" @click="start(c)"><IconPlayerPlay :size="16" class="icon" />Jalankan</button>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </LoadState>
          </div>
        </div>
        <div v-if="sys.data.value" class="col-lg-4">
          <div class="card">
            <div class="card-header"><h3 class="card-title">Host</h3></div>
            <div class="card-body">
              <div class="datagrid">
                <div class="datagrid-item"><div class="datagrid-title">Docker</div><div class="datagrid-content">{{ sys.data.value.server_version }}</div></div>
                <div class="datagrid-item"><div class="datagrid-title">Arsitektur</div><div class="datagrid-content">{{ sys.data.value.architecture }}</div></div>
                <div class="datagrid-item"><div class="datagrid-title">Sistem</div><div class="datagrid-content text-break-all">{{ sys.data.value.operating_system }}</div></div>
                <div class="datagrid-item"><div class="datagrid-title">Kernel</div><div class="datagrid-content text-break-all">{{ sys.data.value.kernel_version }}</div></div>
                <div class="datagrid-item"><div class="datagrid-title">Total container</div><div class="datagrid-content">{{ sys.data.value.containers }}</div></div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
