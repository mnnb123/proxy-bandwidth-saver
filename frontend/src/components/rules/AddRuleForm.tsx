import { useState } from 'react'
import { X } from 'lucide-react'
import { useRulesStore } from '../../stores/rulesStore'

const ACTIONS = [
  { value: 'bypass', label: 'Bypass', desc: 'Không qua proxy, sử dụng mạng bình thường', color: 'text-[var(--color-success)]' },
  { value: 'block', label: 'Block', desc: 'Chặn hoàn toàn traffic của domain', color: 'text-[var(--color-danger)]' },
  { value: 'bypass_vps', label: 'Bypass VPS', desc: 'Không dùng proxy upstream, dùng IP VPS', color: 'text-[var(--color-info-text)]' },
]

interface Props {
  onClose: () => void
}

export function AddRuleForm({ onClose }: Props) {
  const { createBulkRules } = useRulesStore()
  const [domains, setDomains] = useState('')
  const [action, setAction] = useState('bypass')
  const [saving, setSaving] = useState(false)

  const domainCount = domains.split('\n').filter((d) => d.trim()).length

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    const patterns = domains.split('\n').map((d) => d.trim()).filter(Boolean)
    if (patterns.length === 0) return
    setSaving(true)
    await createBulkRules(patterns, action, 100)
    setSaving(false)
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl w-[520px] max-h-[80vh] overflow-hidden shadow-2xl" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-3 border-b border-[var(--color-border)]">
          <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">Add Domain Rules</h2>
          <button onClick={onClose} className="p-1 rounded hover:bg-[var(--color-sidebar-hover)] text-[var(--color-text-muted)]">
            <X size={16} />
          </button>
        </div>

        <form onSubmit={handleSubmit} className="p-5 space-y-4">
          {/* Action Selection */}
          <div>
            <label className="block text-xs font-medium text-[var(--color-text-muted)] mb-2">Action</label>
            <div className="grid grid-cols-3 gap-2">
              {ACTIONS.map((a) => (
                <button
                  key={a.value}
                  type="button"
                  onClick={() => setAction(a.value)}
                  className={`px-3 py-2.5 rounded-lg border text-left transition-all ${
                    action === a.value
                      ? 'border-[var(--color-primary)] bg-[var(--color-primary)]/10 ring-1 ring-[var(--color-primary)]'
                      : 'border-[var(--color-border)] bg-[var(--color-bg-elevated)] hover:bg-[var(--color-sidebar-hover)]'
                  }`}
                >
                  <div className={`text-xs font-semibold ${a.color}`}>{a.label}</div>
                  <div className="text-[10px] text-[var(--color-text-muted)] mt-0.5 leading-tight">{a.desc}</div>
                </button>
              ))}
            </div>
          </div>

          {/* Domain List */}
          <div>
            <div className="flex items-center justify-between mb-2">
              <label className="text-xs font-medium text-[var(--color-text-muted)]">Domain List</label>
              <span className="text-[10px] text-[var(--color-text-muted)]">{domainCount} domain{domainCount !== 1 ? 's' : ''}</span>
            </div>
            <textarea
              value={domains}
              onChange={(e) => setDomains(e.target.value)}
              placeholder={"example.com\n*.google.com\nfacebook\nyoutube.com\n*.tiktok.com"}
              rows={10}
              className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-lg px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] font-mono resize-none leading-relaxed"
              autoFocus
            />
            <p className="text-[10px] text-[var(--color-text-muted)] mt-1">
              Mỗi dòng 1 domain. Định dạng: <b>example.com</b> (domain + subdomain) | <b>*.google.com</b> (wildcard) | <b>facebook</b> (từ khóa, match tất cả domain chứa từ này)
            </p>
          </div>

          {/* Actions */}
          <div className="flex items-center justify-end gap-2 pt-2">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-xs rounded-lg bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)] transition-colors"
            >
              Cancel
            </button>
            <button
              type="submit"
              disabled={domainCount === 0 || saving}
              className="px-4 py-2 text-xs rounded-lg bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-50"
            >
              {saving ? 'Adding...' : `Add ${domainCount} Rule${domainCount !== 1 ? 's' : ''}`}
            </button>
          </div>
        </form>
      </div>
    </div>
  )
}
