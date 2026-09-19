<script setup lang="ts">
import { computed, ref } from 'vue'
import { RouterLink } from 'vue-router'
import { api, type Container } from '@/api'
import { containerAction } from '@/actions'
import LoadState from '@/components/LoadState.vue'
import PortLinks from '@/components/PortLinks.vue'
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
  <main class="page">
    <div class="page-head">
      <div>
        <h1>Container</h1>
        <p v-if="data">{{ running }} berjalan, {{ total - running }} tidak berjalan</p>
      </div>
      <div class="filters">
        <label class="sr-only" for="q">Cari container</label>
        <input id="q" v-model="query" type="search" placeholder="Cari nama atau image">
        <label class="check"><input v-model="showStopped" type="checkbox"> Tampilkan yang berhenti</label>
      </div>
    </div>

    <LoadState :loading="loading" :error="error" :empty="groups.length === 0" what="container" @retry="reload">
      <template #empty>
        <template v-if="total === 0">
          <strong>Belum ada container.</strong>
          Jalankan container lewat <code>docker run</code> atau <code>docker compose up</code> di server, nanti muncul di sini.
        </template>
        <template v-else>
          <strong>Tidak ada yang cocok.</strong>
          Ubah kata kunci atau centang "Tampilkan yang berhenti".
        </template>
      </template>

      <section v-for="[project, items] in groups" :key="project" class="group">
        <h2>{{ project || 'Tanpa compose project' }} <span class="count">{{ items.length }}</span></h2>
        <table class="table">
          <thead>
            <tr><th>Nama</th><th>Status</th><th>Image</th><th>Port</th><th><span class="sr-only">Aksi</span></th></tr>
          </thead>
          <tbody>
            <tr v-for="c in items" :key="c.id" :aria-busy="busy === c.id">
              <td><RouterLink :to="`/containers/${c.id}`" class="name">{{ c.name }}</RouterLink></td>
              <td data-label="Status"><span class="state" :class="c.state">{{ c.status }}</span></td>
              <td data-label="Image" class="mono image">{{ c.image }}</td>
              <td data-label="Port"><PortLinks :ports="c.ports" /></td>
              <td class="actions">
                <button v-if="c.state === 'running'" class="btn btn-sm" type="button" :disabled="busy === c.id" @click="act(c, 'stop')">Hentikan</button>
                <button v-else class="btn btn-sm" type="button" :disabled="busy === c.id" @click="act(c, 'start')">Jalankan</button>
                <button class="btn btn-sm" type="button" :disabled="busy === c.id" @click="act(c, 'restart')">Restart</button>
                <button class="btn btn-sm btn-danger" type="button" :disabled="busy === c.id" @click="act(c, 'remove')">Hapus</button>
              </td>
            </tr>
          </tbody>
        </table>
      </section>
    </LoadState>
  </main>
</template>

<style scoped>
.filters { display: flex; flex-wrap: wrap; align-items: center; gap: 8px 16px; }
.filters input[type='search'] { width: 260px; }
.group + .group { margin-top: 40px; }
/* Each compose group is its own table; fixed widths keep columns aligned across groups. */
.table { table-layout: fixed; }
.table th:nth-child(1) { width: 26%; }
.table th:nth-child(2) { width: 18%; }
.table th:nth-child(3) { width: 20%; }
.table th:nth-child(4) { width: 13%; }
.table th:nth-child(5) { width: 23%; }
.group h2 { font-size: 16px; font-weight: 500; margin-bottom: 4px; overflow-wrap: anywhere; }
.count { color: var(--muted); font-weight: 400; margin-left: 4px; }
.name { font-weight: 500; overflow-wrap: anywhere; }
.image { overflow-wrap: anywhere; }
.actions .btn + .btn { margin-left: 4px; }
@media (max-width: 767px) {
  .filters, .filters input[type='search'] { width: 100%; }
  .actions .btn + .btn { margin-left: 0; }
}
</style>
