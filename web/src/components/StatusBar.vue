<script setup lang="ts">
import { computed } from 'vue'
import { Activity, Cpu, HardDrive, Server } from 'lucide-vue-next'
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
  <footer class="h-8 shrink-0 flex items-center justify-between border-t border-border bg-card/60 px-4 text-xs text-muted-foreground font-mono backdrop-blur-xs select-none">
    <div class="flex items-center gap-3">
      <div class="flex items-center gap-1.5">
        <span
          class="size-2 rounded-full"
          :class="up ? 'bg-emerald-500 shadow-[0_0_8px_rgba(16,185,129,0.5)]' : 'bg-rose-500'"
        />
        <span :class="up ? 'text-foreground font-medium' : 'text-rose-500'">
          {{ up ? 'Engine Connected' : 'Engine Offline' }}
        </span>
      </div>

      <template v-if="sys">
        <span class="text-border">|</span>
        <div class="hidden md:flex items-center gap-1.5">
          <Cpu class="size-3.5 text-muted-foreground" />
          <span>CPU {{ cpu.toFixed(1) }}%</span>
        </div>

        <span class="hidden md:inline text-border">|</span>
        <div class="hidden md:flex items-center gap-1.5">
          <Activity class="size-3.5 text-muted-foreground" />
          <span>RAM {{ bytes(mem) }} / {{ bytes(sys.mem_total) }}</span>
        </div>
      </template>
    </div>

    <div v-if="sys" class="flex items-center gap-3">
      <div class="hidden lg:flex items-center gap-1.5 text-muted-foreground truncate" :title="sys.name">
        <Server class="size-3.5" />
        <span class="truncate max-w-[200px]">{{ sys.name }}</span>
      </div>
      <span class="hidden lg:inline text-border">|</span>
      <span class="text-muted-foreground">Docker v{{ sys.server_version }}</span>
    </div>
  </footer>
</template>
