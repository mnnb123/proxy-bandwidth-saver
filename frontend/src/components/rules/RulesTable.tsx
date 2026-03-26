import { useState } from 'react'
import { Pencil, Trash2, GripVertical, Check, X } from 'lucide-react'
import { useRulesStore } from '../../stores/rulesStore'
import { RouteBadge } from '../ui/Badge'
import { Toggle } from '../ui/Toggle'
import { ConfirmDialog } from '../ui/ConfirmDialog'
import { formatBytes } from '../../lib/format'

interface Rule {
  id: number; ruleType: string; pattern: string; action: string
  priority: number; enabled: boolean; hitCount: number; bytesSaved: number; createdAt: any
}

interface Props {
  rules: Rule[]
}

const ACTIONS = [
  { value: 'bypass', label: 'Bypass' },
  { value: 'block', label: 'Block' },
  { value: 'bypass_vps', label: 'Bypass VPS' },
  { value: 'direct', label: 'Direct' },
  { value: 'datacenter', label: 'Datacenter' },
  { value: 'residential', label: 'Residential' },
]

export function RulesTable({ rules }: Props) {
  const { updateRule, deleteRule, toggleRule } = useRulesStore()
  const [editId, setEditId] = useState<number | null>(null)
  const [editPattern, setEditPattern] = useState('')
  const [editAction, setEditAction] = useState('')
  const [deleteId, setDeleteId] = useState<number | null>(null)

  const startEdit = (rule: Rule) => {
    setEditId(rule.id)
    setEditPattern(rule.pattern)
    setEditAction(rule.action)
  }

  const saveEdit = async (rule: Rule) => {
    await updateRule(rule.id, rule.ruleType, editPattern, editAction, rule.priority, rule.enabled)
    setEditId(null)
  }

  const cancelEdit = () => setEditId(null)

  return (
    <>
      <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl overflow-hidden">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b border-[var(--color-border)]">
              <th className="w-8 px-3 py-2.5"></th>
              <th className="text-left px-3 py-2.5 text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide">Pattern</th>
              <th className="text-left px-3 py-2.5 text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide w-32">Route</th>
              <th className="text-center px-3 py-2.5 text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide w-20">Priority</th>
              <th className="text-right px-3 py-2.5 text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide w-24">Hits</th>
              <th className="text-right px-3 py-2.5 text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide w-24">Saved</th>
              <th className="text-center px-3 py-2.5 text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide w-16">On</th>
              <th className="w-20 px-3 py-2.5"></th>
            </tr>
          </thead>
          <tbody>
            {rules.map((rule) => (
              <tr key={rule.id} className="border-b border-[var(--color-border-subtle)] hover:bg-[var(--color-sidebar-hover)] transition-colors">
                <td className="px-3 py-2 text-[var(--color-text-muted)]">
                  <GripVertical size={14} className="cursor-grab" />
                </td>
                <td className="px-3 py-2">
                  {editId === rule.id ? (
                    <input
                      value={editPattern}
                      onChange={(e) => setEditPattern(e.target.value)}
                      onKeyDown={(e) => { if (e.key === 'Enter') saveEdit(rule); if (e.key === 'Escape') cancelEdit() }}
                      className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-focus)]/50 rounded px-2 py-1 text-xs text-[var(--color-text-primary)] outline-none"
                      autoFocus
                    />
                  ) : (
                    <span className="font-mono text-xs text-[var(--color-text-secondary)]">{rule.pattern}</span>
                  )}
                </td>
                <td className="px-3 py-2">
                  {editId === rule.id ? (
                    <select
                      value={editAction}
                      onChange={(e) => setEditAction(e.target.value)}
                      className="bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded px-2 py-1 text-xs text-[var(--color-text-primary)] outline-none"
                    >
                      {ACTIONS.map((a) => <option key={a.value} value={a.value}>{a.label}</option>)}
                    </select>
                  ) : (
                    <RouteBadge route={rule.action} />
                  )}
                </td>
                <td className="px-3 py-2 text-center text-xs text-[var(--color-text-muted)] tabular-nums">{rule.priority}</td>
                <td className="px-3 py-2 text-right text-xs text-[var(--color-text-muted)] tabular-nums">{rule.hitCount.toLocaleString()}</td>
                <td className="px-3 py-2 text-right text-xs text-[var(--color-text-muted)] tabular-nums">{formatBytes(rule.bytesSaved)}</td>
                <td className="px-3 py-2 text-center">
                  <Toggle checked={rule.enabled} onChange={(v) => toggleRule(rule.id, v)} />
                </td>
                <td className="px-3 py-2">
                  <div className="flex items-center justify-end gap-1">
                    {editId === rule.id ? (
                      <>
                        <button onClick={() => saveEdit(rule)} aria-label="Save" className="p-1 rounded hover:bg-[var(--color-success-bg)] text-[var(--color-success)]" title="Save">
                          <Check size={14} />
                        </button>
                        <button onClick={cancelEdit} aria-label="Cancel" className="p-1 rounded hover:bg-[var(--color-sidebar-hover)] text-[var(--color-text-muted)]" title="Cancel">
                          <X size={14} />
                        </button>
                      </>
                    ) : (
                      <>
                        <button onClick={() => startEdit(rule)} aria-label="Edit" className="p-1 rounded hover:bg-[var(--color-sidebar-hover)] text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]" title="Edit">
                          <Pencil size={14} />
                        </button>
                        <button onClick={() => setDeleteId(rule.id)} aria-label="Delete" className="p-1 rounded hover:bg-[var(--color-danger-bg)] text-[var(--color-text-muted)] hover:text-[var(--color-danger)]" title="Delete">
                          <Trash2 size={14} />
                        </button>
                      </>
                    )}
                  </div>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      <ConfirmDialog
        open={deleteId !== null}
        onClose={() => setDeleteId(null)}
        onConfirm={() => { if (deleteId !== null) deleteRule(deleteId) }}
        title="Delete Rule"
        message="Are you sure you want to delete this rule? This action cannot be undone."
        confirmText="Delete"
        destructive
      />
    </>
  )
}
