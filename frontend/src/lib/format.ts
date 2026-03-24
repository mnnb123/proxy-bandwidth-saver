export function formatBytes(bytes: number): string {
  if (bytes === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const k = 1024
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  const val = bytes / Math.pow(k, i)
  return `${val.toFixed(i > 1 ? 2 : 0)} ${units[i]}`
}

export function formatBytesPerSec(bps: number): string {
  return `${formatBytes(bps)}/s`
}

export function formatCost(usd: number): string {
  return `$${usd.toFixed(2)}`
}

export function formatPercent(ratio: number): string {
  return `${(ratio * 100).toFixed(1)}%`
}
