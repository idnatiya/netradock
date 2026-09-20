import { api, notify } from '@/api'
import { confirm } from '@/confirm'

const verbs = { start: 'started', stop: 'stopped', restart: 'restarted' } as const

// Returns true when the action ran, so callers know to reload.
export async function containerAction(name: string, id: string, action: 'start' | 'stop' | 'restart' | 'remove', running = false) {
  if (action === 'stop' && !(await confirm(`Stop ${name}?`, 'Processes inside the container will receive SIGTERM.', 'Stop'))) return false
  if (action === 'remove') {
    const body = running
      ? 'Container is running and will be forcefully stopped. Any data outside persistent volumes will be permanently lost.'
      : 'Any data outside persistent volumes will be permanently lost. This action cannot be undone.'
    if (!(await confirm(`Remove ${name}?`, body, 'Remove'))) return false
  }
  try {
    if (action === 'remove') {
      await api('DELETE', `/containers/${id}?force=${running}`)
      notify('ok', `${name} removed.`)
    } else {
      await api('POST', `/containers/${id}/${action}`)
      notify('ok', `${name} ${verbs[action]}.`)
    }
    return true
  } catch (e) {
    notify('error', `Failed: ${(e as Error).message}`)
    return false
  }
}

export async function removeResource(kind: string, label: string, path: string, body = 'This action cannot be undone.') {
  if (!(await confirm(`Delete ${kind} ${label}?`, body, 'Delete'))) return false
  try {
    await api('DELETE', path)
    notify('ok', `${kind} ${label} deleted.`)
    return true
  } catch (e) {
    notify('error', `Failed to delete ${label}: ${(e as Error).message}`)
    return false
  }
}
