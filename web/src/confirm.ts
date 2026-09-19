import { ref } from 'vue'

export const pending = ref<{ title: string; body: string; action: string; resolve: (ok: boolean) => void } | null>(null)

export function confirm(title: string, body: string, action: string): Promise<boolean> {
  return new Promise((resolve) => {
    pending.value = { title, body, action, resolve }
  })
}
