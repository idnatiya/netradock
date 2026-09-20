<script setup lang="ts">
import { computed, onUnmounted, ref, watch, type Component } from 'vue'
import { useRouter } from 'vue-router'
import { Boxes, HardDrive, Layers, Network, Search } from 'lucide-vue-next'
import { api, type Container, type Image, type Network as NetType, type Volume } from '@/api'
import { shortId } from '@/format'
import { Badge } from '@/components/ui/badge'

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
      api<NetType[]>('GET', '/networks'),
    ])
    pool.value = [
      ...containers.map((c) => ({ key: `c-${c.id}`, label: c.name, sub: c.image, kind: 'Container', icon: Boxes, to: `/containers/${c.id}` })),
      ...images.map((i) => ({ key: `i-${i.id}`, label: i.tags[0] ?? shortId(i.id), sub: shortId(i.id), kind: 'Image', icon: Layers, to: '/images' })),
      ...volumes.map((v) => ({ key: `v-${v.name}`, label: v.name, sub: v.driver, kind: 'Volume', icon: HardDrive, to: '/volumes' })),
      ...networks.map((n) => ({ key: `n-${n.id}`, label: n.name, sub: n.driver, kind: 'Network', icon: Network, to: '/networks' })),
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

// Cmd/Ctrl+K from anywhere focuses the box
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
  <div ref="root" class="relative w-full max-w-sm sm:max-w-md">
    <div class="relative flex items-center">
      <Search class="absolute left-2.5 size-4 text-muted-foreground pointer-events-none" />
      <input
        ref="input"
        v-model="query"
        type="search"
        class="h-8.5 w-full rounded-md border border-input bg-muted/40 pl-8 pr-12 text-xs text-foreground placeholder:text-muted-foreground transition-colors focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring focus-visible:bg-background"
        placeholder="Search containers, images, volumes..."
        aria-label="Quick search resources"
        role="combobox"
        aria-controls="search-results"
        :aria-expanded="open && results.length > 0"
        :aria-activedescendant="open && results.length ? `search-hit-${active}` : undefined"
        @focus="show"
        @input="show"
        @keydown="onKey"
      />
      <div class="absolute right-2 pointer-events-none hidden sm:flex items-center">
        <kbd class="inline-flex h-4.5 items-center gap-0.5 rounded border border-border bg-muted/60 px-1 font-mono text-[10px] font-medium text-muted-foreground">
          ⌘K
        </kbd>
      </div>
    </div>

    <!-- Dropdown results -->
    <div
      v-if="open && query.trim()"
      id="search-results"
      class="absolute top-full mt-1.5 left-0 z-50 w-full overflow-hidden rounded-md border border-border bg-popover p-1 text-popover-foreground shadow-md animate-in fade-in-0 zoom-in-95"
      role="listbox"
    >
      <div v-if="!results.length" class="p-3 text-center text-xs text-muted-foreground">
        No matching resources found.
      </div>
      <button
        v-for="(h, i) in results"
        :id="`search-hit-${i}`"
        :key="h.key"
        type="button"
        role="option"
        :aria-selected="i === active"
        class="flex w-full items-center gap-2.5 rounded-sm px-2.5 py-1.5 text-left text-xs transition-colors hover:bg-accent hover:text-accent-foreground cursor-pointer"
        :class="{ 'bg-accent text-accent-foreground': i === active }"
        @click="go(h)"
        @mousemove="active = i"
      >
        <component :is="h.icon" class="size-4 shrink-0 text-muted-foreground" />
        <div class="min-w-0 flex-1">
          <div class="font-medium text-foreground truncate">{{ h.label }}</div>
          <div class="font-mono text-[11px] text-muted-foreground truncate">{{ h.sub }}</div>
        </div>
        <Badge variant="outline" class="text-[10px] py-0 px-1.5 font-mono shrink-0">
          {{ h.kind }}
        </Badge>
      </button>
    </div>
  </div>
</template>
