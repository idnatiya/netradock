<script setup lang="ts">
defineProps<{ pretitle?: string; title: string }>()
defineSlots<{ meta?(): unknown; actions?(): unknown; avatar?(): unknown }>()
</script>

<template>
  <div class="border-b border-border bg-card/30 px-6 py-5">
    <div class="max-w-7xl mx-auto flex flex-col md:flex-row md:items-center justify-between gap-4">
      <div class="flex items-center gap-3 min-w-0">
        <slot name="avatar" />
        <div class="min-w-0">
          <div v-if="pretitle" class="text-[11px] uppercase tracking-wider font-semibold text-muted-foreground font-mono mb-0.5">
            {{ pretitle }}
          </div>
          <h1 class="text-2xl font-bold tracking-tight text-foreground truncate">
            {{ title }}
          </h1>
          <div v-if="$slots.meta" class="mt-1 text-xs text-muted-foreground">
            <slot name="meta" />
          </div>
        </div>
      </div>

      <div v-if="$slots.actions" class="flex flex-wrap items-center gap-2">
        <slot name="actions" />
      </div>
    </div>
  </div>
</template>
