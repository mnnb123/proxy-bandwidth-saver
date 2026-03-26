import { useEffect, useState, useCallback, useMemo } from 'react'
import { Plus, Download, Upload, FlaskConical, Layers, Shield, Ban, Globe } from 'lucide-react'
import { useRulesStore } from '../stores/rulesStore'
import { RulesTable } from '../components/rules/RulesTable'
import { AddRuleForm } from '../components/rules/AddRuleForm'
import { RuleTester } from '../components/rules/RuleTester'
import { ImportExportModal } from '../components/rules/ImportExportModal'
import { EmptyState } from '../components/ui/EmptyState'
import { TableSkeleton } from '../components/ui/Skeleton'

type FilterKey = 'all' | 'bypass' | 'block' | 'bypass_vps'

const FILTERS: { key: FilterKey; label: string; icon: typeof Globe }[] = [
  { key: 'all', label: 'All', icon: Layers },
  { key: 'bypass', label: 'Bypass', icon: Globe },
  { key: 'block', label: 'Block', icon: Ban },
  { key: 'bypass_vps', label: 'Bypass VPS', icon: Shield },
]

export default function RulesPage() {
  const rules = useRulesStore((s) => s.rules)
  const loading = useRulesStore((s) => s.loading)
  const fetchRules = useRulesStore((s) => s.fetchRules)
  const [filter, setFilter] = useState<FilterKey>('all')
  const [showAdd, setShowAdd] = useState(false)
  const [showTester, setShowTester] = useState(false)
  const [importExport, setImportExport] = useState<'import' | 'export' | null>(null)

  useEffect(() => { fetchRules() }, [fetchRules])

  const filtered = useMemo(() => {
    if (filter === 'all') return rules
    // bypass also includes "direct" (legacy)
    if (filter === 'bypass') return rules.filter((r) => r.action === 'bypass' || r.action === 'direct')
    return rules.filter((r) => r.action === filter)
  }, [rules, filter])

  const counts = useMemo(() => ({
    all: rules.length,
    bypass: rules.filter((r) => r.action === 'bypass' || r.action === 'direct').length,
    block: rules.filter((r) => r.action === 'block').length,
    bypass_vps: rules.filter((r) => r.action === 'bypass_vps').length,
  }), [rules])

  const handleAdded = useCallback(() => {
    setShowAdd(false)
  }, [])

  return (
    <div className="p-6 overflow-y-auto h-full space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold text-[var(--color-text-primary)]">Domain Rules</h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowTester(!showTester)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors"
          >
            <FlaskConical size={14} /> Test
          </button>
          <button
            onClick={() => setImportExport('import')}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors"
          >
            <Upload size={14} /> Import
          </button>
          <button
            onClick={() => setImportExport('export')}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors"
          >
            <Download size={14} /> Export
          </button>
          <button
            onClick={() => setShowAdd(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors"
          >
            <Plus size={14} /> Add Rules
          </button>
        </div>
      </div>

      {/* Rule Tester */}
      {showTester && <RuleTester />}

      {/* Stats Cards */}
      <div className="grid grid-cols-4 gap-3">
        {FILTERS.map((f) => {
          const Icon = f.icon
          return (
            <button
              key={f.key}
              onClick={() => setFilter(f.key)}
              className={`bg-[var(--color-bg-surface)] border rounded-lg px-4 py-3 text-left transition-all ${
                filter === f.key
                  ? 'border-[var(--color-primary)] ring-1 ring-[var(--color-primary)]/30'
                  : 'border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)]'
              }`}
            >
              <div className="flex items-center gap-1.5">
                <Icon size={12} className="text-[var(--color-text-muted)]" />
                <span className="text-xs text-[var(--color-text-muted)]">{f.label}</span>
              </div>
              <div className="text-xl font-bold text-[var(--color-text-primary)] tabular-nums mt-1">{counts[f.key]}</div>
            </button>
          )
        })}
      </div>

      {/* Table */}
      {loading ? (
        <TableSkeleton rows={5} cols={6} />
      ) : filtered.length === 0 ? (
        <EmptyState
          icon={Layers}
          title="No rules"
          description="Add domain rules to control traffic routing"
          action={
            <button
              onClick={() => setShowAdd(true)}
              className="px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors"
            >
              Add Rules
            </button>
          }
        />
      ) : (
        <RulesTable rules={filtered} />
      )}

      {/* Add Rules Modal */}
      {showAdd && <AddRuleForm onClose={handleAdded} />}

      {/* Import/Export Modal */}
      {importExport && (
        <ImportExportModal
          mode={importExport}
          onClose={() => setImportExport(null)}
        />
      )}
    </div>
  )
}
