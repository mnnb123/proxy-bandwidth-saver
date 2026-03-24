import { useState } from 'react'
import { FlaskConical } from 'lucide-react'
import { useRulesStore } from '../../stores/rulesStore'
import { RouteBadge } from '../ui/Badge'

export function RuleTester() {
  const { testResult, testRule } = useRulesStore()
  const [domain, setDomain] = useState('')
  const [url, setUrl] = useState('')
  const [contentType, setContentType] = useState('')

  const handleTest = async () => {
    if (!domain && !url) return
    await testRule(domain, url, contentType)
  }

  return (
    <div className="bg-[var(--color-bg-surface)] border border-[var(--color-primary)]/30 rounded-xl p-4">
      <div className="flex items-center gap-2 mb-3">
        <FlaskConical size={14} className="text-[var(--color-primary)]" />
        <span className="text-xs font-medium text-[var(--color-text-primary)]">Rule Tester</span>
      </div>
      <div className="grid grid-cols-4 gap-3">
        <div>
          <label className="block text-xs text-[var(--color-text-muted)] mb-1">Domain</label>
          <input
            value={domain}
            onChange={(e) => setDomain(e.target.value)}
            placeholder="example.com"
            className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] font-mono"
          />
        </div>
        <div>
          <label className="block text-xs text-[var(--color-text-muted)] mb-1">URL Path</label>
          <input
            value={url}
            onChange={(e) => setUrl(e.target.value)}
            placeholder="/api/data"
            className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] font-mono"
          />
        </div>
        <div>
          <label className="block text-xs text-[var(--color-text-muted)] mb-1">Content-Type</label>
          <input
            value={contentType}
            onChange={(e) => setContentType(e.target.value)}
            placeholder="text/html"
            className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] font-mono"
          />
        </div>
        <div className="flex items-end">
          <button
            onClick={handleTest}
            className="w-full px-4 py-2 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors"
          >
            Test Classification
          </button>
        </div>
      </div>
      {testResult && (
        <div className="mt-3 pt-3 border-t border-[var(--color-border)] flex items-center gap-3">
          <span className="text-xs text-[var(--color-text-muted)]">Result:</span>
          <RouteBadge route={testResult} />
        </div>
      )}
    </div>
  )
}
