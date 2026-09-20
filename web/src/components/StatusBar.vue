<script setup lang="ts">
import { computed } from 'vue'
import { IconBrandDocker } from '@tabler/icons-vue'
import type { System } from '@/api'
import { bytes } from '@/format'
import { useLiveStats } from '@/liveStats'

const props = defineProps<{ sys: System | null; error?: string | null }>()
const live = useLiveStats()

const up = computed(() => !!props.sys && !props.error)
// Docker reports CPU per core (100% = one core); show the share of the whole host.
const cpu = computed(() => (live.total.cpu.at(-1) ?? 0) / (props.sys?.ncpu || 1))
const mem = computed(() => live.total.mem.at(-1) ?? 0)
</script>

<template>
  <footer class="statusbar d-flex align-items-center gap-2 px-3 small text-secondary">
    <IconBrandDocker :size="16" class="flex-shrink-0" :class="up ? 'text-primary' : 'text-danger'" />
    <span class="status" :class="up ? 'status-green' : 'status-red'">
      <span class="status-dot"></span>{{ up ? 'Engine berjalan' : 'Engine tidak terjangkau' }}
    </span>
    <template v-if="sys">
      <span class="sep d-none d-md-inline">·</span>
      <span class="d-none d-md-inline tnum">CPU {{ cpu.toFixed(1) }}%</span>
      <span class="sep d-none d-md-inline">·</span>
      <span class="d-none d-md-inline tnum">RAM {{ bytes(mem) }} / {{ bytes(sys.mem_total) }}</span>
      <span class="ms-auto text-truncate d-none d-lg-inline" :title="sys.name">{{ sys.name }}</span>
      <span class="sep d-none d-lg-inline">·</span>
      <span class="d-none d-sm-inline ms-auto ms-lg-0">v{{ sys.server_version }}</span>
    </template>
  </footer>
</template>

<style scoped>
.statusbar {
  height: var(--statusbar-height);
  font-size: 12px;
  background: var(--tblr-bg-surface-secondary);
  border-top: var(--tblr-border-width) solid var(--tblr-border-color);
}
.status { height: auto; padding: 0; background: transparent; font-size: inherit; color: inherit; }
.sep { opacity: 0.4; }
.tnum { font-variant-numeric: tabular-nums; }
</style>
