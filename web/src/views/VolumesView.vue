<script setup lang="ts">
import { computed } from 'vue'
import { HardDrive, Trash2 } from 'lucide-vue-next'
import { api, type Volume } from '@/api'
import { removeResource } from '@/actions'
import { useLoad } from '@/useLoad'
import { Button } from '@/components/ui/button'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'

const { data, error, loading, reload } = useLoad(() => api<Volume[]>('GET', '/volumes'))
const sorted = computed(() => [...(data.value ?? [])].sort((a, b) => a.name.localeCompare(b.name)))
const fmt = new Intl.DateTimeFormat('en-US', { dateStyle: 'medium', timeStyle: 'short' })
const date = (s: string) => (s ? fmt.format(new Date(s)) : '-')

async function remove(v: Volume) {
  if (await removeResource('Volume', v.name, `/volumes/${encodeURIComponent(v.name)}`, 'All persistent data in this volume will be permanently deleted. Docker rejects removal if in use by a container.')) await reload()
}
</script>

<template>
  <div>
    <!-- Page Header -->
    <PageHeader title="Volumes">
      <template #meta>
        <span v-if="data" class="font-mono text-sm text-muted-foreground">
          {{ data.length }} persistent volume{{ data.length === 1 ? '' : 's' }}
        </span>
      </template>
    </PageHeader>

    <div class="p-6 w-full space-y-4">
      <div class="rounded-xl border border-border bg-card overflow-hidden shadow-xs">
        <LoadState :loading="loading" :error="error" :empty="sorted.length === 0" what="volumes" @retry="reload">
          <template #empty>
            <div class="p-12 text-center space-y-3">
              <div class="size-12 rounded-xl bg-muted text-muted-foreground flex items-center justify-center mx-auto">
                <HardDrive class="size-6" />
              </div>
              <h3 class="text-base font-semibold text-foreground">No Volumes Found</h3>
              <p class="text-sm text-muted-foreground max-w-sm mx-auto">
                Volumes are automatically created when containers use <code>-v name:/path</code> or via Docker Compose.
              </p>
            </div>
          </template>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="text-sm">Volume Name</TableHead>
                <TableHead class="text-sm">Driver</TableHead>
                <TableHead class="text-sm">Mountpoint</TableHead>
                <TableHead class="text-sm">Created</TableHead>
                <TableHead class="text-sm text-right w-16">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="v in sorted" :key="v.name">
                <TableCell class="font-mono text-sm font-semibold text-foreground break-all">
                  {{ v.name }}
                </TableCell>
                <TableCell class="text-xs text-muted-foreground font-mono">
                  {{ v.driver }}
                </TableCell>
                <TableCell class="font-mono text-xs text-muted-foreground break-all">
                  {{ v.mountpoint }}
                </TableCell>
                <TableCell class="text-xs text-muted-foreground whitespace-nowrap">
                  {{ date(v.created_at) }}
                </TableCell>
                <TableCell class="text-right">
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    class="size-7 text-muted-foreground hover:text-destructive"
                    :title="`Delete volume ${v.name}`"
                    @click="remove(v)"
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
