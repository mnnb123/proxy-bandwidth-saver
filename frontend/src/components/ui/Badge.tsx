const routeColors: Record<string, string> = {
  direct: 'bg-[var(--color-success-bg)] text-[var(--color-success-text)] border-[var(--color-success)]/30',
  datacenter: 'bg-[var(--color-info-bg)] text-[var(--color-info-text)] border-[var(--color-info)]/30',
  residential: 'bg-[var(--color-danger-bg)] text-[var(--color-danger-text)] border-[var(--color-danger)]/30',
}

export function RouteBadge({ route }: { route: string }) {
  const colors = routeColors[route] || 'bg-[var(--color-bg-elevated)] text-[var(--color-text-muted)] border-[var(--color-border)]'
  return (
    <span className={`inline-flex items-center px-2 py-0.5 rounded text-xs font-medium border ${colors}`}>
      {route}
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
