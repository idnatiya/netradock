<script setup lang="ts">
import { computed } from 'vue'
import { api, type Network } from '@/api'
import { removeResource } from '@/actions'
import { ago, shortId } from '@/format'
import LoadState from '@/components/LoadState.vue'
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
  <main class="page">
    <div class="page-head">
      <div>
        <h1>Network</h1>
        <p v-if="data">{{ data.length }} network</p>
      </div>
    </div>
    <LoadState :loading="loading" :error="error" :empty="sorted.length === 0" what="network" @retry="reload">
      <template #empty><strong>Tidak ada network.</strong></template>
      <table class="table">
        <thead><tr><th>Nama</th><th>ID</th><th>Driver</th><th>Scope</th><th>Dibuat</th><th><span class="sr-only">Aksi</span></th></tr></thead>
        <tbody>
          <tr v-for="n in sorted" :key="n.id">
            <td class="mono name">{{ n.name }}</td>
            <td data-label="ID" class="mono">{{ shortId(n.id) }}</td>
            <td data-label="Driver">{{ n.driver }}</td>
            <td data-label="Scope">{{ n.scope }}</td>
            <td data-label="Dibuat">{{ ago(n.created) }}</td>
            <td class="actions">
              <span v-if="BUILTIN.has(n.name)" class="sub">bawaan Docker</span>
              <button v-else class="btn btn-sm btn-danger" type="button" @click="remove(n)">Hapus</button>
            </td>
          </tr>
        </tbody>
      </table>
    </LoadState>
  </main>
</template>

<style scoped>
.name { overflow-wrap: anywhere; }
</style>
