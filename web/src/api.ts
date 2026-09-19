import { ref } from 'vue'
import router from './router'

export interface Port { ip?: string; private_port: number; public_port?: number; type: string }
export interface Container { id: string; name: string; image: string; state: string; status: string; created: number; ports: Port[]; project?: string }
export interface Image { id: string; tags: string[]; size: number; created: number; containers: number }
export interface Volume { name: string; driver: string; mountpoint: string; created_at: string }
export interface Network { id: string; name: string; driver: string; scope: string; created: number }
export interface System {
  name: string; server_version: string; operating_system: string; kernel_version: string; architecture: string
  ncpu: number; mem_total: number; containers: number; containers_running: number; containers_paused: number
  containers_stopped: number; images: number
}
export interface Stats { cpu_percent: number; mem_usage: number; mem_limit: number; net_rx: number; net_tx: number }

export const currentUser = ref<string | null>(null)

// Global action feedback, shown in the status bar under the nav.
export const notice = ref<{ kind: 'ok' | 'error'; text: string } | null>(null)
export function notify(kind: 'ok' | 'error', text: string) {
  notice.value = { kind, text }
}

export async function api<T>(method: string, path: string, body?: unknown): Promise<T> {
  const res = await fetch('/api' + path, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : undefined,
    body: body ? JSON.stringify(body) : undefined,
  })
  const json = await res.json().catch(() => ({}))
  if (res.status === 401 && path !== '/auth/login') {
    currentUser.value = null
    router.push({ name: 'login', query: { next: router.currentRoute.value.fullPath } })
  }
  if (!res.ok) throw new Error(json.errors || `${res.status} ${res.statusText}`)
  return json.data as T
}

export function wsURL(path: string) {
  return `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws${path}`
}
