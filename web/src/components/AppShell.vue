<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { IconBox, IconDatabase, IconLayoutDashboard, IconLogout, IconMenu2, IconMoon, IconNetwork, IconStack2, IconSun } from '@tabler/icons-vue'
import { api, currentUser, notice, type System } from '@/api'
import { theme, toggleTheme } from '@/theme'
import { useLoad } from '@/useLoad'
import BrandLogo from '@/components/BrandLogo.vue'
import Dropdown from '@/components/Dropdown.vue'
import GlobalSearch from '@/components/GlobalSearch.vue'
import StatusBar from '@/components/StatusBar.vue'

const route = useRoute()
const router = useRouter()
const menuOpen = ref(false)
const sys = useLoad(() => api<System>('GET', '/system'), 10000)

const links = [
  { to: '/', label: 'Ringkasan', icon: IconLayoutDashboard, count: () => undefined },
  { to: '/containers', label: 'Container', icon: IconBox, count: () => sys.data.value && `${sys.data.value.containers_running}/${sys.data.value.containers}` },
  { to: '/images', label: 'Image', icon: IconStack2, count: () => sys.data.value?.images },
  { to: '/volumes', label: 'Volume', icon: IconDatabase, count: () => undefined },
  { to: '/networks', label: 'Network', icon: IconNetwork, count: () => undefined },
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
  <div class="shell">
    <div class="shell-brand">
      <button
        class="btn btn-icon btn-ghost-secondary d-lg-none"
        type="button"
        aria-controls="sidebar-menu"
        :aria-expanded="menuOpen"
        aria-label="Buka menu"
        @click="menuOpen = !menuOpen"
      >
        <IconMenu2 :size="20" />
      </button>
      <RouterLink to="/" class="text-reset text-decoration-none"><BrandLogo /></RouterLink>
    </div>

    <header class="shell-top">
      <GlobalSearch />
      <div class="d-flex align-items-center gap-2 ms-auto">
        <button type="button" class="btn btn-ghost-secondary btn-icon" :aria-label="theme === 'dark' ? 'Pakai tema terang' : 'Pakai tema gelap'" :title="theme === 'dark' ? 'Tema terang' : 'Tema gelap'" @click="toggleTheme">
          <IconSun v-if="theme === 'dark'" :size="20" />
          <IconMoon v-else :size="20" />
        </button>
        <Dropdown label="Menu akun" end class="nav-link d-flex lh-1 text-reset p-0 border-0 bg-transparent">
          <template #toggle>
            <span class="avatar avatar-sm bg-primary-lt">{{ initials() }}</span>
            <span class="ps-2 text-start d-none d-md-block">
              <span class="d-block">{{ currentUser }}</span>
              <span class="d-block mt-1 small text-secondary">Administrator</span>
            </span>
          </template>
          <button type="button" class="dropdown-item" role="menuitem" @click="logout"><IconLogout :size="18" class="icon dropdown-item-icon" />Keluar</button>
        </Dropdown>
      </div>
    </header>

    <aside id="sidebar-menu" class="shell-nav" :class="{ open: menuOpen }">
      <ul class="nav-list">
        <li v-for="l in links" :key="l.to">
          <RouterLink class="nav-item" :class="{ active: isActive(l.to) }" :to="l.to" :aria-current="isActive(l.to) ? 'page' : undefined">
            <component :is="l.icon" :size="18" class="flex-shrink-0" />
            <span class="flex-grow-1">{{ l.label }}</span>
            <span v-if="l.count() !== undefined" class="badge bg-secondary-lt">{{ l.count() }}</span>
          </RouterLink>
        </li>
      </ul>
      <button type="button" class="btn w-100 mt-3 d-lg-none" @click="logout"><IconLogout :size="18" class="icon" />Keluar ({{ currentUser }})</button>
    </aside>
    <div v-if="menuOpen" class="shell-scrim d-lg-none" @click="menuOpen = false"></div>

    <main class="shell-main">
      <div class="page-wrapper">
        <div v-if="notice" class="container-xl pt-3">
          <div class="alert alert-dismissible mb-0" :class="notice.kind === 'ok' ? 'alert-success' : 'alert-danger'" role="status">
            <div class="text-break-all">{{ notice.text }}</div>
            <button type="button" class="btn-close" aria-label="Tutup pesan" @click="notice = null"></button>
          </div>
        </div>
        <RouterView />
      </div>
    </main>

    <StatusBar class="shell-status" :sys="sys.data.value ?? null" :error="sys.error.value" />
  </div>
</template>

<style scoped>
.shell {
  display: grid;
  grid-template-columns: var(--nav-width) 1fr;
  grid-template-rows: var(--topbar-height) 1fr var(--statusbar-height);
  grid-template-areas: 'brand top' 'nav main' 'status status';
  height: 100dvh;
}
.shell-brand {
  grid-area: brand;
  display: flex;
  align-items: center;
  gap: 0.5rem;
  padding: 0 0.75rem;
  background: var(--tblr-bg-surface-secondary);
  border-right: var(--tblr-border-width) solid var(--tblr-border-color);
  border-bottom: var(--tblr-border-width) solid var(--tblr-border-color);
}
.shell-top {
  grid-area: top;
  display: flex;
  align-items: center;
  gap: 0.75rem;
  padding: 0 0.75rem;
  background: var(--tblr-body-bg);
  border-bottom: var(--tblr-border-width) solid var(--tblr-border-color);
}
.shell-nav {
  grid-area: nav;
  padding: 0.5rem;
  overflow-y: auto;
  background: var(--tblr-bg-surface-secondary);
  border-right: var(--tblr-border-width) solid var(--tblr-border-color);
}
.shell-main { grid-area: main; overflow: auto; background: var(--tblr-body-bg); }
.shell-status { grid-area: status; }

.nav-list { list-style: none; margin: 0; padding: 0; }
.nav-item {
  display: flex;
  align-items: center;
  gap: 0.625rem;
  padding: 0.5rem 0.625rem;
  margin-bottom: 2px;
  border-radius: var(--tblr-border-radius);
  color: var(--tblr-body-color);
  text-decoration: none;
}
.nav-item:hover { background: var(--tblr-border-color-translucent); }
.nav-item.active { background: var(--tblr-primary-lt); color: var(--tblr-primary); font-weight: 500; }

@media (max-width: 991.98px) {
  .shell {
    grid-template-columns: 1fr;
    grid-template-rows: var(--topbar-height) auto 1fr var(--statusbar-height);
    grid-template-areas: 'brand' 'top' 'main' 'status';
  }
  .shell-brand { border-right: 0; }
  .shell-nav {
    position: fixed;
    top: var(--topbar-height);
    bottom: var(--statusbar-height);
    left: 0;
    z-index: 1040;
    width: var(--nav-width);
    transform: translateX(-100%);
    transition: transform 0.15s ease-out;
  }
  .shell-nav.open { transform: none; }
  .shell-scrim { position: fixed; inset: 0; z-index: 1030; background: rgb(24 36 51 / 0.4); }
}
</style>
