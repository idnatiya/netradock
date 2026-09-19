<script setup lang="ts">
import { ref, watch, nextTick } from 'vue'
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
  <!-- Native dialog: focus trap and Escape come from the browser. -->
  <dialog ref="dialog" aria-labelledby="confirm-title" @cancel.prevent="close(false)">
    <template v-if="pending">
      <h2 id="confirm-title">{{ pending.title }}</h2>
      <p>{{ pending.body }}</p>
      <div class="row">
        <button class="btn" type="button" autofocus @click="close(false)">Batal</button>
        <button class="btn btn-primary" type="button" @click="close(true)">{{ pending.action }}</button>
      </div>
    </template>
  </dialog>
</template>

<style scoped>
dialog {
  width: min(440px, calc(100vw - 32px));
  border: 0;
  border-radius: var(--r-lg);
  padding: 24px;
  color: var(--body);
  box-shadow: 0 12px 40px rgb(24 29 38 / 0.2); /* the only floating layer in the app */
}
dialog::backdrop { background: rgb(24 29 38 / 0.45); }
p { margin: 8px 0 24px; overflow-wrap: anywhere; }
.row { display: flex; justify-content: flex-end; gap: 8px; flex-wrap: wrap; }
</style>
