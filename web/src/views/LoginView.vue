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
  <main class="login">
    <form @submit.prevent="submit" :aria-busy="busy">
      <p class="wordmark">Netradock</p>
      <h1>Masuk untuk mengelola Docker di server ini</h1>
      <label>
        Username
        <input v-model="username" type="text" name="username" autocomplete="username" required autofocus>
      </label>
      <label>
        Password
        <input v-model="password" type="password" name="password" autocomplete="current-password" required>
      </label>
      <p v-if="error" class="error" role="alert">{{ error }}</p>
      <button class="btn btn-primary" type="submit" :disabled="busy">{{ busy ? 'Memeriksa...' : 'Masuk' }}</button>
      <p class="hint">Akun diatur lewat <code>NETRADOCK_USERNAME</code> dan <code>NETRADOCK_PASSWORD</code>.</p>
    </form>
  </main>
</template>

<style scoped>
.login { min-height: 100dvh; display: grid; place-items: center; padding: 24px 16px; background: var(--surface-soft); }
form {
  width: min(400px, 100%);
  display: grid;
  gap: 16px;
  padding: 32px;
  background: var(--canvas);
  border: 1px solid var(--hairline);
  border-radius: var(--r-lg);
}
.wordmark { margin: 0; color: var(--ink); font-weight: 500; font-size: 18px; }
h1 { font-size: 24px; line-height: 1.3; margin-bottom: 8px; }
label { display: grid; gap: 6px; color: var(--ink); font-weight: 500; }
.error { margin: 0; color: var(--coral); }
.hint { margin: 0; color: var(--muted); font-size: 13px; }
@media (max-width: 480px) { form { padding: 24px 20px; } }
</style>
