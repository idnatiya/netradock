<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { IconBox, IconDatabase, IconLayoutDashboard, IconLogout, IconMoon, IconNetwork, IconStack2, IconSun } from '@tabler/icons-vue'
import { api, currentUser, notice } from '@/api'
import { theme, toggleTheme } from '@/theme'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

const route = useRoute()
const router = useRouter()
const menuOpen = ref(false)

const links = [
  { to: '/', label: 'Ringkasan', icon: IconLayoutDashboard },
  { to: '/containers', label: 'Container', icon: IconBox },
  { to: '/images', label: 'Image', icon: IconStack2 },
  { to: '/volumes', label: 'Volume', icon: IconDatabase },
  { to: '/networks', label: 'Network', icon: IconNetwork },
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
</script>

<template>
  <RouterView v-if="route.meta.guest" />
  <div v-else class="page">
    <header class="navbar navbar-expand-md d-print-none">
      <div class="container-xl">
        <button
          class="navbar-toggler"
          type="button"
          aria-controls="navbar-menu"
          :aria-expanded="menuOpen"
          aria-label="Buka menu"
          @click="menuOpen = !menuOpen"
        >
          <span class="navbar-toggler-icon"></span>
        </button>
        <RouterLink to="/" class="navbar-brand pe-0 pe-md-3">Netradock</RouterLink>
        <div class="navbar-nav flex-row order-md-last align-items-center gap-2">
          <button
            type="button"
            class="btn btn-ghost-secondary btn-icon"
            :aria-label="theme === 'dark' ? 'Pakai tema terang' : 'Pakai tema gelap'"
            :title="theme === 'dark' ? 'Tema terang' : 'Tema gelap'"
            @click="toggleTheme"
          >
            <IconSun v-if="theme === 'dark'" :size="20" />
            <IconMoon v-else :size="20" />
          </button>
          <span class="d-none d-sm-inline text-secondary">{{ currentUser }}</span>
          <button type="button" class="btn btn-ghost-secondary" @click="logout">
            <IconLogout :size="20" class="icon" />
            <span class="d-none d-sm-inline">Keluar</span>
            <span class="visually-hidden d-sm-none">Keluar</span>
          </button>
        </div>
      </div>
    </header>
    <header class="navbar-expand-md">
      <div id="navbar-menu" class="collapse navbar-collapse" :class="{ show: menuOpen }">
        <div class="navbar">
          <div class="container-xl">
            <ul class="navbar-nav">
              <li v-for="l in links" :key="l.to" class="nav-item" :class="{ active: isActive(l.to) }">
                <RouterLink class="nav-link" :to="l.to" :aria-current="isActive(l.to) ? 'page' : undefined">
                  <span class="nav-link-icon d-md-none d-lg-inline-block"><component :is="l.icon" :size="20" /></span>
                  <span class="nav-link-title">{{ l.label }}</span>
                </RouterLink>
              </li>
            </ul>
          </div>
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
  <ConfirmDialog />
</template>
