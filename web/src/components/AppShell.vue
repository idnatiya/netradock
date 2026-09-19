<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { IconBox, IconDatabase, IconLayoutDashboard, IconLogout, IconMoon, IconNetwork, IconStack2, IconSun } from '@tabler/icons-vue'
import { api, currentUser, notice, type System } from '@/api'
import { theme, toggleTheme } from '@/theme'
import { useLoad } from '@/useLoad'
import BrandLogo from '@/components/BrandLogo.vue'
import Dropdown from '@/components/Dropdown.vue'

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
  <aside class="navbar navbar-vertical navbar-expand-lg" data-bs-theme="dark">
      <div class="container-fluid">
        <button
          class="navbar-toggler"
          type="button"
          aria-controls="sidebar-menu"
          :aria-expanded="menuOpen"
          aria-label="Buka menu"
          @click="menuOpen = !menuOpen"
        >
          <span class="navbar-toggler-icon"></span>
        </button>
        <div class="navbar-brand navbar-brand-autodark">
          <RouterLink to="/" class="text-reset text-decoration-none"><BrandLogo /></RouterLink>
        </div>
        <div class="navbar-nav flex-row d-lg-none">
          <button type="button" class="btn btn-ghost-secondary btn-icon" :aria-label="theme === 'dark' ? 'Pakai tema terang' : 'Pakai tema gelap'" @click="toggleTheme">
            <IconSun v-if="theme === 'dark'" :size="20" />
            <IconMoon v-else :size="20" />
          </button>
        </div>
        <div id="sidebar-menu" class="collapse navbar-collapse" :class="{ show: menuOpen }">
          <ul class="navbar-nav pt-lg-3">
            <li v-for="l in links" :key="l.to" class="nav-item" :class="{ active: isActive(l.to) }">
              <RouterLink class="nav-link" :to="l.to" :aria-current="isActive(l.to) ? 'page' : undefined">
                <span class="nav-link-icon d-inline-block"><component :is="l.icon" :size="20" /></span>
                <span class="nav-link-title">{{ l.label }}</span>
                <span v-if="l.count() !== undefined" class="badge badge-sm bg-secondary-lt ms-auto">{{ l.count() }}</span>
              </RouterLink>
            </li>
          </ul>
          <div v-if="sys.data.value" class="host mt-auto px-3 py-3 small">
            <div class="text-uppercase fw-bold opacity-50 mb-1">Host</div>
            <div class="text-truncate" :title="sys.data.value.name">{{ sys.data.value.name }}</div>
            <div class="opacity-75">Docker {{ sys.data.value.server_version }} · {{ sys.data.value.architecture }}</div>
          </div>
          <div class="d-lg-none px-3 pb-3">
            <button type="button" class="btn btn-outline-light w-100" @click="logout"><IconLogout :size="18" class="icon" />Keluar ({{ currentUser }})</button>
          </div>
        </div>
      </div>
  </aside>

  <div class="page">
    <header class="navbar navbar-expand-md d-none d-lg-flex d-print-none">
      <div class="container-xl">
        <div class="navbar-nav flex-row order-md-last ms-auto align-items-center gap-2">
          <button type="button" class="btn btn-ghost-secondary btn-icon" :aria-label="theme === 'dark' ? 'Pakai tema terang' : 'Pakai tema gelap'" :title="theme === 'dark' ? 'Tema terang' : 'Tema gelap'" @click="toggleTheme">
            <IconSun v-if="theme === 'dark'" :size="20" />
            <IconMoon v-else :size="20" />
          </button>
          <Dropdown label="Menu akun" end class="nav-link d-flex lh-1 text-reset p-0 border-0 bg-transparent">
            <template #toggle>
              <span class="avatar avatar-sm bg-primary-lt">{{ initials() }}</span>
              <span class="ps-2 text-start">
                <span class="d-block">{{ currentUser }}</span>
                <span class="d-block mt-1 small text-secondary">Administrator</span>
              </span>
            </template>
            <button type="button" class="dropdown-item" role="menuitem" @click="logout"><IconLogout :size="18" class="icon dropdown-item-icon" />Keluar</button>
          </Dropdown>
        </div>
      </div>
    </header>

    <div class="page-wrapper">
      <div v-if="notice" class="container-xl pt-3">
        <div class="alert alert-dismissible mb-0" :class="notice.kind === 'ok' ? 'alert-success' : 'alert-danger'" role="status">
          <div class="text-break-all">{{ notice.text }}</div>
          <button type="button" class="btn-close" aria-label="Tutup pesan" @click="notice = null"></button>
        </div>
      </div>
      <RouterView />
    </div>
  </div>
</template>

<style scoped>
.navbar-vertical .navbar-collapse { flex-direction: column; align-items: stretch; }
.nav-link { width: 100%; }
/* Tabler floats .nav-link .badge like a notification bubble; here it is an inline count. */
.navbar .navbar-nav .nav-link .badge { position: static; transform: none; }
.host { border-top: 1px solid var(--tblr-border-color-translucent); color: var(--tblr-navbar-color); }
</style>
