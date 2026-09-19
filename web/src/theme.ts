import { ref } from 'vue'

type Theme = 'light' | 'dark'
const KEY = 'netradock-theme'

function initial(): Theme {
  try {
    const saved = localStorage.getItem(KEY)
    if (saved === 'light' || saved === 'dark') return saved
  } catch {
    // Storage blocked: fall back to the system preference.
  }
  return matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

export const theme = ref<Theme>(initial())
document.documentElement.dataset.bsTheme = theme.value

export function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  document.documentElement.dataset.bsTheme = theme.value
  try {
    localStorage.setItem(KEY, theme.value)
  } catch {
    // Not persisted; the choice still applies for this visit.
  }
}
