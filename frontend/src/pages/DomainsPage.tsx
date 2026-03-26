import { useEffect, useState, useCallback, useRef, useMemo } from 'react'
import { Globe, RefreshCw, ArrowUpDown, Trash2, Clock, Settings2, Check } from 'lucide-react'
import { GetDomainStats, ClearDomainStats, GetOutputProxies, GetSettings, UpdateSetting } from '../lib/api'
import { formatBytes } from '../lib/format'
import { TableSkeleton } from '../components/ui/Skeleton'
import { EmptyState } from '../components/ui/EmptyState'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'

interface DomainStat {
  domain: string
  totalBytes: number
  requests: number
  route: string
  cacheHitPct: number
  proxyId: number
}

interface OutputProxy {
  proxyId: number
  localAddr: string
  localPort: number
  protocol: string
  upstream: string
  type: string
}

const PERIODS = [
  { value: '1h', label: '1 giờ' },
  { value: '24h', label: '24 giờ' },
  { value: '7d', label: '7 ngày' },
  { value: '30d', label: '30 ngày' },
]

const AUTO_CLEAR_OPTIONS = [
  { value: 0, label: 'Tắt' },
  { value: 1, label: '1 phút' },
  { value: 5, label: '5 phút' },
  { value: 10, label: '10 phút' },
  { value: 30, label: '30 phút' },
  { value: 60, label: '1 giờ' },
]

const routeColors: Record<string, { bg: string; text: string; label: string }> = {
  residential: { bg: 'bg-red-500/10', text: 'text-red-400', label: 'Proxy' },
  datacenter: { bg: 'bg-blue-500/10', text: 'text-blue-400', label: 'Datacenter' },
  direct: { bg: 'bg-emerald-500/10', text: 'text-emerald-400', label: 'Bypass VPS' },
  bypass: { bg: 'bg-cyan-500/10', text: 'text-cyan-400', label: 'Bypass' },
  bypass_vps: { bg: 'bg-emerald-500/10', text: 'text-emerald-400', label: 'Bypass VPS' },
  block: { bg: 'bg-orange-500/10', text: 'text-orange-400', label: 'Block' },
}

export default function DomainsPage() {
  const [stats, setStats] = useState<DomainStat[]>([])
  const [loading, setLoading] = useState(true)
  const [period, setPeriod] = useState('24h')
  const [sortBy, setSortBy] = useState<'totalBytes' | 'requests'>('totalBytes')
  const [showClearConfirm, setShowClearConfirm] = useState(false)
  const [showSettings, setShowSettings] = useState(false)
  const [autoClearMinutes, setAutoClearMinutes] = useState(0)
  const [outputProxies, setOutputProxies] = useState<OutputProxy[]>([])
  const [selectedProxyId, setSelectedProxyId] = useState(0) // 0 = all

  // Load output proxies and auto-clear setting
  useEffect(() => {
    GetOutputProxies().then((p: OutputProxy[]) => {
      setOutputProxies(p || [])
    }).catch(() => {})
    GetSettings().then((s: Record<string, string>) => {
      const val = parseInt(s?.domain_report_clear_minutes || '0', 10)
      setAutoClearMinutes(isNaN(val) ? 0 : val)
    }).catch(() => {})
  }, [])

  // Get unique proxy IDs from output proxies (HTTP only, skip SOCKS5 duplicates)
  const proxyTabs = useMemo(() => {
    const seen = new Set<number>()
    const tabs: { id: number; label: string; port: number }[] = []
    for (const p of outputProxies) {
      if (p.protocol === 'http' && !seen.has(p.proxyId)) {
        seen.add(p.proxyId)
        tabs.push({ id: p.proxyId, label: `Port ${p.localPort}`, port: p.localPort })
      }
    }
    return tabs.sort((a, b) => a.port - b.port)
  }, [outputProxies])

  const fetchStats = useCallback(async (p: string, proxyId: number, silent = false) => {
    if (!silent) setLoading(true)
    try {
      const data = await GetDomainStats(p, proxyId)
      setStats(data || [])
    } catch {
      if (!silent) setStats([])
    }
    if (!silent) setLoading(false)
  }, [])

  // Initial load
  useEffect(() => {
    fetchStats(period, selectedProxyId)
  }, [period, selectedProxyId, fetchStats])

  // Auto-refresh every 3 seconds
  const periodRef = useRef(period)
  const proxyIdRef = useRef(selectedProxyId)
  periodRef.current = period
  proxyIdRef.current = selectedProxyId
  useEffect(() => {
    const interval = setInterval(() => fetchStats(periodRef.current, proxyIdRef.current, true), 3000)
    return () => clearInterval(interval)
  }, [fetchStats])

  const sorted = [...stats].sort((a, b) => b[sortBy] - a[sortBy])
  const totalBytes = stats.reduce((sum, s) => sum + s.totalBytes, 0)
  const totalRequests = stats.reduce((sum, s) => sum + s.requests, 0)

  // Count by route
  const routeCounts = useMemo(() => {
    const counts: Record<string, number> = {}
    for (const s of stats) {
      counts[s.route] = (counts[s.route] || 0) + 1
    }
    return counts
  }, [stats])

  const handleClear = async () => {
    await ClearDomainStats()
    setStats([])
    setShowClearConfirm(false)
  }

  const handleAutoClearChange = async (minutes: number) => {
    setAutoClearMinutes(minutes)
    await UpdateSetting('domain_report_clear_minutes', String(minutes))
  }

  return (
    <div className="p-6 overflow-y-auto h-full space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold text-[var(--color-text-primary)]">Domain Report</h1>
        <div className="flex items-center gap-2">
          {/* Auto Clear Setting */}
          <button
            onClick={() => setShowSettings(!showSettings)}
            className={`flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] border transition-colors ${
              autoClearMinutes > 0
                ? 'bg-[var(--color-success)]/10 text-[var(--color-success)] border-[var(--color-success)]/30'
                : 'bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)]'
            }`}
          >
            <Clock size={14} />
            {autoClearMinutes > 0 ? `Auto Clear: ${autoClearMinutes}m` : 'Auto Clear'}
          </button>
          {/* Clear All */}
          {stats.length > 0 && (
            <button
              onClick={() => setShowClearConfirm(true)}
              className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-danger-bg)] text-[var(--color-danger)] hover:bg-[var(--color-danger)]/20 border border-[var(--color-danger)]/30 transition-colors"
            >
              <Trash2 size={14} /> Clear Data
            </button>
          )}
          {/* Period Selector */}
          {PERIODS.map((p) => (
            <button
              key={p.value}
              onClick={() => setPeriod(p.value)}
              className={`px-3 py-1.5 text-xs rounded-[var(--radius-lg)] border transition-colors ${
                period === p.value
                  ? 'bg-[var(--color-primary)] text-white border-[var(--color-primary)]'
                  : 'bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)]'
              }`}
            >
              {p.label}
            </button>
          ))}
          <button
            onClick={() => fetchStats(period, selectedProxyId)}
            className="p-1.5 rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)] transition-colors"
            aria-label="Refresh"
          >
            <RefreshCw size={14} />
          </button>
        </div>
      </div>

      {/* Auto Clear Settings Dropdown */}
      {showSettings && (
        <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-lg p-4">
          <div className="flex items-center gap-2 mb-3">
            <Settings2 size={14} className="text-[var(--color-text-muted)]" />
            <span className="text-xs font-medium text-[var(--color-text-primary)]">Auto Clear Domain Data</span>
          </div>
          <p className="text-[11px] text-[var(--color-text-muted)] mb-3">
            Tự động xóa dữ liệu domain report cũ hơn thời gian đã chọn. Giúp giảm RAM và dung lượng database.
          </p>
          <div className="flex flex-wrap gap-2">
            {AUTO_CLEAR_OPTIONS.map((opt) => (
              <button
                key={opt.value}
                onClick={() => handleAutoClearChange(opt.value)}
                className={`flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-lg border transition-colors ${
                  autoClearMinutes === opt.value
                    ? 'bg-[var(--color-primary)] text-white border-[var(--color-primary)]'
                    : 'bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)]'
                }`}
              >
                {autoClearMinutes === opt.value && <Check size={12} />}
                {opt.label}
              </button>
            ))}
          </div>
        </div>
      )}

      {/* Port Tabs */}
      {proxyTabs.length > 0 && (
        <div className="flex items-center gap-1 bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-lg p-1">
          <button
            onClick={() => setSelectedProxyId(0)}
            className={`px-3 py-1.5 text-xs rounded-md transition-colors ${
              selectedProxyId === 0
                ? 'bg-[var(--color-primary)] text-white'
                : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)]'
            }`}
          >
            All Ports
          </button>
          {proxyTabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setSelectedProxyId(tab.id)}
              className={`px-3 py-1.5 text-xs rounded-md font-mono transition-colors ${
                selectedProxyId === tab.id
                  ? 'bg-[var(--color-primary)] text-white'
                  : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)]'
              }`}
            >
              {tab.label}
            </button>
          ))}
        </div>
      )}

      {/* Summary */}
      <div className="grid grid-cols-3 gap-3">
        <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-lg px-4 py-3">
          <div className="text-xs text-[var(--color-text-muted)]">Domains</div>
          <div className="text-xl font-bold text-[var(--color-text-primary)] tabular-nums">{stats.length}</div>
          {Object.keys(routeCounts).length > 0 && (
            <div className="flex flex-wrap gap-1.5 mt-1.5">
              {Object.entries(routeCounts).map(([route, count]) => {
                const rc = routeColors[route] || { bg: 'bg-gray-500/10', text: 'text-gray-400', label: route }
                return (
                  <span key={route} className={`text-[10px] px-1.5 py-0.5 rounded ${rc.bg} ${rc.text}`}>
                    {rc.label}: {count}
                  </span>
                )
              })}
            </div>
          )}
        </div>
        <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-lg px-4 py-3">
          <div className="text-xs text-[var(--color-text-muted)]">Total Bandwidth</div>
          <div className="text-xl font-bold text-[var(--color-text-primary)] tabular-nums">{formatBytes(totalBytes)}</div>
        </div>
        <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-lg px-4 py-3">
          <div className="text-xs text-[var(--color-text-muted)]">Total Requests</div>
          <div className="text-xl font-bold text-[var(--color-text-primary)] tabular-nums">{totalRequests.toLocaleString()}</div>
        </div>
      </div>

      {/* Table */}
      {loading ? (
        <TableSkeleton rows={8} cols={6} />
      ) : stats.length === 0 ? (
        <EmptyState
          icon={Globe}
          title="Chưa có dữ liệu"
          description={selectedProxyId > 0
            ? `Chưa có traffic qua port của proxy #${selectedProxyId}`
            : "Hãy start proxy và gửi traffic để xem thống kê domain"
          }
        />
      ) : (
        <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl overflow-hidden">
          <div className="max-h-[calc(100vh-420px)] overflow-y-auto">
            <table className="w-full text-sm">
              <thead className="sticky top-0 bg-[var(--color-bg-elevated)] z-10">
                <tr className="border-b border-[var(--color-border)]">
                  <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-text-muted)] w-8">#</th>
                  <th className="text-left px-4 py-2.5 text-xs font-medium text-[var(--color-text-muted)]">Domain</th>
                  <th
                    className="text-right px-4 py-2.5 text-xs font-medium text-[var(--color-text-muted)] cursor-pointer select-none hover:text-[var(--color-text-primary)]"
                    onClick={() => setSortBy('totalBytes')}
                  >
                    <span className="inline-flex items-center gap-1">
                      Bandwidth {sortBy === 'totalBytes' && <ArrowUpDown size={10} />}
                    </span>
                  </th>
                  <th
                    className="text-right px-4 py-2.5 text-xs font-medium text-[var(--color-text-muted)] cursor-pointer select-none hover:text-[var(--color-text-primary)]"
                    onClick={() => setSortBy('requests')}
                  >
                    <span className="inline-flex items-center gap-1">
                      Requests {sortBy === 'requests' && <ArrowUpDown size={10} />}
                    </span>
                  </th>
                  <th className="text-center px-4 py-2.5 text-xs font-medium text-[var(--color-text-muted)]">Route</th>
                  {selectedProxyId === 0 && (
                    <th className="text-center px-4 py-2.5 text-xs font-medium text-[var(--color-text-muted)]">Port</th>
                  )}
                  <th className="text-right px-4 py-2.5 text-xs font-medium text-[var(--color-text-muted)]">Cache Hit</th>
                </tr>
              </thead>
              <tbody>
                {sorted.map((stat, i) => {
                  const pct = totalBytes > 0 ? (stat.totalBytes / totalBytes) * 100 : 0
                  const rc = routeColors[stat.route] || { bg: 'bg-gray-500/10', text: 'text-gray-400', label: stat.route }
                  // Find the port for this proxy ID
                  const proxyPort = outputProxies.find((p) => p.proxyId === stat.proxyId && p.protocol === 'http')
                  return (
                    <tr
                      key={`${stat.domain}-${stat.proxyId}`}
                      className="border-b border-[var(--color-border-subtle)] hover:bg-[var(--color-sidebar-hover)] transition-colors"
                    >
                      <td className="px-4 py-2 text-xs text-[var(--color-text-muted)] tabular-nums">{i + 1}</td>
                      <td className="px-4 py-2">
                        <div className="font-mono text-xs text-[var(--color-text-primary)]">{stat.domain}</div>
                        <div className="mt-1 h-1 bg-[var(--color-border)] rounded-full overflow-hidden w-48">
                          <div
                            className="h-full bg-[var(--color-primary)] rounded-full transition-all"
                            style={{ width: `${Math.max(pct, 1)}%` }}
                          />
                        </div>
                      </td>
                      <td className="px-4 py-2 text-right font-mono text-xs text-[var(--color-text-secondary)] tabular-nums">
                        {formatBytes(stat.totalBytes)}
                        <span className="text-[var(--color-text-muted)] ml-1">({pct.toFixed(1)}%)</span>
                      </td>
                      <td className="px-4 py-2 text-right font-mono text-xs text-[var(--color-text-secondary)] tabular-nums">
                        {stat.requests.toLocaleString()}
                      </td>
                      <td className="px-4 py-2 text-center">
                        <span className={`text-[11px] font-medium px-2 py-0.5 rounded ${rc.bg} ${rc.text}`}>
                          {rc.label}
                        </span>
                      </td>
                      {selectedProxyId === 0 && (
                        <td className="px-4 py-2 text-center">
                          {proxyPort ? (
                            <span className="text-[11px] font-mono text-[var(--color-text-muted)]">
                              {proxyPort.localPort}
                            </span>
                          ) : stat.proxyId > 0 ? (
                            <span className="text-[11px] font-mono text-[var(--color-text-muted)]">
                              #{stat.proxyId}
                            </span>
                          ) : (
                            <span className="text-[11px] text-[var(--color-text-muted)]">main</span>
                          )}
                        </td>
                      )}
                      <td className="px-4 py-2 text-right text-xs tabular-nums">
                        <span className={stat.cacheHitPct > 50 ? 'text-[var(--color-success)]' : 'text-[var(--color-text-muted)]'}>
                          {stat.cacheHitPct.toFixed(1)}%
                        </span>
                      </td>
                    </tr>
                  )
                })}
              </tbody>
            </table>
          </div>
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
