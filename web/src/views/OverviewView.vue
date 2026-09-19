<script setup lang="ts">
import { computed } from 'vue'
import { RouterLink } from 'vue-router'
import { api, type Container, type System } from '@/api'
import { containerAction } from '@/actions'
import { bytes } from '@/format'
import LoadState from '@/components/LoadState.vue'
import { useLoad } from '@/useLoad'

const sys = useLoad(() => api<System>('GET', '/system'), 10000)
const list = useLoad(() => api<Container[]>('GET', '/containers?all=true'), 5000)

const down = computed(() => (list.data.value ?? []).filter((c) => c.state !== 'running'))

async function start(c: Container) {
  if (await containerAction(c.name, c.id, 'start')) list.reload()
}
</script>

<template>
  <main class="page">
    <LoadState :loading="sys.loading.value" :error="sys.error.value" :empty="!sys.data.value" what="info Docker" @retry="sys.reload">
      <section v-if="sys.data.value" class="host" aria-labelledby="host-name">
        <p class="label">Host</p>
        <h1 id="host-name">{{ sys.data.value.name }}</h1>
        <p class="headline">
          <strong>{{ sys.data.value.containers_running }}</strong> dari {{ sys.data.value.containers }} container berjalan
        </p>
        <dl>
          <div><dt>Docker</dt><dd>{{ sys.data.value.server_version }}</dd></div>
          <div><dt>Sistem</dt><dd>{{ sys.data.value.operating_system }}</dd></div>
          <div><dt>Kernel</dt><dd>{{ sys.data.value.kernel_version }} · {{ sys.data.value.architecture }}</dd></div>
          <div><dt>CPU</dt><dd>{{ sys.data.value.ncpu }} core</dd></div>
          <div><dt>Memori</dt><dd>{{ bytes(sys.data.value.mem_total) }}</dd></div>
          <div><dt>Image</dt><dd>{{ sys.data.value.images }}</dd></div>
        </dl>
      </section>
    </LoadState>

    <section class="attention" aria-labelledby="attention-title">
      <div class="page-head">
        <div>
          <h2 id="attention-title">Tidak berjalan</h2>
          <p>Container yang berhenti, keluar, atau sedang restart.</p>
        </div>
        <RouterLink class="btn" to="/containers">Semua container</RouterLink>
      </div>
      <LoadState :loading="list.loading.value" :error="list.error.value" :empty="down.length === 0" what="container" @retry="list.reload">
        <template #empty>
          <strong>Semua container berjalan.</strong>
          Container yang berhenti akan muncul di sini.
        </template>
        <table class="table">
          <thead><tr><th>Nama</th><th>Status</th><th>Image</th><th><span class="sr-only">Aksi</span></th></tr></thead>
          <tbody>
            <tr v-for="c in down" :key="c.id">
              <td><RouterLink :to="`/containers/${c.id}`">{{ c.name }}</RouterLink></td>
              <td data-label="Status"><span class="state" :class="c.state">{{ c.status }}</span></td>
              <td data-label="Image" class="mono">{{ c.image }}</td>
              <td class="actions"><button class="btn btn-sm" type="button" @click="start(c)">Jalankan</button></td>
            </tr>
          </tbody>
        </table>
      </LoadState>
    </section>
  </main>
</template>

<style scoped>
/* hero-card-dark from DESIGN.md: the one dark surface, carrying the host summary. */
.host {
  background: var(--surface-dark);
  color: #fff;
  border-radius: var(--r-lg);
  padding: 40px 48px;
}
.host h1 { color: #fff; overflow-wrap: anywhere; }
.label { margin: 0 0 4px; color: #c9ccd2; }
.headline { font-size: clamp(20px, 3vw, 24px); margin: 16px 0 32px; color: #fff; }
.headline strong { font-weight: 500; font-size: clamp(40px, 7vw, 56px); line-height: 1; margin-right: 4px; }
dl { display: grid; grid-template-columns: repeat(auto-fit, minmax(160px, 1fr)); gap: 16px 32px; margin: 0; }
dt { color: #c9ccd2; }
dd { margin: 2px 0 0; color: #fff; overflow-wrap: anywhere; }
.attention { margin-top: 56px; }
@media (max-width: 767px) {
  .host { padding: 24px 20px; }
  .attention { margin-top: 40px; }
}
</style>
