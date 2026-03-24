import { useState, useEffect } from 'react'
import { Copy, Check } from 'lucide-react'
import { Modal } from '../ui/Modal'
import { useRulesStore } from '../../stores/rulesStore'

interface Props {
  mode: 'import' | 'export'
  onClose: () => void
}

export function ImportExportModal({ mode, onClose }: Props) {
  const { importRules, exportRules } = useRulesStore()
  const [text, setText] = useState('')
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    if (mode === 'export') {
      exportRules().then(setText)
    }
  }, [mode, exportRules])

  const handleImport = async () => {
    if (!text.trim()) return
    await importRules(text)
    onClose()
  }

  const handleCopy = async () => {
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <Modal open onClose={onClose} title={mode === 'import' ? 'Import Rules' : 'Export Rules'} wide>
      <div className="space-y-4">
        <textarea
          value={text}
          onChange={(e) => setText(e.target.value)}
          readOnly={mode === 'export'}
          placeholder={mode === 'import' ? 'Paste JSON rules here...' : ''}
          rows={12}
          className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] font-mono resize-none"
        />
        <div className="flex justify-end gap-3">
          {mode === 'export' && (
            <button
              onClick={handleCopy}
              className="flex items-center gap-1.5 px-4 py-2 text-xs rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors"
            >
              {copied ? <Check size={14} /> : <Copy size={14} />}
              {copied ? 'Copied!' : 'Copy'}
            </button>
          )}
          {mode === 'import' && (
            <button
              onClick={handleImport}
              className="px-4 py-2 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors"
            >
              Import
            </button>
          )}
          <button
            onClick={onClose}
            className="px-4 py-2 text-xs rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors"
          >
            Close
          </button>
        </div>
      </div>
    </Modal>
  )
}
