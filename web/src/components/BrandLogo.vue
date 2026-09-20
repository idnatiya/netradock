<script setup lang="ts">
import { Boxes } from 'lucide-vue-next'
</script>

<template>
  <div class="flex items-center gap-2.5 select-none">
    <div class="flex size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground shadow-xs">
      <Boxes class="size-4.5" :stroke-width="2.2" />
    </div>
    <div class="flex flex-col">
      <span class="text-sm font-semibold tracking-tight text-foreground leading-none">Netradock</span>
      <span class="text-[10px] text-muted-foreground font-mono tracking-wider leading-tight">CONTAINERS</span>
    </div>
  </div>
</template>
