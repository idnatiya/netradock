<script setup lang="ts">
import { nextTick, onUnmounted, ref } from 'vue'

defineOptions({ inheritAttrs: false })
defineProps<{ label: string; end?: boolean }>()

const open = ref(false)
const root = ref<HTMLElement>()
const toggle = ref<HTMLButtonElement>()

function onDocClick(e: MouseEvent) {
  if (!root.value?.contains(e.target as Node)) close(false)
}
function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') close(true)
}
async function show() {
  open.value = true
  document.addEventListener('click', onDocClick)
  document.addEventListener('keydown', onKey)
  await nextTick()
  root.value?.querySelector<HTMLElement>('[role="menuitem"]')?.focus()
}
function close(refocus: boolean) {
  open.value = false
  document.removeEventListener('click', onDocClick)
  document.removeEventListener('keydown', onKey)
  if (refocus) toggle.value?.focus()
}
onUnmounted(() => close(false))
</script>

<template>
  <div ref="root" class="relative inline-block text-left">
    <button
      ref="toggle"
      type="button"
      :aria-label="label"
      :title="label"
      aria-haspopup="menu"
      :aria-expanded="open"
      v-bind="$attrs"
      @click.stop="open ? close(false) : show()"
    >
      <slot name="toggle" />
    </button>
    <div
      v-if="open"
      class="absolute mt-1.5 z-50 min-w-[8rem] overflow-hidden rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95"
      :class="end ? 'right-0' : 'left-0'"
      role="menu"
      @click="close(false)"
    >
      <slot />
    </div>
  </div>
</template>
