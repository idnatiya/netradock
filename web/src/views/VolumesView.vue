<script setup lang="ts">
import { computed } from 'vue'
import { api, type Volume } from '@/api'
import { removeResource } from '@/actions'
import LoadState from '@/components/LoadState.vue'
import { useLoad } from '@/useLoad'

const { data, error, loading, reload } = useLoad(() => api<Volume[]>('GET', '/volumes'))
const sorted = computed(() => [...(data.value ?? [])].sort((a, b) => a.name.localeCompare(b.name)))
const fmt = new Intl.DateTimeFormat('id', { dateStyle: 'medium', timeStyle: 'short' })
const date = (s: string) => (s ? fmt.format(new Date(s)) : '')

async function remove(v: Volume) {
  if (await removeResource('Volume', v.name, `/volumes/${encodeURIComponent(v.name)}`, 'Semua data di volume ini hilang permanen. Docker menolak jika volume masih dipakai container.')) await reload()
}
</script>

<template>
  <main class="page">
    <div class="page-head">
      <div>
        <h1>Volume</h1>
        <p v-if="data">{{ data.length }} volume</p>
      </div>
    </div>
    <LoadState :loading="loading" :error="error" :empty="sorted.length === 0" what="volume" @retry="reload">
      <template #empty>
        <strong>Belum ada volume.</strong>
        Volume dibuat otomatis saat container memakai <code>-v nama:/path</code> atau lewat compose.
      </template>
      <table class="table">
        <thead><tr><th>Nama</th><th>Driver</th><th>Mountpoint</th><th>Dibuat</th><th><span class="sr-only">Aksi</span></th></tr></thead>
        <tbody>
          <tr v-for="v in sorted" :key="v.name">
            <td class="mono name">{{ v.name }}</td>
            <td data-label="Driver">{{ v.driver }}</td>
            <td data-label="Mountpoint" class="mono sub path">{{ v.mountpoint }}</td>
            <td data-label="Dibuat">{{ date(v.created_at) }}</td>
            <td class="actions"><button class="btn btn-sm btn-danger" type="button" @click="remove(v)">Hapus</button></td>
          </tr>
        </tbody>
      </table>
    </LoadState>
  </main>
</template>

<style scoped>
.name, .path { overflow-wrap: anywhere; max-width: 360px; }
@media (max-width: 767px) { .name, .path { max-width: none; } }
</style>
