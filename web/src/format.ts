export function bytes(n: number) {
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  let i = 0
  while (n >= 1024 && i < units.length - 1) {
    n /= 1024
    i++
  }
  return `${n.toFixed(i === 0 ? 0 : 1)} ${units[i]}`
}

const rtf = new Intl.RelativeTimeFormat('en', { numeric: 'auto' })

export function ago(unixSeconds: number) {
  const diff = unixSeconds - Date.now() / 1000
  const steps: [number, Intl.RelativeTimeFormatUnit][] = [[60, 'second'], [60, 'minute'], [24, 'hour'], [30, 'day'], [12, 'month']]
  let v = diff
  for (const [size, unit] of steps) {
    if (Math.abs(v) < size) return rtf.format(Math.round(v), unit)
    v /= size
  }
  return rtf.format(Math.round(v), 'year')
}

export function shortId(id: string) {
  return id.replace(/^sha256:/, '').slice(0, 12)
}
