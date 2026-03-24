import { type ReactNode } from 'react'
import { Inbox } from 'lucide-react'

interface EmptyStateProps {
  icon?: typeof Inbox
  title: string
  description?: string
  action?: ReactNode
}

export function EmptyState({ icon: Icon = Inbox, title, description, action }: EmptyStateProps) {
  return (
    <div className="flex flex-col items-center justify-center py-12 text-center">
      <div className="p-3 rounded-xl bg-[var(--color-bg-elevated)] mb-4">
        <Icon size={32} className="text-[var(--color-text-muted)]" />
      </div>
      <h3 className="text-sm font-medium text-[var(--color-text-primary)] mb-1">{title}</h3>
      {description && <p className="text-xs text-[var(--color-text-muted)] max-w-xs mb-4">{description}</p>}
      {action}
    </div>
  )
}
