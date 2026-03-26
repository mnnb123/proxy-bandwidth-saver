import { useState } from 'react'
import { X } from 'lucide-react'
import { useRulesStore } from '../../stores/rulesStore'

const ACTIONS = [
  { value: 'bypass', label: 'Bypass', desc: 'Không qua proxy, sử dụng mạng bình thường', color: 'text-[var(--color-success)]' },
  { value: 'block', label: 'Block', desc: 'Chặn hoàn toàn traffic của domain', color: 'text-[var(--color-danger)]' },
  { value: 'bypass_vps', label: 'Bypass VPS', desc: 'Không dùng proxy upstream, dùng IP VPS', color: 'text-[var(--color-info-text)]' },
]

const EXT_GROUPS = [
  { label: 'Images', exts: ['.jpg', '.jpeg', '.png', '.gif', '.webp', '.svg', '.ico', '.avif', '.bmp'] },
  { label: 'Video', exts: ['.mp4', '.webm', '.avi', '.mkv', '.mov', '.flv', '.wmv', '.m3u8', '.ts'] },
  { label: 'Audio', exts: ['.mp3', '.wav', '.ogg', '.flac', '.aac', '.m4a', '.wma'] },
  { label: 'Fonts', exts: ['.woff', '.woff2', '.ttf', '.eot', '.otf'] },
  { label: 'Static', exts: ['.css', '.js', '.map', '.json', '.xml', '.txt'] },
  { label: 'Files', exts: ['.zip', '.rar', '.7z', '.tar', '.gz', '.pdf', '.doc', '.xls'] },
]

type TabKey = 'domain' | 'extension'

interface Props {
  onClose: () => void
}

export function AddRuleForm({ onClose }: Props) {
  const { createBulkRules, createRule } = useRulesStore()
  const [tab, setTab] = useState<TabKey>('domain')
  const [domains, setDomains] = useState('')
  const [action, setAction] = useState('bypass')
  const [saving, setSaving] = useState(false)
  const [selectedExts, setSelectedExts] = useState<Set<string>>(new Set())

  const domainCount = domains.split('\n').filter((d) => d.trim()).length

  const toggleExt = (ext: string) => {
    setSelectedExts((prev) => {
      const next = new Set(prev)
      if (next.has(ext)) next.delete(ext)
      else next.add(ext)
      return next
    })
  }

  const toggleGroup = (exts: string[]) => {
    setSelectedExts((prev) => {
      const next = new Set(prev)
      const allSelected = exts.every((e) => next.has(e))
      if (allSelected) {
        exts.forEach((e) => next.delete(e))
      } else {
        exts.forEach((e) => next.add(e))
      }
      return next
    })
  }

  const handleSubmitDomains = async (e: React.FormEvent) => {
    e.preventDefault()
    const patterns = domains.split('\n').map((d) => d.trim()).filter(Boolean)
    if (patterns.length === 0) return
    setSaving(true)
    await createBulkRules(patterns, action, 100)
    setSaving(false)
    onClose()
  }

  const handleSubmitExts = async (e: React.FormEvent) => {
    e.preventDefault()
    if (selectedExts.size === 0) return
    setSaving(true)
    // Create URL pattern rules for each extension
    for (const ext of selectedExts) {
      await createRule('url_pattern', `*${ext}`, 'bypass', 50)
    }
    setSaving(false)
    onClose()
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onClose}>
      <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl w-[560px] max-h-[85vh] overflow-hidden shadow-2xl" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="flex items-center justify-between px-5 py-3 border-b border-[var(--color-border)]">
          <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">Add Rules</h2>
          <button onClick={onClose} className="p-1 rounded hover:bg-[var(--color-sidebar-hover)] text-[var(--color-text-muted)]">
            <X size={16} />
          </button>
        </div>

        {/* Tabs */}
        <div className="flex border-b border-[var(--color-border)]">
          <button
            onClick={() => setTab('domain')}
            className={`flex-1 px-4 py-2 text-xs font-medium border-b-2 transition-colors ${
              tab === 'domain'
                ? 'border-[var(--color-primary)] text-[var(--color-primary)]'
                : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
            }`}
          >
            Domain List
          </button>
          <button
            onClick={() => setTab('extension')}
            className={`flex-1 px-4 py-2 text-xs font-medium border-b-2 transition-colors ${
              tab === 'extension'
                ? 'border-[var(--color-primary)] text-[var(--color-primary)]'
                : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
            }`}
          >
            File Extensions Bypass
          </button>
        </div>

        {tab === 'domain' ? (
          <form onSubmit={handleSubmitDomains} className="p-5 space-y-4">
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
                rows={8}
                className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-lg px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] font-mono resize-none leading-relaxed"
                autoFocus
              />
              <p className="text-[10px] text-[var(--color-text-muted)] mt-1">
                Mỗi dòng 1 domain. Hỗ trợ: <b>example.com</b> (domain + subdomain) | <b>*.google.com</b> (wildcard) | <b>facebook</b> (keyword)
              </p>
            </div>

            <div className="flex items-center justify-end gap-2 pt-2">
              <button type="button" onClick={onClose} className="px-4 py-2 text-xs rounded-lg bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)] transition-colors">
                Cancel
              </button>
              <button type="submit" disabled={domainCount === 0 || saving} className="px-4 py-2 text-xs rounded-lg bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-50">
                {saving ? 'Adding...' : `Add ${domainCount} Rule${domainCount !== 1 ? 's' : ''}`}
              </button>
            </div>
          </form>
        ) : (
          <form onSubmit={handleSubmitExts} className="p-5 space-y-4">
            <p className="text-xs text-[var(--color-text-muted)]">
              Chọn các đuôi file muốn bypass (không qua proxy). Traffic tới URL có đuôi này sẽ kết nối trực tiếp.
            </p>
            <div className="space-y-3 max-h-[400px] overflow-y-auto">
              {EXT_GROUPS.map((group) => {
                const allSelected = group.exts.every((e) => selectedExts.has(e))
                const someSelected = group.exts.some((e) => selectedExts.has(e))
                return (
                  <div key={group.label} className="bg-[var(--color-bg-elevated)] border border-[var(--color-border)] rounded-lg p-3">
                    <label className="flex items-center gap-2 cursor-pointer mb-2">
                      <input
                        type="checkbox"
                        checked={allSelected}
                        ref={(el) => { if (el) el.indeterminate = someSelected && !allSelected }}
                        onChange={() => toggleGroup(group.exts)}
                        className="accent-[var(--color-primary)]"
                      />
                      <span className="text-xs font-semibold text-[var(--color-text-primary)]">{group.label}</span>
                      <span className="text-[10px] text-[var(--color-text-muted)]">({group.exts.length})</span>
                    </label>
                    <div className="flex flex-wrap gap-1.5 ml-5">
                      {group.exts.map((ext) => (
                        <button
                          key={ext}
                          type="button"
                          onClick={() => toggleExt(ext)}
                          className={`px-2 py-0.5 rounded text-[11px] font-mono border transition-all ${
                            selectedExts.has(ext)
                              ? 'bg-[var(--color-primary)]/15 text-[var(--color-primary)] border-[var(--color-primary)]/40'
                              : 'bg-[var(--color-bg-surface)] text-[var(--color-text-muted)] border-[var(--color-border)] hover:text-[var(--color-text-secondary)]'
                          }`}
                        >
                          {ext}
                        </button>
                      ))}
                    </div>
                  </div>
                )
              })}
            </div>

            <div className="flex items-center justify-between pt-2">
              <span className="text-[11px] text-[var(--color-text-muted)]">{selectedExts.size} extension{selectedExts.size !== 1 ? 's' : ''} selected</span>
              <div className="flex items-center gap-2">
                <button type="button" onClick={onClose} className="px-4 py-2 text-xs rounded-lg bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)] transition-colors">
                  Cancel
                </button>
                <button type="submit" disabled={selectedExts.size === 0 || saving} className="px-4 py-2 text-xs rounded-lg bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors disabled:opacity-50">
                  {saving ? 'Adding...' : `Bypass ${selectedExts.size} Extension${selectedExts.size !== 1 ? 's' : ''}`}
                </button>
              </div>
            </div>
          </form>
        )}
      </div>
    </div>
  )
}
