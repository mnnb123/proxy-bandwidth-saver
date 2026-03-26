import { useEffect, useState, useCallback, useRef } from 'react'
import { Globe, RefreshCw, ArrowUpDown } from 'lucide-react'
import { GetDomainStats } from '../lib/api'
import { formatBytes } from '../lib/format'
import { TableSkeleton } from '../components/ui/Skeleton'
import { EmptyState } from '../components/ui/EmptyState'

interface DomainStat {
  domain: string
  totalBytes: number
  requests: number
  route: string
  cacheHitPct: number
}

const PERIODS = [
  { value: '1h', label: '1 giờ' },
  { value: '24h', label: '24 giờ' },
  { value: '7d', label: '7 ngày' },
  { value: '30d', label: '30 ngày' },
]

const routeColors: Record<string, string> = {
  residential: 'text-[var(--color-danger)]',
  datacenter: 'text-[var(--color-primary)]',
  direct: 'text-[var(--color-success)]',
}

export default function DomainsPage() {
  const [stats, setStats] = useState<DomainStat[]>([])
  const [loading, setLoading] = useState(true)
  const [period, setPeriod] = useState('24h')
  const [sortBy, setSortBy] = useState<'totalBytes' | 'requests'>('totalBytes')

  const fetchStats = useCallback(async (p: string, silent = false) => {
    if (!silent) setLoading(true)
    try {
      const data = await GetDomainStats(p)
      setStats(data || [])
    } catch {
      if (!silent) setStats([])
    }
    if (!silent) setLoading(false)
  }, [])

  // Initial load
  useEffect(() => {
    fetchStats(period)
  }, [period, fetchStats])

  // Auto-refresh every 3 seconds (silent, no loading spinner)
  const periodRef = useRef(period)
  periodRef.current = period
  useEffect(() => {
    const interval = setInterval(() => fetchStats(periodRef.current, true), 3000)
    return () => clearInterval(interval)
  }, [fetchStats])

  const sorted = [...stats].sort((a, b) => b[sortBy] - a[sortBy])

  const totalBytes = stats.reduce((sum, s) => sum + s.totalBytes, 0)
  const totalRequests = stats.reduce((sum, s) => sum + s.requests, 0)

  return (
    <div className="p-6 overflow-y-auto h-full space-y-4">
      {/* Header */}
      <div className="flex items-center justify-between">
        <h1 className="text-lg font-semibold text-[var(--color-text-primary)]">Domain Report</h1>
        <div className="flex items-center gap-2">
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
            onClick={() => fetchStats(period)}
            className="p-1.5 rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] border border-[var(--color-border)] hover:bg-[var(--color-sidebar-hover)] transition-colors"
            aria-label="Refresh"
          >
            <RefreshCw size={14} />
          </button>
        </div>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-3 gap-3">
        <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-lg px-4 py-3">
          <div className="text-xs text-[var(--color-text-muted)]">Domains</div>
          <div className="text-xl font-bold text-[var(--color-text-primary)] tabular-nums">{stats.length}</div>
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
        <TableSkeleton rows={8} cols={5} />
      ) : stats.length === 0 ? (
        <EmptyState
          icon={Globe}
          title="Chưa có dữ liệu"
          description="Hãy start proxy và gửi traffic để xem thống kê domain"
        />
      ) : (
        <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl overflow-hidden">
          <div className="max-h-[calc(100vh-320px)] overflow-y-auto">
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
                  <th className="text-right px-4 py-2.5 text-xs font-medium text-[var(--color-text-muted)]">Cache Hit</th>
                </tr>
              </thead>
              <tbody>
                {sorted.map((stat, i) => {
                  const pct = totalBytes > 0 ? (stat.totalBytes / totalBytes) * 100 : 0
                  return (
                    <tr
                      key={stat.domain}
                      className="border-b border-[var(--color-border-subtle)] hover:bg-[var(--color-sidebar-hover)] transition-colors"
                    >
                      <td className="px-4 py-2 text-xs text-[var(--color-text-muted)] tabular-nums">{i + 1}</td>
                      <td className="px-4 py-2">
                        <div className="font-mono text-xs text-[var(--color-text-primary)]">{stat.domain}</div>
                        {/* Bar */}
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
                        <span className={`text-xs font-medium ${routeColors[stat.route] || 'text-[var(--color-text-muted)]'}`}>
                          {stat.route}
                        </span>
                      </td>
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
    </div>
  )
}
