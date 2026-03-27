import { useEffect, useState, useCallback, useMemo, memo } from 'react'
import { Globe, Trash2, Clock, Settings2, Check, Copy, Search, ArrowUpDown, ChevronUp, ChevronDown, Save } from 'lucide-react'
import { GetDomainStats, ClearDomainStats, GetOutputProxies, GetSettings, UpdateSetting } from '../lib/api'
import { useToastStore } from '../stores/toastStore'
import { formatBytes, copyToClipboard } from '../lib/format'
import { EmptyState } from '../components/ui/EmptyState'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'

interface DomainStat {
  domain: string
  totalBytes: number
  requests: number
  route: string
  cacheHitPct: number
  proxyId: number
  lastSeen: string
}

interface OutputProxy {
  proxyId: number
  localAddr: string
  localPort: number
  protocol: string
  upstream: string
  type: string
}

const AUTO_CLEAR_OPTIONS = [
  { value: 0, label: 'Tắt' },
  { value: 1, label: '1 phút' },
  { value: 5, label: '5 phút' },
  { value: 10, label: '10 phút' },
  { value: 30, label: '30 phút' },
  { value: 60, label: '1 giờ' },
]

const SHOW_OPTIONS = [25, 50, 100, 200]

const ROUTE_CONFIG: Record<string, { label: string; color: string; bg: string }> = {
  residential: { label: 'PROXY', color: 'text-red-600 dark:text-red-400', bg: 'bg-red-50 dark:bg-red-950/30' },
  datacenter: { label: 'DATACENTER', color: 'text-blue-600 dark:text-blue-400', bg: 'bg-blue-50 dark:bg-blue-950/30' },
  direct: { label: 'BYPASS VPS', color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-950/30' },
  bypass: { label: 'BYPASS', color: 'text-amber-600 dark:text-amber-400', bg: 'bg-yellow-50 dark:bg-yellow-950/20' },
  bypass_vps: { label: 'BYPASS VPS', color: 'text-emerald-600 dark:text-emerald-400', bg: 'bg-emerald-50 dark:bg-emerald-950/30' },
  block: { label: 'BLOCK', color: 'text-orange-600 dark:text-orange-400', bg: 'bg-orange-50 dark:bg-orange-950/30' },
}

interface DomainRowProps {
  stat: DomainStat
  port: number | string
  copiedDomain: string
  copyDomain: (domain: string) => void
}

const DomainRow = memo(function DomainRow({ stat, port, copiedDomain, copyDomain }: DomainRowProps) {
  const cfg = ROUTE_CONFIG[stat.route]
  const bg = cfg?.bg || ''
  const label = cfg?.label || stat.route.toUpperCase()
  const color = cfg?.color || 'text-[var(--color-text-muted)]'
  return (
    <tr
      className={`border-b border-[var(--color-border-subtle)] ${bg} hover:brightness-95 dark:hover:brightness-110 transition-all`}
    >
      <td className="px-4 py-2 text-xs font-mono font-bold text-[var(--color-text-primary)]">
        {port || 'main'}
      </td>
      <td className="px-4 py-2">
        <span className="inline-flex items-center gap-1.5">
          <span className="text-xs font-mono text-[var(--color-text-primary)]">{stat.domain}</span>
          <button
            onClick={() => copyDomain(stat.domain)}
            className="p-0.5 rounded hover:bg-black/10 dark:hover:bg-white/10 text-[var(--color-text-muted)] opacity-60 hover:opacity-100 transition-opacity"
            title="Copy domain"
          >
            {copiedDomain === stat.domain
              ? <Check size={12} className="text-emerald-500" />
              : <Copy size={12} />
            }
          </button>
        </span>
      </td>
      <td className="px-4 py-2">
        <span className={`text-xs font-bold ${color}`}>{label}</span>
      </td>
      <td className="px-4 py-2 text-right text-xs font-mono font-semibold text-[var(--color-text-primary)]">
        {formatBytes(stat.totalBytes)}
      </td>
      <td className="px-4 py-2 text-right text-xs font-mono text-[var(--color-text-muted)]">
        {stat.lastSeen || '-'}
      </td>
    </tr>
  )
})

export default function DomainsPage() {
  const [stats, setStats] = useState<DomainStat[]>([])
  const [loading, setLoading] = useState(true)
  const [showClearConfirm, setShowClearConfirm] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [autoClearMinutes, setAutoClearMinutes] = useState(0)
  const [pendingAutoClear, setPendingAutoClear] = useState<number | null>(null)
  const [savingAutoClear, setSavingAutoClear] = useState(false)
  const [outputProxies, setOutputProxies] = useState<OutputProxy[]>([])
  const [search, setSearch] = useState('')
  const [showCount, setShowCount] = useState(50)
  const [sortField, setSortField] = useState<'totalBytes' | 'lastSeen' | 'route' | 'domain' | 'port'>('totalBytes')
  const [sortDir, setSortDir] = useState<'asc' | 'desc'>('desc')
  const [copiedDomain, setCopiedDomain] = useState('')

  const toggleSort = (field: typeof sortField) => {
    if (sortField === field) {
      setSortDir(sortDir === 'desc' ? 'asc' : 'desc')
    } else {
      setSortField(field)
      setSortDir(field === 'domain' || field === 'route' ? 'asc' : 'desc')
    }
  }

  // Build port lookup map
  const portMap = useMemo(() => {
    const map: Record<number, number> = {}
    for (const p of outputProxies) {
      if (p.protocol === 'http') {
        map[p.proxyId] = p.localPort
      }
    }
    return map
  }, [outputProxies])

  // Load output proxies + settings
  useEffect(() => {
    GetOutputProxies().then((p: OutputProxy[]) => setOutputProxies(p || [])).catch(() => {})
    GetSettings().then((s: Record<string, string>) => {
      const val = parseInt(s?.domain_report_clear_minutes || '0', 10)
      setAutoClearMinutes(isNaN(val) ? 0 : val)
    }).catch(() => {})
  }, [])

  const fetchStats = useCallback(async (silent = false) => {
    if (!silent) setLoading(true)
    try {
      const data = await GetDomainStats('24h', 0)
      setStats(data || [])
    } catch {
      if (!silent) setStats([])
    }
    if (!silent) setLoading(false)
  }, [])

  // Initial load
  useEffect(() => { fetchStats() }, [fetchStats])

  // Real-time refresh every 1 second
  useEffect(() => {
    const interval = setInterval(() => fetchStats(true), 1000)
    return () => clearInterval(interval)
  }, [fetchStats])

  // Filter + sort
  const filtered = useMemo(() => {
    let result = stats
    if (search.trim()) {
      const searchQuery = search.toLowerCase()
      result = result.filter((s) => s.domain.toLowerCase().includes(searchQuery))
    }
    const dir = sortDir === 'asc' ? 1 : -1
    result = [...result].sort((a, b) => {
      switch (sortField) {
        case 'totalBytes': return (a.totalBytes - b.totalBytes) * dir
        case 'lastSeen': return (a.lastSeen || '').localeCompare(b.lastSeen || '') * dir
        case 'route': {
          const labelA = (ROUTE_CONFIG[a.route]?.label) || a.route
          const labelB = (ROUTE_CONFIG[b.route]?.label) || b.route
          return labelA.localeCompare(labelB) * dir
        }
        case 'domain': return a.domain.localeCompare(b.domain) * dir
        case 'port': return ((portMap[a.proxyId] || a.proxyId) - (portMap[b.proxyId] || b.proxyId)) * dir
        default: return 0
      }
    })
    return result.slice(0, showCount)
  }, [stats, search, showCount, sortField, sortDir, portMap])

  const handleClear = async () => {
    await ClearDomainStats()
    setStats([])
    setShowClearConfirm(false)
  }

  const handleAutoClearSelect = (minutes: number) => {
    setPendingAutoClear(minutes)
  }

  const handleAutoClearSave = async () => {
    if (pendingAutoClear === null) return
    setSavingAutoClear(true)
    try {
      await UpdateSetting('domain_report_clear_minutes', String(pendingAutoClear))
      setAutoClearMinutes(pendingAutoClear)
      setPendingAutoClear(null)
      useToastStore.getState().addToast('success', pendingAutoClear === 0 ? 'Auto clear disabled' : `Auto clear set to ${pendingAutoClear} minutes`)
    } catch (e) {
      useToastStore.getState().addToast('error', `Failed to save: ${e}`)
    } finally {
      setSavingAutoClear(false)
    }
  }

  const currentAutoClear = pendingAutoClear !== null ? pendingAutoClear : autoClearMinutes
  const hasUnsavedAutoClear = pendingAutoClear !== null && pendingAutoClear !== autoClearMinutes

  const copyDomain = (domain: string) => {
    copyToClipboard(domain)
    setCopiedDomain(domain)
    setTimeout(() => setCopiedDomain(''), 1500)
  }

  return (
    <div className="p-6 overflow-y-auto h-full space-y-3">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold text-[var(--color-text-primary)]">Domain Report</h1>
        <div className="flex items-center gap-2">
          <button
            onClick={() => setShowSettings(!showSettings)}
            className={`flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg border transition-colors ${
              autoClearMinutes > 0
                ? 'bg-emerald-500/10 text-emerald-500 border-emerald-500/30'
                : 'bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)]'
            }`}
          >
            <Clock size={14} />
            {autoClearMinutes > 0 ? `Auto Clear: ${autoClearMinutes}m` : 'Auto Clear'}
          </button>
          {stats.length > 0 && (
            <button
              onClick={() => setShowClearConfirm(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg bg-red-500/10 text-red-500 hover:bg-red-500/20 border border-red-500/30 transition-colors"
            >
              <Trash2 size={14} /> Clear Data
            </button>
          )}
        </div>
      </div>

      {/* Auto Clear Settings */}
      {showSettings && (
        <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-lg p-4">
          <div className="flex items-center justify-between mb-2">
            <div className="flex items-center gap-2">
              <Settings2 size={14} className="text-[var(--color-text-muted)]" />
              <span className="text-xs font-medium text-[var(--color-text-primary)]">Auto Clear Domain Data</span>
            </div>
            <button
              onClick={handleAutoClearSave}
              disabled={!hasUnsavedAutoClear || savingAutoClear}
              className={`flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg border transition-colors ${
                hasUnsavedAutoClear
                  ? 'bg-[var(--color-primary)] text-white border-[var(--color-primary)] hover:bg-[var(--color-primary-hover)]'
                  : 'bg-[var(--color-bg-elevated)] text-[var(--color-text-muted)] border-[var(--color-border)] opacity-50 cursor-not-allowed'
              }`}
            >
              <Save size={12} />
              {savingAutoClear ? 'Saving...' : 'Save'}
            </button>
          </div>
          <p className="text-[11px] text-[var(--color-text-muted)] mb-3">
            Tự động xóa dữ liệu cũ hơn thời gian đã chọn. Giúp giảm RAM và database. Chọn thời gian rồi nhấn Save.
          </p>
          <div className="flex flex-wrap gap-2">
            {AUTO_CLEAR_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                onClick={() => handleAutoClearSelect(opt.value)}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg border transition-colors ${
                  currentAutoClear === opt.value
                    ? 'bg-[var(--color-primary)] text-white border-[var(--color-primary)]'
                    : 'bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)]'
                }`}
              >
                {currentAutoClear === opt.value && <Check size={12} />}
                {opt.label}
              </button>
            ))}
          </div>
          {hasUnsavedAutoClear && (
            <p className="text-[11px] text-amber-500 mt-2">
              Chưa lưu — nhấn Save để áp dụng.
            </p>
          )}
        </div>
      )}

      {/* Toolbar: Show entries + Sort + Search */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="flex items-center gap-2 text-xs text-[var(--color-text-muted)]">
            <span>Show</span>
            <select
              value={showCount}
              onChange={(e) => setShowCount(Number(e.target.value))}
              className="bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded px-2 py-1 text-xs text-[var(--color-text-primary)] outline-none"
            >
              {SHOW_OPTIONS.map((n) => (
                <option key={n} value={n}>{n}</option>
              ))}
            </select>
            <span>entries</span>
          </div>
          <span className="text-[10px] text-[var(--color-text-muted)]">Click column header to sort</span>
        </div>
        <div className="relative">
          <Search size={14} className="absolute left-2.5 top-1/2 -translate-y-1/2 text-[var(--color-text-muted)]" />
          <input
            type="text"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Search domain..."
            className="bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-lg pl-8 pr-3 py-1.5 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] w-56"
          />
        </div>
      </div>

      {/* Live indicator */}
      <div className="flex items-center gap-2">
        <span className="relative flex h-2 w-2">
          <span className="animate-ping absolute inline-flex h-full w-full rounded-full bg-emerald-400 opacity-75"></span>
          <span className="relative inline-flex rounded-full h-2 w-2 bg-emerald-500"></span>
        </span>
        <span className="text-[11px] text-[var(--color-text-muted)]">
          Real-time · {stats.length} domains · {formatBytes(stats.reduce((s, d) => s + d.totalBytes, 0))} total
        </span>
      </div>

      {/* Table */}
      {loading ? (
        <div className="text-center py-8 text-sm text-[var(--color-text-muted)]">Loading...</div>
      ) : stats.length === 0 ? (
        <EmptyState
          icon={Globe}
          title="Chưa có dữ liệu"
          description="Hãy start proxy và gửi traffic để xem thống kê domain"
        />
      ) : (
        <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl overflow-hidden">
          <div className="max-h-[calc(100vh-300px)] overflow-y-auto">
            <table className="w-full">
              <thead className="sticky top-0 bg-[var(--color-bg-elevated)] z-10">
                <tr className="border-b border-[var(--color-border)]">
                  {([
                    { key: 'port' as const, label: 'Port', align: 'left', width: 'w-20' },
                    { key: 'domain' as const, label: 'HostName', align: 'left', width: '' },
                    { key: 'route' as const, label: 'Status', align: 'left', width: 'w-32' },
                    { key: 'totalBytes' as const, label: 'Bandwidth', align: 'right', width: 'w-28' },
                    { key: 'lastSeen' as const, label: 'Created', align: 'right', width: 'w-44' },
                  ]).map((col) => (
                    <th
                      key={col.key}
                      onClick={() => toggleSort(col.key)}
                      className={`${col.align === 'right' ? 'text-right' : 'text-left'} px-4 py-2.5 text-xs font-semibold text-[var(--color-text-primary)] ${col.width} cursor-pointer select-none hover:bg-[var(--color-sidebar-hover)] transition-colors`}
                    >
                      <span className="inline-flex items-center gap-1">
                        {col.label}
                        {sortField === col.key ? (
                          sortDir === 'asc' ? <ChevronUp size={12} /> : <ChevronDown size={12} />
                        ) : (
                          <ArrowUpDown size={10} className="opacity-30" />
                        )}
                      </span>
                    </th>
                  ))}
                </tr>
              </thead>
              <tbody>
                {filtered.map((stat, i) => (
                  <DomainRow
                    key={`${stat.domain}-${stat.proxyId}-${i}`}
                    stat={stat}
                    port={portMap[stat.proxyId] || stat.proxyId}
                    copiedDomain={copiedDomain}
                    copyDomain={copyDomain}
                  />
                ))}
              </tbody>
            </table>
          </div>
        </div>
      )}

      {/* Footer */}
      {filtered.length > 0 && (
        <div className="text-xs text-[var(--color-text-muted)]">
          Showing {filtered.length} of {stats.length} entries
          {search && ` (filtered from ${stats.length} total)`}
        </div>
      )}

      {/* Clear Confirm */}
      <ConfirmDialog
        open={showClearConfirm}
        onClose={() => setShowClearConfirm(false)}
        onConfirm={handleClear}
        title="Clear Domain Data"
        message="Xóa toàn bộ dữ liệu bandwidth domain report? Thao tác không thể hoàn tác."
        confirmText="Clear Data"
        destructive
      />
    </div>
  )
}
