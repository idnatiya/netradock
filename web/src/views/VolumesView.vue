<script setup lang="ts">
import { computed } from 'vue'
import { api, type Volume } from '@/api'
import { removeResource } from '@/actions'
import { IconDatabase, IconTrash } from '@tabler/icons-vue'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
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
  <PageHeader title="Volumes">
    <template #meta>
      <div v-if="data" class="text-secondary mt-1">{{ data.length }} volume</div>
    </template>
  </PageHeader>
  <div class="page-body">
    <div class="container-xl">
      <div>
        <LoadState :loading="loading" :error="error" :empty="sorted.length === 0" what="volume" @retry="reload">
          <template #empty>
            <div class="empty-icon"><IconDatabase :size="40" /></div>
            <p class="empty-title">Belum ada volume</p>
            <p class="empty-subtitle text-secondary">Volume dibuat otomatis saat container memakai <code>-v nama:/path</code> atau lewat compose.</p>
          </template>
          <div class="table-responsive-md">
            <table class="table table-vcenter table-stack dd-table">
              <thead><tr><th>Nama</th><th>Driver</th><th>Mountpoint</th><th>Dibuat</th><th class="w-1"><span class="visually-hidden">Aksi</span></th></tr></thead>
              <tbody>
                <tr v-for="v in sorted" :key="v.name">
                  <td class="font-monospace text-break-all">{{ v.name }}</td>
                  <td data-label="Driver">{{ v.driver }}</td>
                  <td data-label="Mountpoint" class="font-monospace text-secondary text-break-all">{{ v.mountpoint }}</td>
                  <td data-label="Dibuat" class="text-secondary text-nowrap">{{ date(v.created_at) }}</td>
                  <td class="text-end"><button class="btn btn-icon btn-ghost-danger" type="button" :aria-label="`Hapus ${v.name}`" title="Hapus" @click="remove(v)"><IconTrash :size="18" /></button></td>
                </tr>
              </tbody>
            </table>
          </div>
        </LoadState>
      </div>
    </div>
  </div>
</template>
