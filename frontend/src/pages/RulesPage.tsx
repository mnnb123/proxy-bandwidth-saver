import { useEffect, useState, useCallback, useMemo } from 'react'
import { Plus, Download, Upload, FlaskConical, Layers } from 'lucide-react'
import { useRulesStore } from '../stores/rulesStore'
import { RulesTable } from '../components/rules/RulesTable'
import { AddRuleForm } from '../components/rules/AddRuleForm'
import { RuleTester } from '../components/rules/RuleTester'
import { ImportExportModal } from '../components/rules/ImportExportModal'
import { EmptyState } from '../components/ui/EmptyState'
import { TableSkeleton } from '../components/ui/Skeleton'

const TABS = [
  { key: 'domain', label: 'Domain Rules' },
  { key: 'content_type', label: 'Content-Type' },
  { key: 'url_pattern', label: 'URL Pattern' },
] as const

type TabKey = typeof TABS[number]['key']

export default function RulesPage() {
  const rules = useRulesStore((s) => s.rules)
  const loading = useRulesStore((s) => s.loading)
  const fetchRules = useRulesStore((s) => s.fetchRules)
  const [tab, setTab] = useState<TabKey>('domain')
  const [showAdd, setShowAdd] = useState(false)
  const [showTester, setShowTester] = useState(false)
  const [importExport, setImportExport] = useState<'import' | 'export' | null>(null)

  useEffect(() => { fetchRules() }, [fetchRules])

  const filtered = useMemo(() => rules.filter((r) => r.ruleType === tab), [rules, tab])
  const tabCounts = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const r of rules) counts[r.ruleType] = (counts[r.ruleType] || 0) + 1
    return counts
  }, [rules])

  const handleAdded = useCallback(() => {
    setShowAdd(false)
  }, [])

  return (
    <div className="p-6 overflow-y-auto h-full space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold text-[var(--color-text-primary)]">Traffic Rules</h1>
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
            onClick={() => setShowAdd(!showAdd)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors"
          >
            <Plus size={14} /> Add Rule
          </button>
        </div>
      </div>

      {/* Rule Tester */}
      {showTester && <RuleTester />}

      {/* Add Rule Form */}
      {showAdd && <AddRuleForm defaultType={tab} onDone={handleAdded} />}

      {/* Tabs */}
      <div className="flex border-b border-[var(--color-border)]">
        {TABS.map((t) => {
          const count = tabCounts[t.key] || 0
          return (
            <button
              key={t.key}
              onClick={() => setTab(t.key)}
              className={`px-4 py-2.5 text-xs font-medium border-b-2 transition-colors ${
                tab === t.key
                  ? 'border-[var(--color-primary)] text-[var(--color-primary)]'
                  : 'border-transparent text-[var(--color-text-muted)] hover:text-[var(--color-text-secondary)]'
              }`}
            >
              {t.label}
              <span className="ml-2 px-1.5 py-0.5 rounded-full bg-[var(--color-bg-elevated)] text-[10px]">{count}</span>
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
          title="No rules in this category"
          description="Add a rule to start classifying traffic"
          action={
            <button
              onClick={() => setShowAdd(true)}
              className="px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors"
            >
              Add Rule
            </button>
          }
        />
      ) : (
        <RulesTable rules={filtered} />
      )}

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
