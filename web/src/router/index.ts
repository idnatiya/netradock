import { createRouter, createWebHistory } from 'vue-router'
import { api, currentUser } from '@/api'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue'), meta: { guest: true } },
    { path: '/', name: 'overview', component: () => import('@/views/OverviewView.vue') },
    { path: '/containers', name: 'containers', component: () => import('@/views/ContainersView.vue') },
    { path: '/containers/:id', name: 'container', component: () => import('@/views/ContainerView.vue') },
    { path: '/images', name: 'images', component: () => import('@/views/ImagesView.vue') },
    { path: '/volumes', name: 'volumes', component: () => import('@/views/VolumesView.vue') },
    { path: '/networks', name: 'networks', component: () => import('@/views/NetworksView.vue') },
    { path: '/:pathMatch(.*)*', redirect: '/' },
  ],
})

router.beforeEach(async (to) => {
  if (to.meta.guest) return true
  if (currentUser.value === null) {
    try {
      currentUser.value = (await api<{ username: string }>('GET', '/auth/me')).username
    } catch {
      return { name: 'login', query: { next: to.fullPath } }
    }
  }
  return true
})

export default router
