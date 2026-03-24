import { useState } from 'react'
import { Modal } from '../ui/Modal'

interface Props {
  onClose: () => void
  onImport: (text: string) => Promise<number>
}

export function BulkImportModal({ onClose, onImport }: Props) {
  const [text, setText] = useState('')

  const handleImport = async () => {
    if (!text.trim()) return
    await onImport(text)
    onClose()
  }

  const lineCount = text.trim() ? text.trim().split('\n').filter(Boolean).length : 0

  return (
    <Modal open onClose={onClose} title="Bulk Import Proxies" wide>
      <div className="space-y-4">
        <div className="text-xs text-[var(--color-text-muted)] space-y-1">
          <p>Paste proxies, one per line. Supported formats:</p>
          <div className="bg-[var(--color-input-bg)] rounded-[var(--radius-lg)] p-2.5 font-mono text-[11px] text-[var(--color-text-secondary)] space-y-0.5">
            <div><span className="text-[var(--color-primary)]">HTTP:</span> host:port | host:port:user:pass | http://user:pass@host:port</div>
            <div><span className="text-[var(--color-warning)]">SOCKS5:</span> host:port:socks5 | host:port:user:pass:socks5 | socks5://user:pass@host:port</div>
          </div>
        </div>
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="192.168.1.1:8080&#10;proxy.example.com:3128:user:pass&#10;http://user:pass@10.0.0.1:8888"
          rows={10}
          className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] font-mono resize-none"
        />
        <div className="flex items-center justify-between">
          <span className="text-xs text-[var(--color-text-muted)]">{lineCount} proxies detected</span>
          <div className="flex gap-3">
            <button onClick={onClose} className="px-4 py-2 text-xs rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors">
              Cancel
            </button>
            <button
              onClick={handleImport}
              disabled={lineCount === 0}
              className="px-4 py-2 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Import {lineCount} Proxies
            </button>
          </div>
        </div>
      </div>
    </Modal>
  )
}
