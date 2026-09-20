<script setup lang="ts">
import { ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import {
  AlertCircle,
  CheckCircle2,
  Loader2,
  Moon,
  ShieldCheck,
  Sun,
} from 'lucide-vue-next'
import { api, currentUser } from '@/api'
import { theme, toggleTheme } from '@/theme'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
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
  <div class="relative min-h-dvh w-full grid lg:grid-cols-2 bg-background text-foreground overflow-hidden">
    <!-- Top-right theme toggle -->
    <div class="absolute right-4 top-4 z-20">
      <Button
        variant="ghost"
        size="icon-sm"
        class="size-8"
        :title="theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'"
        @click="toggleTheme"
      >
        <Sun v-if="theme === 'dark'" class="size-4 text-muted-foreground hover:text-foreground" />
        <Moon v-else class="size-4 text-muted-foreground hover:text-foreground" />
      </Button>
    </div>

    <!-- Left Section: Hero Showcase (visible on lg screens) -->
    <div class="hidden lg:flex relative flex-col justify-between p-12 bg-sidebar border-r border-sidebar select-none overflow-hidden">
      <!-- Ambient Background Glow -->
      <div class="pointer-events-none absolute -top-24 -left-24 size-96 rounded-full bg-primary/10 blur-3xl" />
      <div class="pointer-events-none absolute -bottom-24 -right-24 size-96 rounded-full bg-sky-500/10 blur-3xl" />

      <!-- Top Header Brand -->
      <div class="relative z-10">
        <BrandLogo />
      </div>

      <!-- Center Hero Graphic & Value Proposition -->
      <div class="relative z-10 space-y-8 my-auto py-8">
        <div class="space-y-3">
          <div class="inline-flex items-center gap-2 px-3 py-1 rounded-full border border-primary/20 bg-primary/10 text-primary text-xs font-mono">
            <span class="size-2 rounded-full bg-primary animate-pulse-dot" />
            <span>Docker Control Plane</span>
          </div>
          <h2 class="text-3xl font-bold tracking-tight text-foreground leading-snug">
            Streamlined Docker infrastructure on your VPS.
          </h2>
          <p class="text-sm text-muted-foreground leading-relaxed max-w-md">
            Real-time container telemetry, streaming logs, web-based interactive terminal shell, and storage diagnostics with zero external dependencies.
          </p>
        </div>

        <!-- Simulated Daemon Mockup Card -->
        <div class="w-full max-w-md rounded-xl border border-border bg-card/80 p-4 shadow-sm backdrop-blur-xs space-y-3 font-mono text-xs">
          <div class="flex items-center justify-between border-b border-border pb-2.5">
            <div class="flex items-center gap-1.5">
              <span class="size-2.5 rounded-full bg-rose-500/80" />
              <span class="size-2.5 rounded-full bg-amber-500/80" />
              <span class="size-2.5 rounded-full bg-emerald-500/80" />
              <span class="text-[11px] text-muted-foreground ml-2">/var/run/docker.sock</span>
            </div>
            <div class="flex items-center gap-1.5 text-[11px] text-emerald-500">
              <span class="size-1.5 rounded-full bg-emerald-500 animate-pulse-dot" />
              <span>connected</span>
            </div>
          </div>

          <div class="space-y-2 text-[12px]">
            <div class="flex items-center justify-between text-muted-foreground text-[11px]">
              <span>CONTAINER</span>
              <span>STATUS</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-foreground font-medium flex items-center gap-1.5">
                <span class="size-1.5 rounded-full bg-emerald-500" />
                netradock-server
              </span>
              <span class="text-emerald-500">Up 18 days</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-foreground font-medium flex items-center gap-1.5">
                <span class="size-1.5 rounded-full bg-emerald-500" />
                postgres-db
              </span>
              <span class="text-emerald-500">Up 18 days</span>
            </div>
            <div class="flex items-center justify-between">
              <span class="text-foreground font-medium flex items-center gap-1.5">
                <span class="size-1.5 rounded-full bg-emerald-500" />
                caddy-gateway
              </span>
              <span class="text-emerald-500">Up 18 days</span>
            </div>
          </div>
        </div>

        <!-- Highlights feature list -->
        <div class="grid grid-cols-2 gap-3 max-w-md text-xs text-muted-foreground pt-2">
          <div class="flex items-center gap-2">
            <CheckCircle2 class="size-4 text-primary shrink-0" />
            <span>WebSocket Live Stats</span>
          </div>
          <div class="flex items-center gap-2">
            <CheckCircle2 class="size-4 text-primary shrink-0" />
            <span>XTerm Shell Attach</span>
          </div>
          <div class="flex items-center gap-2">
            <CheckCircle2 class="size-4 text-primary shrink-0" />
            <span>Lightweight Single Binary</span>
          </div>
          <div class="flex items-center gap-2">
            <CheckCircle2 class="size-4 text-primary shrink-0" />
            <span>Stateless Token Auth</span>
          </div>
        </div>
      </div>

      <!-- Bottom Quote / Reliability Footer -->
      <div class="relative z-10 pt-6 border-t border-sidebar-border">
        <blockquote class="text-xs text-muted-foreground italic">
          "Simplicity is prerequisite for reliability."
        </blockquote>
        <div class="text-[11px] font-mono text-muted-foreground/80 mt-1">
          — Edsger W. Dijkstra
        </div>
      </div>
    </div>

    <!-- Right Section: Authentication Form -->
    <div class="flex items-center justify-center p-6 sm:p-12">
      <div class="w-full max-w-[380px] space-y-6">
        <!-- Mobile Logo (hidden on lg) -->
        <div class="flex justify-center lg:hidden mb-2">
          <BrandLogo />
        </div>

        <!-- Form Heading -->
        <div class="space-y-1.5 text-center lg:text-left">
          <h1 class="text-2xl font-bold tracking-tight text-foreground">Sign In</h1>
          <p class="text-sm text-muted-foreground">
            Enter your credentials to access the Docker management console
          </p>
        </div>

        <!-- Form -->
        <form class="space-y-4" autocomplete="on" :aria-busy="busy" @submit.prevent="submit">
          <div class="space-y-1.5">
            <label class="text-sm font-medium text-foreground" for="username">Username</label>
            <Input
              id="username"
              v-model="username"
              type="text"
              autocomplete="username"
              required
              autofocus
              placeholder="admin"
              class="h-9.5 text-sm"
            />
          </div>

          <div class="space-y-1.5">
            <label class="text-sm font-medium text-foreground" for="password">Password</label>
            <Input
              id="password"
              v-model="password"
              type="password"
              autocomplete="current-password"
              required
              placeholder="••••••••"
              class="h-9.5 text-sm"
            />
          </div>

          <!-- Error Banner -->
          <div
            v-if="error"
            class="flex items-center gap-2.5 p-3 rounded-md border border-destructive/20 bg-destructive/10 text-xs text-destructive"
            role="alert"
          >
            <AlertCircle class="size-4 shrink-0" />
            <span>{{ error }}</span>
          </div>

          <!-- Submit Button -->
          <Button
            type="submit"
            class="w-full h-9.5 text-sm font-semibold gap-2 shadow-sm cursor-pointer"
            :disabled="busy"
          >
            <Loader2 v-if="busy" class="size-4 animate-spin" />
            <span>{{ busy ? 'Authenticating...' : 'Sign In' }}</span>
          </Button>
        </form>

        <!-- Credentials Hint Card -->
        <div class="p-3.5 rounded-lg border border-border bg-card/60 text-center space-y-1">
          <p class="text-xs text-muted-foreground">Configured via environment variables:</p>
          <div class="flex items-center justify-center gap-1.5 flex-wrap">
            <code class="px-1.5 py-0.5 rounded bg-muted font-mono text-[11px] text-foreground">NETRADOCK_USERNAME</code>
            <span class="text-xs text-muted-foreground">&</span>
            <code class="px-1.5 py-0.5 rounded bg-muted font-mono text-[11px] text-foreground">NETRADOCK_PASSWORD</code>
          </div>
        </div>

        <!-- Security Note -->
        <p class="text-center text-xs text-muted-foreground flex items-center justify-center gap-1.5">
          <ShieldCheck class="size-3.5 text-emerald-500" />
          <span>Protected with HTTP-only cookie and HMAC token security.</span>
        </p>
      </div>
    </div>
  </div>
</template>
