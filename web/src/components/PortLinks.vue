<script setup lang="ts">
import { computed } from 'vue'
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
  <span v-if="shown.length === 0" class="none">Tidak ada</span>
  <span v-else class="ports">
    <template v-for="p in shown" :key="`${p.public_port}:${p.private_port}/${p.type}`">
      <a
        v-if="p.public_port && p.type === 'tcp' && p.ip !== '127.0.0.1' && p.ip !== '::1'"
        :href="`http://${host}:${p.public_port}`"
        target="_blank"
        rel="noopener"
        class="font-monospace"
      >{{ p.public_port }}:{{ p.private_port }}</a>
      <span v-else class="font-monospace">{{ p.public_port ? `${p.public_port}:` : '' }}{{ p.private_port }}/{{ p.type }}</span>
    </template>
  </span>
</template>

<style scoped>
.ports { display: inline-flex; flex-wrap: wrap; gap: 4px 12px; }
.none { color: var(--tblr-secondary); }
</style>
