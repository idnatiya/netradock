<script setup lang="ts">
import { AlertTriangle, Loader2, RotateCcw } from 'lucide-vue-next'
import { Button } from '@/components/ui/button'

defineProps<{ loading: boolean; error: string; empty: boolean; what: string }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <div v-if="loading" class="flex items-center justify-center gap-2.5 p-8 text-sm text-muted-foreground" role="status">
    <Loader2 class="size-4 animate-spin text-primary" />
    <span>Loading {{ what }}...</span>
  </div>

  <div v-else-if="error && empty" class="m-4 p-5 rounded-xl border border-destructive/30 bg-destructive/10 text-destructive text-sm" role="alert">
    <div class="flex items-start gap-3">
      <AlertTriangle class="size-5 shrink-0 mt-0.5" />
      <div class="space-y-1 flex-1">
        <h4 class="font-semibold text-foreground">Cannot load {{ what }}</h4>
        <p class="text-xs text-muted-foreground break-all">{{ error }}. Ensure the Docker socket is mounted properly.</p>
        <div class="pt-2">
          <Button variant="destructive" size="sm" class="gap-1.5" type="button" @click="$emit('retry')">
            <RotateCcw class="size-3.5" />
            Retry
          </Button>
        </div>
      </div>
    </div>
  </div>

  <div v-else-if="empty" class="p-8 text-center text-muted-foreground">
    <slot name="empty" />
  </div>

  <template v-else>
    <div v-if="error" class="m-4 p-3 rounded-lg border border-amber-500/30 bg-amber-500/10 text-amber-600 dark:text-amber-400 text-xs" role="alert">
      Data might be outdated: {{ error }}
    </div>
    <slot />
  </template>
</template>
