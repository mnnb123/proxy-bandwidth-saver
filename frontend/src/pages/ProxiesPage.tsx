import { useEffect, useState, useMemo, useCallback } from 'react'
import { Plus, Upload, Globe, Trash2, Wifi, WifiOff, Copy, Check, ExternalLink } from 'lucide-react'
import { useProxiesStore } from '../stores/proxiesStore'
import { EmptyState } from '../components/ui/EmptyState'
import { TableSkeleton } from '../components/ui/Skeleton'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'
import { TypeBadge } from '../components/ui/Badge'
import { AddProxyModal } from '../components/proxies/AddProxyModal'
import { BulkImportModal } from '../components/proxies/BulkImportModal'

export default function ProxiesPage() {
  const proxies = useProxiesStore((s) => s.proxies)
  const outputProxies = useProxiesStore((s) => s.outputProxies)
  const loading = useProxiesStore((s) => s.loading)
  const fetchProxies = useProxiesStore((s) => s.fetchProxies)
  const fetchOutputProxies = useProxiesStore((s) => s.fetchOutputProxies)
  const addProxy = useProxiesStore((s) => s.addProxy)
  const deleteProxy = useProxiesStore((s) => s.deleteProxy)
  const importProxies = useProxiesStore((s) => s.importProxies)
  const [showAdd, setShowAdd] = useState(false)
  const [showImport, setShowImport] = useState(false)
  const [deleteId, setDeleteId] = useState<number | null>(null)
  const [copied, setCopied] = useState(false)

  useEffect(() => {
    fetchProxies()
    fetchOutputProxies()
  }, [fetchProxies, fetchOutputProxies])

  const stats = useMemo(() => [
    { label: 'Total', value: proxies.length },
    { label: 'Residential', value: proxies.filter((p) => p.category === 'residential').length },
    { label: 'Datacenter', value: proxies.filter((p) => p.category === 'datacenter').length },
    { label: 'Output Ports', value: outputProxies.length },
  ], [proxies, outputProxies.length])

  const copyOutputList = useCallback(async () => {
    const text = outputProxies.map((o) => o.localAddr).join('\n')
    await navigator.clipboard.writeText(text)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }, [outputProxies])

  return (
    <div className="p-6 overflow-y-auto h-full space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold text-[var(--color-text-primary)]">Proxy Pool</h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowImport(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors"
          >
            <Upload size={14} /> Bulk Import
          </button>
          <button
            onClick={() => setShowAdd(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors"
          >
            <Plus size={14} /> Add Proxy
          </button>
        </div>
      </div>

      {/* Stats Bar */}
      <div className="grid grid-cols-4 gap-3">
        {stats.map((s, i) => (
          <div key={s.label} className={`bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-lg px-4 py-3 transition-transform duration-200 hover:-translate-y-0.5 hover:shadow-md animate-fade-in-up stagger-${i + 1}`}>
            <div className="text-xs text-[var(--color-text-muted)]">{s.label}</div>
            <div className="text-xl font-bold text-[var(--color-text-primary)] tabular-nums">{s.value}</div>
          </div>
        ))}
      </div>

      {/* Two-column: Input Proxies + Output Proxies */}
      <div className="grid grid-cols-2 gap-4">
        {/* Input Proxies Table */}
        <div>
          <h2 className="text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide mb-2">Input Proxies</h2>
          {loading ? (
            <TableSkeleton rows={3} cols={4} />
          ) : proxies.length === 0 ? (
            <EmptyState
              icon={Globe}
              title="No proxies configured"
              description="Add proxies to generate output ports"
              action={
                <button
                  onClick={() => setShowImport(true)}
                  className="px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-primary)] text-white hover:bg-[var(--color-primary-hover)] transition-colors"
                >
                  Import Proxies
                </button>
              }
            />
          ) : (
            <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl overflow-hidden max-h-[500px] overflow-y-auto">
              <table className="w-full text-sm">
                <thead className="sticky top-0 bg-[var(--color-bg-elevated)]">
                  <tr className="border-b border-[var(--color-border)]">
                    <th className="w-8 px-2 py-2 text-xs font-medium text-[var(--color-text-muted)]"></th>
                    <th className="text-left px-2 py-2 text-xs font-medium text-[var(--color-text-muted)]">Address</th>
                    <th className="text-left px-2 py-2 text-xs font-medium text-[var(--color-text-muted)] w-16">Type</th>
                    <th className="w-10 px-2 py-2"></th>
                  </tr>
                </thead>
                <tbody>
                  {proxies.map((proxy) => (
                    <tr key={proxy.id} className="border-b border-[var(--color-border-subtle)] hover:bg-[var(--color-sidebar-hover)] transition-colors">
                      <td className="px-2 py-1.5 text-center">
                        {proxy.failCount === 0 ? (
                          <Wifi size={12} className="mx-auto text-[var(--color-success)]" />
                        ) : (
                          <WifiOff size={12} className="mx-auto text-[var(--color-danger)]" />
                        )}
                      </td>
                      <td className="px-2 py-1.5">
                        <span className="font-mono text-[11px] text-[var(--color-text-secondary)]">{proxy.address}</span>
                        {proxy.username && (
                          <span className="ml-1 text-[10px] text-[var(--color-text-muted)]">({proxy.username}:****)</span>
                        )}
                      </td>
                      <td className="px-2 py-1.5">
                        <TypeBadge type={proxy.type || 'http'} />
                      </td>
                      <td className="px-2 py-1.5">
                        <button
                          onClick={() => setDeleteId(proxy.id)}
                          aria-label="Remove proxy"
                          className="p-0.5 rounded hover:bg-[var(--color-danger-bg)] text-[var(--color-text-muted)] hover:text-[var(--color-danger)]"
                        >
                          <Trash2 size={12} />
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Output Proxies Panel */}
        <div>
          <div className="flex items-center justify-between mb-2">
            <h2 className="text-xs font-medium text-[var(--color-text-muted)] uppercase tracking-wide">
              <ExternalLink size={12} className="inline mr-1" />
              List Output Proxies
            </h2>
            {outputProxies.length > 0 && (
              <button
                onClick={copyOutputList}
                className="flex items-center gap-1 px-2 py-1 text-[11px] rounded bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors"
              >
                {copied ? <Check size={11} /> : <Copy size={11} />}
                {copied ? 'Copied!' : 'Copy All'}
              </button>
            )}
          </div>
          <div className="text-[11px] text-[var(--color-text-muted)] mb-2">
            HTTP Proxy format: host:port (no auth needed)
          </div>
          {outputProxies.length === 0 ? (
            <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl p-8 text-center">
              <p className="text-xs text-[var(--color-text-muted)]">Add proxies to see output ports here</p>
            </div>
          ) : (
            <div className="bg-[var(--color-bg-base)] border border-[var(--color-success)]/30 rounded-xl overflow-hidden">
              <div className="max-h-[500px] overflow-y-auto">
                <div className="p-3 space-y-0.5">
                  {outputProxies.map((op) => (
                    <div
                      key={op.proxyId}
                      className="flex items-center justify-between px-3 py-1.5 rounded hover:bg-[var(--color-sidebar-hover)] transition-colors group"
                    >
                      <span className="font-mono text-xs text-[var(--color-success)]">{op.localAddr}</span>
                      <span className="text-[10px] text-[var(--color-text-muted)] group-hover:text-[var(--color-text-secondary)] transition-colors">
                        → {op.upstream} ({op.type})
                      </span>
                    </div>
                  ))}
                </div>
              </div>
              <div className="border-t border-[var(--color-border)] px-3 py-2 bg-[var(--color-bg-surface)]">
                <span className="text-[11px] text-[var(--color-text-muted)]">
                  {outputProxies.length} output ports | Base port: 30000
                </span>
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Add Proxy Modal */}
      {showAdd && <AddProxyModal onClose={() => setShowAdd(false)} onAdd={addProxy} />}

      {/* Bulk Import Modal */}
      {showImport && <BulkImportModal onClose={() => setShowImport(false)} onImport={importProxies} />}

      {/* Delete Confirm */}
      <ConfirmDialog
        open={deleteId !== null}
        onClose={() => setDeleteId(null)}
        onConfirm={() => { if (deleteId !== null) deleteProxy(deleteId) }}
        title="Remove Proxy"
        message="Are you sure you want to remove this proxy from the pool?"
        confirmText="Remove"
        destructive
      />
    </div>
  )
}
