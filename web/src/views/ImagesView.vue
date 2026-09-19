<script setup lang="ts">
import { computed, ref } from 'vue'
import { api, notify, type Image } from '@/api'
import { removeResource } from '@/actions'
import { ago, bytes, shortId } from '@/format'
import { IconDownload, IconStack2, IconTrash } from '@tabler/icons-vue'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'
import { useLoad } from '@/useLoad'

const { data, error, loading, reload } = useLoad(() => api<Image[]>('GET', '/images'))
const sorted = computed(() => [...(data.value ?? [])].sort((a, b) => b.created - a.created))
const totalSize = computed(() => (data.value ?? []).reduce((n, i) => n + i.size, 0))
const unused = computed(() => (data.value ?? []).filter((i) => i.containers === 0))
const unusedSize = computed(() => unused.value.reduce((n, i) => n + i.size, 0))
const untagged = computed(() => (data.value ?? []).filter((i) => i.tags.length === 0).length)

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
  <PageHeader pretitle="Docker" title="Image">
    <template #meta>
      <div v-if="data" class="text-secondary mt-1">{{ data.length }} image, total {{ bytes(totalSize) }}</div>
    </template>
    <template #actions>
      <form class="input-group pull" @submit.prevent="pull">
        <input v-model="ref_" type="text" class="form-control" placeholder="nginx:alpine" aria-label="Nama image yang mau di-pull" required :disabled="pulling" autocomplete="off" spellcheck="false">
        <button class="btn btn-primary" type="submit" :disabled="pulling || !ref_.trim()">
          <span v-if="pulling" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
          <IconDownload v-else :size="18" class="icon" />
          {{ pulling ? 'Menarik...' : 'Pull image' }}
        </button>
      </form>
    </template>
  </PageHeader>

  <div class="page-body">
    <div class="container-xl">
      <div v-if="data" class="row row-cards mb-3">
        <div class="col-6 col-lg-3"><div class="card card-sm"><div class="card-body">
          <div class="subheader">Total image</div><div class="h2 mb-0">{{ data.length }}</div>
        </div></div></div>
        <div class="col-6 col-lg-3"><div class="card card-sm"><div class="card-body">
          <div class="subheader">Ukuran di disk</div><div class="h2 mb-0">{{ bytes(totalSize) }}</div>
        </div></div></div>
        <div class="col-6 col-lg-3"><div class="card card-sm"><div class="card-body">
          <div class="subheader">Tidak dipakai</div><div class="h2 mb-0">{{ unused.length }} <span class="fs-4 text-secondary">· {{ bytes(unusedSize) }}</span></div>
        </div></div></div>
        <div class="col-6 col-lg-3"><div class="card card-sm"><div class="card-body">
          <div class="subheader">Tanpa tag</div><div class="h2 mb-0">{{ untagged }}</div>
        </div></div></div>
      </div>
      <div class="card">
        <LoadState :loading="loading" :error="error" :empty="sorted.length === 0" what="image" @retry="reload">
          <template #empty>
            <div class="empty-icon"><IconStack2 :size="40" /></div>
            <p class="empty-title">Belum ada image</p>
            <p class="empty-subtitle text-secondary">Tulis nama image di kolom atas, misalnya <code>nginx:alpine</code>, lalu pilih Pull image.</p>
          </template>
          <div class="table-responsive-md">
            <table class="table card-table table-vcenter table-stack">
              <thead><tr><th>Tag</th><th>ID</th><th>Ukuran</th><th>Dibuat</th><th>Dipakai</th><th class="w-1"><span class="visually-hidden">Aksi</span></th></tr></thead>
              <tbody>
                <tr v-for="i in sorted" :key="i.id">
                  <td>
                    <div class="tags">
                      <span v-for="t in i.tags" :key="t" class="font-monospace">{{ t }}</span>
                      <span v-if="i.tags.length === 0" class="text-secondary">tanpa tag</span>
                    </div>
                  </td>
                  <td data-label="ID" class="font-monospace text-secondary">{{ shortId(i.id) }}</td>
                  <td data-label="Ukuran">{{ bytes(i.size) }}</td>
                  <td data-label="Dibuat" class="text-secondary">{{ ago(i.created) }}</td>
                  <td data-label="Dipakai">
                    <span v-if="i.containers > 0" class="badge bg-green-lt">{{ i.containers }} container</span>
                    <span v-else class="text-secondary">tidak</span>
                  </td>
                  <td class="text-end"><button class="btn btn-icon btn-ghost-danger" type="button" :aria-label="`Hapus ${i.tags[0] ?? shortId(i.id)}`" title="Hapus" @click="remove(i)"><IconTrash :size="18" /></button></td>
                </tr>
              </tbody>
            </table>
          </div>
        </LoadState>
      </div>
    </div>
  </div>
</template>

<style scoped>
.pull { min-width: min(420px, 100%); }
.tags { display: grid; gap: 2px; overflow-wrap: anywhere; }
</style>
