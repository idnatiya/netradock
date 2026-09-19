<script setup lang="ts">
import { computed, ref } from 'vue'
import { api, notify, type Image } from '@/api'
import { removeResource } from '@/actions'
import { ago, bytes, shortId } from '@/format'
import LoadState from '@/components/LoadState.vue'
import { useLoad } from '@/useLoad'

const { data, error, loading, reload } = useLoad(() => api<Image[]>('GET', '/images'))
const sorted = computed(() => [...(data.value ?? [])].sort((a, b) => b.created - a.created))
const totalSize = computed(() => (data.value ?? []).reduce((n, i) => n + i.size, 0))

const ref_ = ref('')
const pulling = ref(false)
async function pull() {
  pulling.value = true
  try {
    await api('POST', '/images/pull', { image: ref_.value.trim() })
    notify('ok', `${ref_.value.trim()} selesai di-pull.`)
    ref_.value = ''
    await reload()
  } catch (e) {
    notify('error', `Pull gagal: ${(e as Error).message}`)
  } finally {
    pulling.value = false
  }
}

async function remove(i: Image) {
  const label = i.tags[0] ?? shortId(i.id)
  const body = i.containers > 0
    ? `Dipakai ${i.containers} container. Docker akan menolak kecuali container itu dihapus dulu.`
    : 'Image bisa di-pull ulang kapan saja.'
  if (await removeResource('Image', label, `/images/${encodeURIComponent(i.id)}`, body)) await reload()
}
</script>

<template>
  <main class="page">
    <div class="page-head">
      <div>
        <h1>Image</h1>
        <p v-if="data">{{ data.length }} image, total {{ bytes(totalSize) }}</p>
      </div>
      <form class="pull" @submit.prevent="pull">
        <label class="sr-only" for="pull-ref">Nama image</label>
        <input id="pull-ref" v-model="ref_" type="text" placeholder="nginx:alpine" required :disabled="pulling" autocomplete="off" spellcheck="false">
        <button class="btn btn-primary" type="submit" :disabled="pulling || !ref_.trim()">{{ pulling ? 'Menarik...' : 'Pull image' }}</button>
      </form>
    </div>

    <LoadState :loading="loading" :error="error" :empty="sorted.length === 0" what="image" @retry="reload">
      <template #empty>
        <strong>Belum ada image.</strong>
        Tulis nama image di kolom di atas, misalnya <code>nginx:alpine</code>, lalu pilih Pull image.
      </template>
      <table class="table">
        <thead><tr><th>Tag</th><th>ID</th><th>Ukuran</th><th>Dibuat</th><th>Dipakai</th><th><span class="sr-only">Aksi</span></th></tr></thead>
        <tbody>
          <tr v-for="i in sorted" :key="i.id">
            <td class="tags">
              <span v-for="t in i.tags" :key="t" class="mono">{{ t }}</span>
              <span v-if="i.tags.length === 0" class="sub">tanpa tag</span>
            </td>
            <td data-label="ID" class="mono">{{ shortId(i.id) }}</td>
            <td data-label="Ukuran">{{ bytes(i.size) }}</td>
            <td data-label="Dibuat">{{ ago(i.created) }}</td>
            <td data-label="Dipakai">{{ i.containers > 0 ? `${i.containers} container` : 'tidak' }}</td>
            <td class="actions"><button class="btn btn-sm btn-danger" type="button" @click="remove(i)">Hapus</button></td>
          </tr>
        </tbody>
      </table>
    </LoadState>
  </main>
</template>

<style scoped>
.pull { display: flex; gap: 8px; width: min(440px, 100%); }
.tags { display: grid; gap: 2px; overflow-wrap: anywhere; }
</style>
