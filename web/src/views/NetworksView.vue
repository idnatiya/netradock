<script setup lang="ts">
import { computed } from 'vue'
import { api, type Network } from '@/api'
import { removeResource } from '@/actions'
import { ago, shortId } from '@/format'
import { IconNetwork, IconTrash } from '@tabler/icons-vue'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
import { useLoad } from '@/useLoad'

// Docker's predefined networks cannot be removed.
const BUILTIN = new Set(['bridge', 'host', 'none'])

const { data, error, loading, reload } = useLoad(() => api<Network[]>('GET', '/networks'))
const sorted = computed(() => [...(data.value ?? [])].sort((a, b) => a.name.localeCompare(b.name)))

async function remove(n: Network) {
  if (await removeResource('Network', n.name, `/networks/${n.id}`, 'Docker menolak jika masih ada container yang terhubung.')) await reload()
}
</script>

<template>
  <PageHeader pretitle="Docker" title="Network">
    <template #meta>
      <div v-if="data" class="text-secondary mt-1">{{ data.length }} network</div>
    </template>
  </PageHeader>
  <div class="page-body">
    <div class="container-xl">
      <div class="card">
        <LoadState :loading="loading" :error="error" :empty="sorted.length === 0" what="network" @retry="reload">
          <template #empty>
            <div class="empty-icon"><IconNetwork :size="40" /></div>
            <p class="empty-title">Tidak ada network</p>
          </template>
          <div class="table-responsive-md">
            <table class="table card-table table-vcenter table-stack">
              <thead><tr><th>Nama</th><th>ID</th><th>Driver</th><th>Scope</th><th>Dibuat</th><th class="w-1"><span class="visually-hidden">Aksi</span></th></tr></thead>
              <tbody>
                <tr v-for="n in sorted" :key="n.id">
                  <td class="font-monospace text-break-all">{{ n.name }}</td>
                  <td data-label="ID" class="font-monospace text-secondary">{{ shortId(n.id) }}</td>
                  <td data-label="Driver">{{ n.driver }}</td>
                  <td data-label="Scope">{{ n.scope }}</td>
                  <td data-label="Dibuat" class="text-secondary">{{ ago(n.created) }}</td>
                  <td class="text-end">
                    <span v-if="BUILTIN.has(n.name)" class="badge bg-secondary-lt">bawaan Docker</span>
                    <button v-else class="btn btn-sm btn-ghost-danger" type="button" @click="remove(n)"><IconTrash :size="16" class="icon" />Hapus</button>
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </LoadState>
      </div>
    </div>
  </div>
</template>
