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
function applyTheme(t: Theme) {
  document.documentElement.dataset.bsTheme = t
  if (t === 'dark') {
    document.documentElement.classList.add('dark')
  } else {
    document.documentElement.classList.remove('dark')
  }
}
applyTheme(theme.value)

export function toggleTheme() {
  theme.value = theme.value === 'dark' ? 'light' : 'dark'
  applyTheme(theme.value)
  try {
    localStorage.setItem(KEY, theme.value)
  } catch {
    // Not persisted; the choice still applies for this visit.
  }
}
