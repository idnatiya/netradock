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
  <div ref="root" class="search position-relative">
    <div class="input-icon">
      <span class="input-icon-addon"><IconSearch :size="16" /></span>
      <input
        ref="input"
        v-model="query"
        type="search"
        class="form-control form-control-sm"
        placeholder="Cari container, image, volume, network"
        aria-label="Cari semua sumber daya"
        role="combobox"
        aria-controls="search-results"
        :aria-expanded="open && results.length > 0"
        :aria-activedescendant="open && results.length ? `search-hit-${active}` : undefined"
        @focus="show"
        @input="show"
        @keydown="onKey"
      >
      <span class="input-icon-addon end text-secondary d-none d-xl-flex"><kbd class="kbd">⌘K</kbd></span>
    </div>

    <div v-if="open && query.trim()" id="search-results" class="dropdown-menu show w-100 mt-1 py-1" role="listbox">
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
/* The box sits on Docker Desktop's blue title bar, so it draws its own light-on-blue skin. */
.search { max-width: 560px; width: 100%; }
.search input {
  padding-right: 3.5rem;
  background: rgb(0 0 0 / 0.22);
  border-color: rgb(255 255 255 / 0.28);
  color: #fff;
}
.search input::placeholder { color: rgb(255 255 255 / 0.65); }
.search input:focus {
  background: rgb(0 0 0 / 0.3);
  border-color: #fff;
  box-shadow: 0 0 0 2px rgb(255 255 255 / 0.25);
  color: #fff;
}
.search .input-icon-addon { color: rgb(255 255 255 / 0.7); }
.input-icon-addon.end { left: auto; right: 0; width: auto; padding-right: 0.5rem; pointer-events: none; }
.kbd { background: rgb(255 255 255 / 0.15); border: 0; color: #fff; }
.dropdown-menu { position: absolute; top: 100%; left: 0; z-index: 1050; max-height: 60vh; overflow-y: auto; }
</style>
