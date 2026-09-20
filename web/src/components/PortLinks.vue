<script setup lang="ts">
import { computed } from 'vue'
import { ExternalLink } from 'lucide-vue-next'
import type { Port } from '@/api'

const props = defineProps<{ ports: Port[] }>()
const host = location.hostname

// Docker lists each binding twice (IPv4 and IPv6); show one per public port.
const shown = computed(() => {
  const seen = new Set<string>()
  return props.ports.filter((p) => {
    const key = `${p.public_port ?? ''}:${p.private_port}/${p.type}`
    if (seen.has(key)) return false
    seen.add(key)
    return true
  })
})
</script>

<template>
  <span v-if="shown.length === 0" class="text-xs text-muted-foreground font-mono">-</span>
  <div v-else class="inline-flex flex-wrap gap-1.5 font-mono text-xs">
    <template v-for="p in shown" :key="`${p.public_port}:${p.private_port}/${p.type}`">
      <a
        v-if="p.public_port && p.type === 'tcp' && p.ip !== '127.0.0.1' && p.ip !== '::1'"
        :href="`http://${host}:${p.public_port}`"
        target="_blank"
        rel="noopener"
        class="inline-flex items-center gap-1 text-primary hover:underline"
      >
        <span>{{ p.public_port }}:{{ p.private_port }}</span>
        <ExternalLink class="size-2.5 opacity-70" />
      </a>
      <span v-else class="text-muted-foreground">
        {{ p.public_port ? `${p.public_port}:` : '' }}{{ p.private_port }}/{{ p.type }}
      </span>
    </template>
  </div>
</template>
