<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { AlertTriangle } from 'lucide-vue-next'
import { pending } from '@/confirm'
import { Button } from '@/components/ui/button'

const dialog = ref<HTMLDialogElement>()
let trigger: HTMLElement | null = null

watch(pending, async (p) => {
  if (!p || dialog.value?.open) return
  trigger = document.activeElement as HTMLElement | null
  await nextTick()
  dialog.value?.showModal()
})

function close(ok: boolean) {
  pending.value?.resolve(ok)
  pending.value = null
  dialog.value?.close()
  trigger?.focus()
}
</script>

<template>
  <dialog
    ref="dialog"
    class="confirm backdrop:bg-black/60 backdrop:backdrop-blur-xs m-auto rounded-xl border border-border bg-card p-0 text-card-foreground shadow-xl focus:outline-none"
    aria-labelledby="confirm-title"
    @cancel.prevent="close(false)"
  >
    <div v-if="pending" class="p-6 space-y-4 max-w-sm">
      <div class="flex items-center gap-3">
        <div class="size-10 rounded-full bg-destructive/15 text-destructive flex items-center justify-center shrink-0">
          <AlertTriangle class="size-5" />
        </div>
        <div>
          <h3 id="confirm-title" class="text-base font-semibold text-foreground">
            {{ pending.title }}
          </h3>
        </div>
      </div>

      <div class="text-xs text-muted-foreground break-words leading-relaxed pl-13">
        {{ pending.body }}
      </div>

      <div class="flex items-center justify-end gap-2 pt-2 border-t border-border">
        <Button variant="outline" size="sm" type="button" autofocus @click="close(false)">
          Cancel
        </Button>
        <Button variant="destructive" size="sm" type="button" @click="close(true)">
          {{ pending.action }}
        </Button>
      </div>
    </div>
  </dialog>
</template>

<style scoped>
.confirm {
  width: min(420px, calc(100vw - 32px));
}
</style>
