const routeColors: Record<string, string> = {
  direct: 'bg-[var(--color-success-bg)] text-[var(--color-success-text)] border-[var(--color-success)]/30',
  bypass: 'bg-[var(--color-success-bg)] text-[var(--color-success-text)] border-[var(--color-success)]/30',
  bypass_vps: 'bg-[var(--color-info-bg)] text-[var(--color-info-text)] border-[var(--color-info)]/30',
  datacenter: 'bg-[var(--color-info-bg)] text-[var(--color-info-text)] border-[var(--color-info)]/30',
  residential: 'bg-[var(--color-danger-bg)] text-[var(--color-danger-text)] border-[var(--color-danger)]/30',
  block: 'bg-red-500/10 text-red-400 border-red-500/30',
}

const routeLabels: Record<string, string> = {
  direct: 'Bypass',
  bypass: 'Bypass',
  bypass_vps: 'Bypass VPS',
  datacenter: 'Datacenter',
  residential: 'Residential',
  block: 'Block',
}

export function RouteBadge({ route }: { route: string }) {
  const colors = routeColors[route] || 'bg-[var(--color-bg-elevated)] text-[var(--color-text-muted)] border-[var(--color-border)]'
  const label = routeLabels[route] || route
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${colors}`}>
      {label}
    </span>
  )
}

export function TypeBadge({ type }: { type: string }) {
  return (
    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border border-[var(--color-border)]">
      {type}
    </span>
  )
}
