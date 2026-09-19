<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
import { IconAlertTriangle } from '@tabler/icons-vue'
import { pending } from '@/confirm'

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
  <!-- Native dialog: focus trap and Escape come from the browser; Tabler supplies the modal look. -->
  <dialog ref="dialog" class="confirm" aria-labelledby="confirm-title" @cancel.prevent="close(false)">
    <div v-if="pending" class="modal-content">
      <div class="modal-status bg-danger"></div>
      <div class="modal-body text-center py-4">
        <IconAlertTriangle :size="40" class="text-danger mb-2" />
        <h3 id="confirm-title">{{ pending.title }}</h3>
        <div class="text-secondary text-break-all">{{ pending.body }}</div>
      </div>
      <div class="modal-footer">
        <div class="w-100">
          <div class="row g-2">
            <div class="col"><button class="btn w-100" type="button" autofocus @click="close(false)">Batal</button></div>
            <div class="col"><button class="btn btn-danger w-100" type="button" @click="close(true)">{{ pending.action }}</button></div>
          </div>
        </div>
      </div>
    </div>
  </dialog>
</template>

<style scoped>
.confirm {
  width: min(380px, calc(100vw - 32px));
  padding: 0;
  border: 0;
  border-radius: var(--tblr-border-radius-lg);
  background: transparent;
  color: var(--tblr-body-color);
}
.confirm .modal-content { position: relative; background: var(--tblr-bg-surface); }
.confirm::backdrop { background: rgb(24 36 51 / 0.5); }
</style>
