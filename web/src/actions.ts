import { api, notify } from '@/api'
import { confirm } from '@/confirm'

const verbs = { start: 'dijalankan', stop: 'dihentikan', restart: 'di-restart' } as const

// Returns true when the action ran, so callers know to reload.
export async function containerAction(name: string, id: string, action: 'start' | 'stop' | 'restart' | 'remove', running = false) {
  if (action === 'stop' && !(await confirm(`Hentikan ${name}?`, 'Proses di dalam container akan menerima SIGTERM.', 'Hentikan'))) return false
  if (action === 'remove') {
    const body = running
      ? 'Container sedang berjalan dan akan dihentikan paksa. Data di luar volume ikut hilang.'
      : 'Data di luar volume ikut hilang. Tindakan ini tidak bisa dibatalkan.'
    if (!(await confirm(`Hapus ${name}?`, body, 'Hapus'))) return false
  }
  try {
    if (action === 'remove') {
      await api('DELETE', `/containers/${id}?force=${running}`)
      notify('ok', `${name} dihapus.`)
    } else {
      await api('POST', `/containers/${id}/${action}`)
      notify('ok', `${name} ${verbs[action]}.`)
    }
    return true
  } catch (e) {
    notify('error', `Gagal: ${(e as Error).message}`)
    return false
  }
}

export async function removeResource(kind: string, label: string, path: string, body = 'Tindakan ini tidak bisa dibatalkan.') {
  if (!(await confirm(`Hapus ${kind} ${label}?`, body, 'Hapus'))) return false
  try {
    await api('DELETE', path)
    notify('ok', `${kind} ${label} dihapus.`)
    return true
  } catch (e) {
    notify('error', `Gagal menghapus ${label}: ${(e as Error).message}`)
    return false
  }
}
