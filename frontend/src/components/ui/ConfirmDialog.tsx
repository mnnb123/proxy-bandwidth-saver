import { Modal } from './Modal'

interface ConfirmDialogProps {
  open: boolean
  onClose: () => void
  onConfirm: () => void
  title: string
  message: string
  confirmText?: string
  destructive?: boolean
}

export function ConfirmDialog({ open, onClose, onConfirm, title, message, confirmText = 'Confirm', destructive }: ConfirmDialogProps) {
  return (
    <Modal open={open} onClose={onClose} title={title}>
      <p className="text-sm text-[var(--color-text-secondary)] mb-6">{message}</p>
      <div className="flex justify-end gap-3">
        <button onClick={onClose} className="px-4 py-2 text-sm rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors">
          Cancel
        </button>
        <button
          onClick={() => { onConfirm(); onClose() }}
          className={`px-4 py-2 text-sm rounded-[var(--radius-lg)] font-medium transition-colors ${
            destructive
              ? 'bg-[var(--color-danger)] text-white hover:bg-[var(--color-danger-hover)]'
              : 'bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)]'
          }`}
        >
          {confirmText}
        </button>
      </div>
    </Modal>
  )
}
