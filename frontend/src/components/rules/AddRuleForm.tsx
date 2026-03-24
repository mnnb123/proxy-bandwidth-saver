import { useState } from 'react'
import { useRulesStore } from '../../stores/rulesStore'

const TYPES = [
  { value: 'domain', label: 'Domain' },
  { value: 'content_type', label: 'Content-Type' },
  { value: 'url_pattern', label: 'URL Pattern' },
]

const ACTIONS = ['direct', 'datacenter', 'residential']

interface Props {
  defaultType: string
  onDone: () => void
}

export function AddRuleForm({ defaultType, onDone }: Props) {
  const { createRule } = useRulesStore()
  const [ruleType, setRuleType] = useState(defaultType)
  const [pattern, setPattern] = useState('')
  const [action, setAction] = useState('direct')
  const [priority, setPriority] = useState(100)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!pattern.trim()) return
    await createRule(ruleType, pattern.trim(), action, priority)
    setPattern('')
    onDone()
  }

  return (
    <form onSubmit={handleSubmit} className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl p-4">
      <div className="grid grid-cols-5 gap-3">
        <div>
          <label className="block text-xs text-[var(--color-text-muted)] mb-1">Type</label>
          <select
            value={ruleType}
            onChange={(e) => setRuleType(e.target.value)}
            className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)]"
          >
            {TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
          </select>
        </div>
        <div className="col-span-2">
          <label className="block text-xs text-[var(--color-text-muted)] mb-1">Pattern</label>
          <input
            value={pattern}
            onChange={(e) => setPattern(e.target.value)}
            placeholder={ruleType === 'domain' ? '*.example.com' : ruleType === 'content_type' ? 'image/*' : '/api/.*'}
            className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] font-mono"
          />
        </div>
        <div>
          <label className="block text-xs text-[var(--color-text-muted)] mb-1">Route</label>
          <select
            value={action}
            onChange={(e) => setAction(e.target.value)}
            className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)]"
          >
            {ACTIONS.map((a) => <option key={a} value={a}>{a}</option>)}
          </select>
        </div>
        <div>
          <label className="block text-xs text-[var(--color-text-muted)] mb-1">Priority</label>
          <div className="flex gap-2">
            <input
              type="number"
              value={priority}
              onChange={(e) => setPriority(Number(e.target.value))}
              className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-2 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)]"
            />
            <button
              type="submit"
              className="px-4 py-2 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors whitespace-nowrap"
            >
              Add
            </button>
          </div>
        </div>
      </div>
    </form>
  )
}
