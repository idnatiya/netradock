<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { api, currentUser } from '@/api'

const route = useRoute()
const router = useRouter()
const username = ref('')
const password = ref('')
const error = ref('')
const busy = ref(false)

async function submit() {
  busy.value = true
  error.value = ''
  try {
    const user = await api<{ username: string }>('POST', '/auth/login', { username: username.value, password: password.value })
    currentUser.value = user.username
    const next = typeof route.query.next === 'string' && route.query.next.startsWith('/') ? route.query.next : '/'
    router.replace(next)
  } catch (e) {
    error.value = (e as Error).message === 'Too Many Requests'
      ? 'Terlalu banyak percobaan. Tunggu satu menit lalu coba lagi.'
      : (e as Error).message === 'invalid username or password'
        ? 'Username atau password salah.'
        : (e as Error).message
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="page page-center">
    <div class="container container-tight py-4">
      <div class="text-center mb-4">
        <span class="navbar-brand fs-2">Netradock</span>
      </div>
      <div class="card card-md">
        <div class="card-body">
          <h2 class="h2 text-center mb-4">Masuk untuk mengelola Docker</h2>
          <form autocomplete="on" :aria-busy="busy" @submit.prevent="submit">
            <div class="mb-3">
              <label class="form-label" for="username">Username</label>
              <input id="username" v-model="username" type="text" class="form-control" autocomplete="username" required autofocus>
            </div>
            <div class="mb-3">
              <label class="form-label" for="password">Password</label>
              <input id="password" v-model="password" type="password" class="form-control" autocomplete="current-password" required>
            </div>
            <div v-if="error" class="alert alert-danger" role="alert">{{ error }}</div>
            <div class="form-footer">
              <button type="submit" class="btn btn-primary w-100" :disabled="busy">
                <span v-if="busy" class="spinner-border spinner-border-sm me-2" aria-hidden="true"></span>
                {{ busy ? 'Memeriksa...' : 'Masuk' }}
              </button>
            </div>
          </form>
        </div>
      </div>
      <div class="text-center text-secondary mt-3">
        Akun diatur lewat <code>NETRADOCK_USERNAME</code> dan <code>NETRADOCK_PASSWORD</code>.
      </div>
    </div>
  </div>
</template>
