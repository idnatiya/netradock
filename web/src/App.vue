<script setup lang="ts">
import { ref, watch } from 'vue'
import { RouterLink, RouterView, useRoute, useRouter } from 'vue-router'
import { api, currentUser, notice } from '@/api'
import ConfirmDialog from '@/components/ConfirmDialog.vue'

const route = useRoute()
const router = useRouter()
const menuOpen = ref(false)

const links = [
  { to: '/', label: 'Ringkasan' },
  { to: '/containers', label: 'Container' },
  { to: '/images', label: 'Image' },
  { to: '/volumes', label: 'Volume' },
  { to: '/networks', label: 'Network' },
]

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
  <header v-if="!route.meta.guest" class="nav">
    <RouterLink to="/" class="wordmark">Netradock</RouterLink>
    <button
      class="btn menu-toggle"
      type="button"
      :aria-expanded="menuOpen"
      aria-controls="nav-links"
      @click="menuOpen = !menuOpen"
    >
      {{ menuOpen ? 'Tutup' : 'Menu' }}
    </button>
    <nav id="nav-links" :class="{ open: menuOpen }" aria-label="Utama">
      <RouterLink
        v-for="l in links"
        :key="l.to"
        :to="l.to"
        :class="{ active: l.to === '/' ? route.path === '/' : route.path.startsWith(l.to) }"
      >
        {{ l.label }}
      </RouterLink>
      <span class="user">
        <span class="sub">{{ currentUser }}</span>
        <button class="btn btn-sm" type="button" @click="logout">Keluar</button>
      </span>
    </nav>
  </header>

  <div v-if="notice" class="notice" :class="notice.kind" role="status">
    <span>{{ notice.text }}</span>
    <button class="btn btn-sm" type="button" @click="notice = null">Tutup</button>
  </div>

  <RouterView />
  <ConfirmDialog />
</template>

<style scoped>
.nav {
  position: sticky;
  top: 0;
  z-index: 10;
  display: flex;
  align-items: center;
  gap: 32px;
  height: 64px;
  padding: 0 48px;
  background: var(--canvas);
  border-bottom: 1px solid var(--hairline);
}
.wordmark { color: var(--ink); font-size: 18px; font-weight: 500; }
nav { display: flex; align-items: center; gap: 4px; flex: 1; }
nav a {
  display: inline-flex;
  align-items: center;
  min-height: 44px;
  padding: 0 12px;
  color: var(--body);
  border-radius: var(--r-sm);
}
/* Current page: ink underline, the nav's single emphasis. */
nav a.active { color: var(--ink); box-shadow: inset 0 -2px 0 var(--ink); border-radius: 0; }
.user { margin-left: auto; display: flex; align-items: center; gap: 12px; }
.sub { color: var(--muted); }
.menu-toggle { display: none; margin-left: auto; }

.notice {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 16px;
  padding: 8px 48px;
  background: var(--surface-soft);
  color: var(--ink);
  border-bottom: 1px solid var(--hairline);
}
.notice.error { background: #fbeae4; color: var(--coral); }
.notice span { overflow-wrap: anywhere; }

@media (max-width: 900px) {
  .nav { padding: 0 16px; gap: 16px; }
  .menu-toggle { display: inline-flex; }
  nav {
    display: none;
    position: fixed;
    inset: 64px 0 0 0;
    flex-direction: column;
    align-items: stretch;
    padding: 16px;
    background: var(--canvas);
  }
  nav.open { display: flex; }
  nav a { font-size: 20px; min-height: 52px; border-bottom: 1px solid var(--hairline); }
  nav a.active { box-shadow: inset 3px 0 0 var(--ink); padding-left: 16px; }
  .user { margin: 24px 0 0; justify-content: space-between; }
  .notice { padding: 8px 16px; }
}
</style>
