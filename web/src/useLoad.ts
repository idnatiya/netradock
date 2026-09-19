import { onMounted, onUnmounted, ref } from 'vue'

// Loads data on mount and optionally re-polls; errors keep the last good data visible.
export function useLoad<T>(fn: () => Promise<T>, pollMs = 0) {
  const data = ref<T>()
  const error = ref('')
  const loading = ref(true)
  let timer: number | undefined

  async function reload() {
    try {
      data.value = await fn()
      error.value = ''
    } catch (e) {
      error.value = (e as Error).message
    } finally {
      loading.value = false
    }
  }

  onMounted(() => {
    reload()
    if (pollMs) timer = window.setInterval(() => document.hidden || reload(), pollMs)
  })
  onUnmounted(() => clearInterval(timer))

  return { data, error, loading, reload }
}
