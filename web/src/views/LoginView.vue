<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { AlertCircle, Loader2 } from 'lucide-vue-next'
import { api, currentUser } from '@/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import BrandLogo from '@/components/BrandLogo.vue'

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
      ? 'Too many attempts. Wait a minute and try again.'
      : (e as Error).message === 'invalid username or password'
        ? 'Invalid username or password.'
        : (e as Error).message
  } finally {
    busy.value = false
  }
}
</script>

<template>
  <div class="min-h-dvh flex flex-col items-center justify-center p-4 bg-background text-foreground">
    <div class="w-full max-w-sm space-y-6">
      <div class="flex justify-center">
        <BrandLogo />
      </div>

      <Card class="shadow-sm border-border">
        <CardHeader class="space-y-1 text-center pb-4">
          <CardTitle class="text-xl font-bold">Sign In</CardTitle>
          <CardDescription class="text-xs">
            Authenticate to manage Docker containers
          </CardDescription>
        </CardHeader>

        <CardContent>
          <form class="space-y-4" autocomplete="on" :aria-busy="busy" @submit.prevent="submit">
            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground" for="username">Username</label>
              <Input
                id="username"
                v-model="username"
                type="text"
                autocomplete="username"
                required
                autofocus
                placeholder="admin"
                class="h-9 text-xs"
              />
            </div>

            <div class="space-y-1.5">
              <label class="text-xs font-medium text-foreground" for="password">Password</label>
              <Input
                id="password"
                v-model="password"
                type="password"
                autocomplete="current-password"
                required
                placeholder="••••••••"
                class="h-9 text-xs"
              />
            </div>

            <div v-if="error" class="flex items-center gap-2 p-2.5 rounded-md border border-destructive/20 bg-destructive/10 text-xs text-destructive">
              <AlertCircle class="size-4 shrink-0" />
              <span>{{ error }}</span>
            </div>

            <Button type="submit" class="w-full h-9 text-xs gap-2 font-medium" :disabled="busy">
              <Loader2 v-if="busy" class="size-3.5 animate-spin" />
              <span>{{ busy ? 'Authenticating...' : 'Sign In' }}</span>
            </Button>
          </form>
        </CardContent>
      </Card>

      <div class="text-center text-[11px] font-mono text-muted-foreground">
        Configured via <code class="px-1 py-0.5 rounded bg-muted">NETRADOCK_USERNAME</code> & <code class="px-1 py-0.5 rounded bg-muted">NETRADOCK_PASSWORD</code>
      </div>
    </div>
  </div>
</template>
