<script setup lang="ts">
import { computed, ref } from 'vue'
import { Download, Layers, Loader2, Trash2 } from 'lucide-vue-next'
import { api, notify, type Image } from '@/api'
import { removeResource } from '@/actions'
import { ago, bytes, shortId } from '@/format'
import { useLoad } from '@/useLoad'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'

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
    notify('ok', `${ref_.value.trim()} successfully pulled.`)
    ref_.value = ''
    await reload()
  } catch (e) {
    notify('error', `Pull failed: ${(e as Error).message}`)
  } finally {
    pulling.value = false
  }
}

async function remove(i: Image) {
  const label = i.tags[0] ?? shortId(i.id)
  const body = i.containers > 0
    ? `In use by ${i.containers} container(s). Docker will reject removal unless containers are stopped and deleted first.`
    : 'Image can be pulled again at any time.'
  if (await removeResource('Image', label, `/images/${encodeURIComponent(i.id)}`, body)) await reload()
}
</script>

<template>
  <div>
    <!-- Page Header -->
    <PageHeader title="Images">
      <template #meta>
        <span v-if="data" class="font-mono text-xs text-muted-foreground">
          {{ data.length }} images, total {{ bytes(totalSize) }}
        </span>
      </template>
      <template #actions>
        <form class="flex items-center gap-2 w-full sm:w-auto" @submit.prevent="pull">
          <Input
            v-model="ref_"
            type="text"
            placeholder="e.g. nginx:alpine or redis:latest"
            class="h-8.5 w-full sm:w-64 text-xs font-mono"
            required
            :disabled="pulling"
            autocomplete="off"
            spellcheck="false"
          />
          <Button type="submit" size="sm" class="h-8.5 text-xs gap-1.5 shrink-0" :disabled="pulling || !ref_.trim()">
            <Loader2 v-if="pulling" class="size-3.5 animate-spin" />
            <Download v-else class="size-3.5" />
            <span>{{ pulling ? 'Pulling...' : 'Pull image' }}</span>
          </Button>
        </form>
      </template>
    </PageHeader>

    <div class="p-6 w-full space-y-6">
      <!-- 4 KPI Summary Cards -->
      <div v-if="data" class="grid grid-cols-2 lg:grid-cols-4 gap-4">
        <Card>
          <CardHeader class="pb-1 p-4">
            <CardTitle class="text-xs uppercase tracking-wider text-muted-foreground">Total Images</CardTitle>
          </CardHeader>
          <CardContent class="p-4 pt-0">
            <div class="text-2xl font-bold font-mono">{{ data.length }}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="pb-1 p-4">
            <CardTitle class="text-xs uppercase tracking-wider text-muted-foreground">Disk Consumption</CardTitle>
          </CardHeader>
          <CardContent class="p-4 pt-0">
            <div class="text-2xl font-bold font-mono text-primary">{{ bytes(totalSize) }}</div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="pb-1 p-4">
            <CardTitle class="text-xs uppercase tracking-wider text-muted-foreground">Dangling / Unused</CardTitle>
          </CardHeader>
          <CardContent class="p-4 pt-0">
            <div class="text-2xl font-bold font-mono text-amber-500">
              {{ unused.length }}
              <span class="text-xs font-normal text-muted-foreground">({{ bytes(unusedSize) }})</span>
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardHeader class="pb-1 p-4">
            <CardTitle class="text-xs uppercase tracking-wider text-muted-foreground">Untagged</CardTitle>
          </CardHeader>
          <CardContent class="p-4 pt-0">
            <div class="text-2xl font-bold font-mono text-muted-foreground">{{ untagged }}</div>
          </CardContent>
        </Card>
      </div>

      <!-- Images Table -->
      <div class="rounded-xl border border-border bg-card overflow-hidden shadow-xs">
        <LoadState :loading="loading" :error="error" :empty="sorted.length === 0" what="images" @retry="reload">
          <template #empty>
            <div class="p-12 text-center space-y-3">
              <div class="size-12 rounded-xl bg-muted text-muted-foreground flex items-center justify-center mx-auto">
                <Layers class="size-6" />
              </div>
              <h3 class="text-base font-semibold text-foreground">No Images Found</h3>
              <p class="text-sm text-muted-foreground max-w-sm mx-auto">
                Enter an image name in the box above (e.g. <code>nginx:alpine</code>) and click "Pull image".
              </p>
            </div>
          </template>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="text-sm">Repository / Tag</TableHead>
                <TableHead class="text-sm">Image ID</TableHead>
                <TableHead class="text-sm">Size</TableHead>
                <TableHead class="text-sm">Created</TableHead>
                <TableHead class="text-sm">In Use</TableHead>
                <TableHead class="text-sm text-right w-16">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="i in sorted" :key="i.id">
                <TableCell>
                  <div class="space-y-0.5">
                    <span v-for="t in i.tags" :key="t" class="font-mono text-sm font-semibold text-foreground block">
                      {{ t }}
                    </span>
                    <span v-if="i.tags.length === 0" class="text-xs text-muted-foreground italic font-mono">
                      &lt;none&gt;:&lt;none&gt;
                    </span>
                  </div>
                </TableCell>
                <TableCell class="font-mono text-xs text-muted-foreground">
                  {{ shortId(i.id) }}
                </TableCell>
                <TableCell class="font-mono text-xs">
                  {{ bytes(i.size) }}
                </TableCell>
                <TableCell class="text-xs text-muted-foreground">
                  {{ ago(i.created) }}
                </TableCell>
                <TableCell>
                  <Badge v-if="i.containers > 0" variant="success" class="text-xs font-mono py-0.5 px-2">
                    {{ i.containers }} container{{ i.containers === 1 ? '' : 's' }}
                  </Badge>
                  <span v-else class="text-xs text-muted-foreground font-mono">unused</span>
                </TableCell>
                <TableCell class="text-right">
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    class="size-7 text-muted-foreground hover:text-destructive"
                    :title="`Delete image ${i.tags[0] ?? shortId(i.id)}`"
                    @click="remove(i)"
                  >
                    <Trash2 class="size-3.5" />
                  </Button>
                </TableCell>
              </TableRow>
            </TableBody>
          </Table>
        </LoadState>
      </div>
    </div>
  </div>
</template>
