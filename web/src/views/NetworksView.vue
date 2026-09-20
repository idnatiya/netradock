<script setup lang="ts">
import { computed } from 'vue'
import { Network as NetIcon, Trash2 } from 'lucide-vue-next'
import { api, type Network } from '@/api'
import { removeResource } from '@/actions'
import { ago, shortId } from '@/format'
import { useLoad } from '@/useLoad'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import LoadState from '@/components/LoadState.vue'
import PageHeader from '@/components/PageHeader.vue'

// Docker predefined networks cannot be removed
const BUILTIN = new Set(['bridge', 'host', 'none'])

const { data, error, loading, reload } = useLoad(() => api<Network[]>('GET', '/networks'))
const sorted = computed(() => [...(data.value ?? [])].sort((a, b) => a.name.localeCompare(b.name)))

async function remove(n: Network) {
  if (await removeResource('Network', n.name, `/networks/${n.id}`, 'Docker rejects removal if any container is still connected.')) await reload()
}
</script>

<template>
  <div>
    <!-- Page Header -->
    <PageHeader title="Networks">
      <template #meta>
        <span v-if="data" class="font-mono text-sm text-muted-foreground">
          {{ data.length }} network{{ data.length === 1 ? '' : 's' }}
        </span>
      </template>
    </PageHeader>

    <div class="p-6 w-full space-y-4">
      <div class="rounded-xl border border-border bg-card overflow-hidden shadow-xs">
        <LoadState :loading="loading" :error="error" :empty="sorted.length === 0" what="networks" @retry="reload">
          <template #empty>
            <div class="p-12 text-center space-y-3">
              <div class="size-12 rounded-xl bg-muted text-muted-foreground flex items-center justify-center mx-auto">
                <NetIcon class="size-6" />
              </div>
              <h3 class="text-base font-semibold text-foreground">No Networks Found</h3>
            </div>
          </template>

          <Table>
            <TableHeader>
              <TableRow>
                <TableHead class="text-sm">Network Name</TableHead>
                <TableHead class="text-sm">Network ID</TableHead>
                <TableHead class="text-sm">Driver</TableHead>
                <TableHead class="text-sm">Scope</TableHead>
                <TableHead class="text-sm">Created</TableHead>
                <TableHead class="text-sm text-right w-24">Action</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              <TableRow v-for="n in sorted" :key="n.id">
                <TableCell class="font-mono text-sm font-semibold text-foreground break-all">
                  {{ n.name }}
                </TableCell>
                <TableCell class="font-mono text-xs text-muted-foreground">
                  {{ shortId(n.id) }}
                </TableCell>
                <TableCell class="text-xs text-muted-foreground font-mono">
                  {{ n.driver }}
                </TableCell>
                <TableCell class="text-xs text-muted-foreground font-mono">
                  {{ n.scope }}
                </TableCell>
                <TableCell class="text-xs text-muted-foreground whitespace-nowrap">
                  {{ ago(n.created) }}
                </TableCell>
                <TableCell class="text-right">
                  <Badge v-if="BUILTIN.has(n.name)" variant="secondary" class="font-mono text-xs py-0.5 px-2">
                    builtin
                  </Badge>
                  <Button
                    v-else
                    variant="ghost"
                    size="icon-sm"
                    class="size-7 text-muted-foreground hover:text-destructive"
                    :title="`Delete network ${n.name}`"
                    @click="remove(n)"
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
