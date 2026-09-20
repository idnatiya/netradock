<script setup lang="ts">
import { computed, onUnmounted, ref, watch, type Component } from 'vue'
import { useRouter } from 'vue-router'
import { IconBox, IconDatabase, IconNetwork, IconSearch, IconStack2 } from '@tabler/icons-vue'
import { api, type Container, type Image, type Network, type Volume } from '@/api'
import { shortId } from '@/format'

type Hit = { key: string; label: string; sub: string; kind: string; icon: Component; to: string }

const PER_KIND = 5
const TTL = 30_000

const router = useRouter()
const root = ref<HTMLElement>()
const input = ref<HTMLInputElement>()
const query = ref('')
const open = ref(false)
const active = ref(0)
const pool = ref<Hit[]>([])
let loadedAt = 0

// Search is a convenience: a failed load just leaves the pool stale, no error banner.
async function load() {
  if (Date.now() - loadedAt < TTL) return
  try {
    const [containers, images, volumes, networks] = await Promise.all([
      api<Container[]>('GET', '/containers?all=true'),
      api<Image[]>('GET', '/images'),
      api<Volume[]>('GET', '/volumes'),
      api<Network[]>('GET', '/networks'),
    ])
    pool.value = [
      ...containers.map((c) => ({ key: `c-${c.id}`, label: c.name, sub: c.image, kind: 'Container', icon: IconBox, to: `/containers/${c.id}` })),
      ...images.map((i) => ({ key: `i-${i.id}`, label: i.tags[0] ?? shortId(i.id), sub: shortId(i.id), kind: 'Image', icon: IconStack2, to: '/images' })),
      ...volumes.map((v) => ({ key: `v-${v.name}`, label: v.name, sub: v.driver, kind: 'Volume', icon: IconDatabase, to: '/volumes' })),
      ...networks.map((n) => ({ key: `n-${n.id}`, label: n.name, sub: n.driver, kind: 'Network', icon: IconNetwork, to: '/networks' })),
    ]
    loadedAt = Date.now()
  } catch {
    // Keep whatever was loaded before.
  }
}

const results = computed(() => {
  const q = query.value.trim().toLowerCase()
  if (!q) return []
  const perKind = new Map<string, number>()
  const out: Hit[] = []
  for (const h of pool.value) {
    if (!h.label.toLowerCase().includes(q) && !h.sub.toLowerCase().includes(q)) continue
    const n = perKind.get(h.kind) ?? 0
    if (n >= PER_KIND) continue
    perKind.set(h.kind, n + 1)
    out.push(h)
  }
  return out
})

watch(results, () => (active.value = 0))

function onDocClick(e: MouseEvent) {
  if (!root.value?.contains(e.target as Node)) close()
}
function show() {
  open.value = true
  load()
  document.addEventListener('click', onDocClick)
}
function close() {
  open.value = false
  document.removeEventListener('click', onDocClick)
}
function go(h: Hit) {
  close()
  query.value = ''
  input.value?.blur()
  router.push(h.to)
}
function onKey(e: KeyboardEvent) {
  if (e.key === 'Escape') return close()
  if (!results.value.length) return
  if (e.key === 'ArrowDown') {
    e.preventDefault()
    active.value = (active.value + 1) % results.value.length
  } else if (e.key === 'ArrowUp') {
    e.preventDefault()
    active.value = (active.value - 1 + results.value.length) % results.value.length
  } else if (e.key === 'Enter') {
    e.preventDefault()
    go(results.value[active.value]!)
  }
}

// Cmd/Ctrl+K from anywhere focuses the box, like Docker Desktop.
function onHotkey(e: KeyboardEvent) {
  if ((e.metaKey || e.ctrlKey) && e.key.toLowerCase() === 'k') {
    e.preventDefault()
    input.value?.focus()
    input.value?.select()
  }
}
window.addEventListener('keydown', onHotkey)
onUnmounted(() => {
  window.removeEventListener('keydown', onHotkey)
  close()
})
</script>

<template>
  <div ref="root" class="search-container position-relative">
    <div class="search-box">
      <span class="search-icon"><IconSearch :size="15" /></span>
      <input
        ref="input"
        v-model="query"
        type="search"
        class="search-input"
        placeholder="Cari container, image, volume, network..."
        aria-label="Cari semua sumber daya"
        role="combobox"
        aria-controls="search-results"
        :aria-expanded="open && results.length > 0"
        :aria-activedescendant="open && results.length ? `search-hit-${active}` : undefined"
        @focus="show"
        @input="show"
        @keydown="onKey"
      >
      <span class="search-shortcut d-none d-sm-flex"><kbd class="kbd">⌘K</kbd></span>
    </div>

    <div v-if="open && query.trim()" id="search-results" class="dropdown-menu search-dropdown show w-100 py-1" role="listbox">
      <p v-if="!results.length" class="dropdown-header mb-0 text-secondary">Tidak ada yang cocok.</p>
      <button
        v-for="(h, i) in results"
        :id="`search-hit-${i}`"
        :key="h.key"
        type="button"
        role="option"
        :aria-selected="i === active"
        class="dropdown-item d-flex align-items-center gap-2"
        :class="{ active: i === active }"
        @click="go(h)"
        @mousemove="active = i"
      >
        <component :is="h.icon" :size="16" class="flex-shrink-0 text-secondary" />
        <span class="min-w-0 flex-grow-1">
          <span class="d-block text-truncate">{{ h.label }}</span>
          <span class="d-block small text-secondary font-monospace text-truncate">{{ h.sub }}</span>
        </span>
        <span class="badge bg-secondary-lt flex-shrink-0">{{ h.kind }}</span>
      </button>
    </div>
  </div>
</template>

<style scoped>
.search-container {
  max-width: 460px;
  width: 100%;
}
.search-box {
  position: relative;
  display: flex;
  align-items: center;
  width: 100%;
}
.search-icon {
  position: absolute;
  left: 0.625rem;
  display: flex;
  align-items: center;
  pointer-events: none;
  color: rgba(255, 255, 255, 0.7);
  z-index: 2;
}
.search-input {
  width: 100%;
  height: 32px;
  font-size: 13px;
  padding: 0 2.5rem 0 2rem;
  border-radius: var(--tblr-border-radius);
  background: rgba(0, 0, 0, 0.18);
  border: 1px solid rgba(255, 255, 255, 0.22);
  color: #ffffff;
  outline: none;
  transition: all 0.15s ease;
}
.search-input::placeholder {
  color: rgba(255, 255, 255, 0.65);
}
.search-input:hover {
  background: rgba(0, 0, 0, 0.24);
  border-color: rgba(255, 255, 255, 0.35);
}
.search-input:focus {
  background: rgba(0, 0, 0, 0.28);
  border-color: rgba(255, 255, 255, 0.75);
  box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.2);
}
.search-shortcut {
  position: absolute;
  right: 0.45rem;
  pointer-events: none;
  z-index: 2;
}
.kbd {
  font-family: inherit;
  font-size: 11px;
  font-weight: 500;
  padding: 1px 5px;
  border-radius: 4px;
  background: rgba(255, 255, 255, 0.16);
  border: 1px solid rgba(255, 255, 255, 0.12);
  color: rgba(255, 255, 255, 0.9);
}
.search-dropdown {
  position: absolute;
  top: calc(100% + 5px);
  left: 0;
  z-index: 1050;
  max-height: 60vh;
  overflow-y: auto;
  border-radius: 8px;
  border: 1px solid var(--tblr-border-color);
  box-shadow: 0 10px 25px rgba(0, 0, 0, 0.25);
  background: var(--tblr-bg-surface);
  color: var(--tblr-body-color);
}
</style>
