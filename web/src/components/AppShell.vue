<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import {
  Boxes,
  CheckCircle2,
  HardDrive,
  LayoutDashboard,
  Layers,
  LogOut,
  Menu,
  Moon,
  Network,
  Sun,
  X,
  XCircle,
} from 'lucide-vue-next'
import { api, currentUser, notice, type System } from '@/api'
import { useLiveStats } from '@/liveStats'
import { theme, toggleTheme } from '@/theme'
import { useLoad } from '@/useLoad'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import BrandLogo from '@/components/BrandLogo.vue'
import GlobalSearch from '@/components/GlobalSearch.vue'
import StatusBar from '@/components/StatusBar.vue'

const route = useRoute()
const router = useRouter()
const menuOpen = ref(false)
const sys = useLoad(() => api<System>('GET', '/system'), 10000)
const live = useLiveStats()

const ncpu = computed(() => sys.data.value?.ncpu || 1)
const memTotal = computed(() => sys.data.value?.mem_total || 1)
const cpuPct = computed(() => {
  const latestCpu = live.total.cpu.at(-1) ?? 0
  return Math.min(100, latestCpu / ncpu.value)
})
const ramPct = computed(() => {
  const latestMem = live.total.mem.at(-1) ?? 0
  return Math.min(100, (latestMem / memTotal.value) * 100)
})

interface NavItem {
  to: string
  label: string
  icon: any
  badge?: () => string | undefined
}

interface NavGroup {
  title?: string
  items: NavItem[]
}

const navGroups: NavGroup[] = [
  {
    title: 'Overview',
    items: [
      { to: '/', label: 'Dashboard', icon: LayoutDashboard },
    ],
  },
  {
    title: 'Resources',
    items: [
      {
        to: '/containers',
        label: 'Containers',
        icon: Boxes,
        badge: () => sys.data.value?.containers_running !== undefined ? `${sys.data.value.containers_running}` : undefined,
      },
      {
        to: '/images',
        label: 'Images',
        icon: Layers,
        badge: () => sys.data.value?.images !== undefined ? `${sys.data.value.images}` : undefined,
      },
      { to: '/volumes', label: 'Volumes', icon: HardDrive },
      { to: '/networks', label: 'Networks', icon: Network },
    ],
  },
]

function isActive(to: string) {
  return to === '/' ? route.path === '/' : route.path.startsWith(to)
}

watch(() => route.fullPath, () => {
  menuOpen.value = false
  notice.value = null
})

async function logout() {
  await api('POST', '/auth/logout')
  currentUser.value = null
  router.push({ name: 'login' })
}

const initials = () => (currentUser.value ?? '?').slice(0, 2).toUpperCase()
</script>

<template>
  <div class="flex h-dvh flex-col bg-background text-foreground overflow-hidden">
    <!-- Top Header -->
    <header class="h-14 shrink-0 border-b border-sidebar bg-sidebar/80 px-4 flex items-center justify-between gap-4 backdrop-blur-md z-30 select-none">
      <div class="flex items-center gap-3">
        <Button
          variant="ghost"
          size="icon-sm"
          class="lg:hidden"
          aria-label="Toggle navigation menu"
          @click="menuOpen = !menuOpen"
        >
          <Menu class="size-4" />
        </Button>
        <RouterLink to="/" class="flex items-center text-foreground hover:opacity-90 transition-opacity">
          <BrandLogo />
        </RouterLink>
      </div>

      <!-- Search in center -->
      <div class="flex-1 max-w-md mx-auto hidden sm:flex justify-center">
        <GlobalSearch />
      </div>

      <!-- Right actions -->
      <div class="flex items-center gap-2">
        <div v-if="sys.data.value" class="hidden md:flex items-center gap-2 px-2.5 py-1 rounded-full border border-border bg-muted/40 text-xs font-mono">
          <span class="size-2 rounded-full bg-emerald-500 animate-pulse-dot" />
          <span class="text-muted-foreground">Engine v{{ sys.data.value.server_version }}</span>
        </div>

        <Button
          variant="ghost"
          size="icon-sm"
          :title="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
          @click="toggleTheme"
        >
          <Sun v-if="theme === 'dark'" class="size-4 text-muted-foreground hover:text-foreground" />
          <Moon v-else class="size-4 text-muted-foreground hover:text-foreground" />
        </Button>

        <div class="flex items-center gap-2 pl-2 border-l border-border">
          <div class="size-7 rounded-full bg-primary/10 border border-primary/20 text-primary flex items-center justify-center font-mono text-xs font-semibold">
            {{ initials() }}
          </div>
          <span class="text-xs font-medium hidden lg:inline max-w-[100px] truncate text-muted-foreground">{{ currentUser }}</span>
          <Button
            variant="ghost"
            size="icon-sm"
            title="Log out"
            @click="logout"
          >
            <LogOut class="size-4 text-muted-foreground hover:text-destructive" />
          </Button>
        </div>
      </div>
    </header>

    <!-- Main Workspace (Sidebar + Content View) -->
    <div class="flex flex-1 min-h-0 relative">
      <!-- Mobile backdrop -->
      <div
        v-if="menuOpen"
        class="fixed inset-0 z-40 bg-background/80 backdrop-blur-xs lg:hidden"
        @click="menuOpen = false"
      />

      <!-- Sidebar (Level 1: Sunken Navigation Rail) -->
      <aside
        class="fixed inset-y-0 left-0 z-50 w-60 border-r border-sidebar bg-sidebar flex flex-col transition-transform duration-200 lg:static lg:translate-x-0"
        :class="menuOpen ? 'translate-x-0' : '-translate-x-full'"
      >
        <!-- Mobile close button -->
        <div class="flex lg:hidden items-center justify-between h-14 px-4 border-b border-sidebar">
          <BrandLogo />
          <Button variant="ghost" size="icon-sm" @click="menuOpen = false">
            <X class="size-4" />
          </Button>
        </div>

        <!-- Navigation items -->
        <div class="flex-1 overflow-y-auto px-3 py-4 space-y-6">
          <div v-for="g in navGroups" :key="g.title || 'main'" class="space-y-1">
            <h4 v-if="g.title" class="px-2 text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {{ g.title }}
            </h4>
            <div class="space-y-0.5 pt-1">
              <RouterLink
                v-for="l in g.items"
                :key="l.to"
                :to="l.to"
                class="flex items-center justify-between gap-3 px-2.5 py-2 rounded-md text-sm font-medium transition-colors"
                :class="
                  isActive(l.to)
                    ? 'bg-accent text-accent-foreground font-semibold shadow-2xs'
                    : 'text-muted-foreground hover:bg-muted/50 hover:text-foreground'
                "
              >
                <div class="flex items-center gap-2.5 min-w-0">
                  <component :is="l.icon" class="size-4 shrink-0" />
                  <span class="truncate">{{ l.label }}</span>
                </div>
                <Badge
                  v-if="l.badge && l.badge() !== undefined"
                  variant="secondary"
                  class="font-mono text-xs py-0.5 px-2"
                >
                  {{ l.badge() }}
                </Badge>
              </RouterLink>
            </div>
          </div>
        </div>

        <!-- Sidebar Telemetry Footer -->
        <div v-if="sys.data.value" class="p-3 border-t border-sidebar bg-sidebar/50 space-y-3">
          <div class="space-y-1.5">
            <div class="flex items-center justify-between text-xs font-mono text-muted-foreground">
              <span>CPU Host</span>
              <span class="font-medium text-foreground">{{ cpuPct.toFixed(1) }}%</span>
            </div>
            <div class="h-1.5 w-full rounded-full bg-muted overflow-hidden">
              <div
                class="h-full rounded-full bg-primary transition-all duration-300"
                :style="{ width: `${Math.min(100, Math.max(2, cpuPct))}%` }"
              />
            </div>
          </div>

          <div class="space-y-1.5">
            <div class="flex items-center justify-between text-xs font-mono text-muted-foreground">
              <span>RAM Host</span>
              <span class="font-medium text-foreground">{{ ramPct.toFixed(1) }}%</span>
            </div>
            <div class="h-1.5 w-full rounded-full bg-muted overflow-hidden">
              <div
                class="h-full rounded-full bg-sky-500 transition-all duration-300"
                :style="{ width: `${Math.min(100, Math.max(2, ramPct))}%` }"
              />
            </div>
          </div>
        </div>
      </aside>

      <!-- Main Scrollable Content Area -->
      <main class="flex-1 min-w-0 overflow-y-auto">
        <!-- Toast / Notice alert -->
        <div v-if="notice" class="p-4 pb-0 w-full">
          <div
            class="flex items-center justify-between gap-3 p-3 rounded-lg border text-xs"
            :class="
              notice.kind === 'ok'
                ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-600 dark:text-emerald-400'
                : 'border-rose-500/20 bg-rose-500/10 text-rose-600 dark:text-rose-400'
            "
          >
            <div class="flex items-center gap-2 truncate">
              <CheckCircle2 v-if="notice.kind === 'ok'" class="size-4 shrink-0" />
              <XCircle v-else class="size-4 shrink-0" />
              <span class="truncate">{{ notice.text }}</span>
            </div>
            <Button variant="ghost" size="icon-sm" class="size-5" @click="notice = null">
              <X class="size-3" />
            </Button>
          </div>
        </div>

        <RouterView />
      </main>
    </div>

    <!-- Bottom Status Bar -->
    <StatusBar :sys="sys.data.value ?? null" :error="sys.error.value" />
  </div>
</template>
