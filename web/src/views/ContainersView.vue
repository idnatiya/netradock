<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { IconBox, IconPlayerPlay, IconPlayerStop, IconRefresh, IconSearch, IconTrash } from '@tabler/icons-vue'
import { api, type Container } from '@/api'
import { containerAction } from '@/actions'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
import PortLinks from '@/components/PortLinks.vue'
import StateBadge from '@/components/StateBadge.vue'
import { useLoad } from '@/useLoad'

const { data, error, loading, reload } = useLoad(() => api<Container[]>('GET', '/containers?all=true'), 5000)
const query = ref('')
const showStopped = ref(true)
const busy = ref<string | null>(null)

// Compose projects become groups, like Docker Desktop; standalone containers go last.
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

const total = computed(() => data.value?.length ?? 0)
const running = computed(() => data.value?.filter((c) => c.state === 'running').length ?? 0)

async function act(c: Container, action: 'start' | 'stop' | 'restart' | 'remove') {
  busy.value = c.id
  if (await containerAction(c.name, c.id, action, c.state === 'running')) await reload()
  busy.value = null
}
</script>

<template>
  <PageHeader pretitle="Docker" title="Container">
    <template #meta>
      <div v-if="data" class="text-secondary mt-1">{{ running }} berjalan, {{ total - running }} tidak berjalan</div>
    </template>
    <template #actions>
      <div class="d-flex flex-wrap align-items-center gap-3">
        <div class="input-icon flex-grow-1">
          <span class="input-icon-addon"><IconSearch :size="18" /></span>
          <input v-model="query" type="search" class="form-control" placeholder="Cari nama atau image" aria-label="Cari container">
        </div>
        <label class="form-check form-switch mb-0">
          <input v-model="showStopped" class="form-check-input" type="checkbox">
          <span class="form-check-label">Tampilkan yang berhenti</span>
        </label>
      </div>
    </template>
  </PageHeader>

  <div class="page-body">
    <div class="container-xl">
      <div v-if="loading || error || groups.length === 0" class="card">
        <LoadState :loading="loading" :error="error" :empty="groups.length === 0" what="container" @retry="reload">
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

      <div class="row row-cards">
        <div v-for="[project, items] in groups" :key="project" class="col-12">
          <div class="card">
            <div class="card-header">
              <h3 class="card-title text-break-all">{{ project || 'Tanpa compose project' }}</h3>
              <span class="badge bg-secondary-lt ms-2">{{ items.length }}</span>
            </div>
            <div class="table-responsive-md">
              <table class="table card-table table-vcenter table-stack containers">
                <thead>
                  <tr><th>Nama</th><th>Status</th><th>Image</th><th>Port</th><th><span class="visually-hidden">Aksi</span></th></tr>
                </thead>
                <tbody>
                  <tr v-for="c in items" :key="c.id" :aria-busy="busy === c.id">
                    <td><RouterLink :to="`/containers/${c.id}`" class="fw-medium text-break-all">{{ c.name }}</RouterLink></td>
                    <td data-label="Status"><StateBadge :state="c.state" :label="c.status" /></td>
                    <td data-label="Image" class="text-secondary font-monospace text-break-all">{{ c.image }}</td>
                    <td data-label="Port"><PortLinks :ports="c.ports" /></td>
                    <td class="text-end">
                      <div class="btn-list justify-content-md-end">
                        <button v-if="c.state === 'running'" class="btn btn-sm" type="button" :disabled="busy === c.id" @click="act(c, 'stop')"><IconPlayerStop :size="16" class="icon" />Hentikan</button>
                        <button v-else class="btn btn-sm" type="button" :disabled="busy === c.id" @click="act(c, 'start')"><IconPlayerPlay :size="16" class="icon" />Jalankan</button>
                        <button class="btn btn-sm" type="button" :disabled="busy === c.id" @click="act(c, 'restart')"><IconRefresh :size="16" class="icon" />Restart</button>
                        <button class="btn btn-sm btn-ghost-danger" type="button" :disabled="busy === c.id" @click="act(c, 'remove')"><IconTrash :size="16" class="icon" />Hapus</button>
                      </div>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Each compose group is its own table; fixed widths keep columns aligned across groups. */
@media (min-width: 768px) {
  .containers { table-layout: fixed; }
  .containers th:nth-child(1) { width: 24%; }
  .containers th:nth-child(2) { width: 18%; }
  .containers th:nth-child(3) { width: 18%; }
  .containers th:nth-child(4) { width: 12%; }
  .containers th:nth-child(5) { width: 28%; }
}
</style>
