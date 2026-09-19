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
  root.value?.querySelector<HTMLElement>('.dropdown-item')?.focus()
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
  <div ref="root" class="dropdown d-inline-block">
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
    <div v-if="open" class="dropdown-menu show" :class="{ 'dropdown-menu-end': end }" role="menu" @click="close(false)">
      <slot />
    </div>
  </div>
</template>

<style scoped>
.dropdown-menu { position: absolute; top: 100%; z-index: 1050; margin-top: 4px; }
.dropdown-menu-end { right: 0; left: auto; }
</style>
