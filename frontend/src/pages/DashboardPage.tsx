import { useEffect, memo, useMemo } from 'react'
import { Activity, DollarSign, Zap, TrendingDown, ArrowDown, ArrowUp, Wifi } from 'lucide-react'
import { AreaChart, Area, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts'
import { useBandwidthStore } from '../stores/bandwidthStore'
import { useWailsEvent } from '../hooks/useWailsEvent'
import { formatBytes, formatBytesPerSec, formatCost, formatPercent } from '../lib/format'

const StatCard = memo(function StatCard({ icon: Icon, label, value, sub, color }: {
  icon: typeof Activity
  label: string
  value: string
  sub?: string
  color: string
}) {
  return (
    <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl p-4 transition-transform duration-200 hover:-translate-y-0.5 hover:shadow-md">
      <div className="flex items-center gap-3 mb-2">
        <div className={`p-2 rounded-lg ${color}`}>
          <Icon size={18} />
        </div>
        <span className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">{label}</span>
      </div>
      <div className="text-2xl font-bold text-[var(--color-text-primary)] tabular-nums">{value}</div>
      {sub && <div className="text-xs text-[var(--color-text-muted)] mt-1">{sub}</div>}
    </div>
  )
})

const SpeedIndicator = memo(function SpeedIndicator() {
  const bytesPerSecond = useBandwidthStore((s) => s.bytesPerSecond)
  const residentialBPS = useBandwidthStore((s) => s.residentialBPS)
  const directBPS = bytesPerSecond - residentialBPS

  return (
    <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl p-4">
      <div className="flex items-center gap-2 mb-3">
        <Wifi size={16} className="text-[var(--color-primary)]" />
        <span className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide">Realtime Speed</span>
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <div className="flex items-center gap-1 text-[var(--color-text-muted)] text-xs mb-1">
            <ArrowDown size={12} className="text-[var(--color-danger)]" />
            Residential
          </div>
          <div className="text-xl font-bold text-[var(--color-danger)] tabular-nums">
            {formatBytesPerSec(residentialBPS)}
          </div>
        </div>
        <div>
          <div className="flex items-center gap-1 text-[var(--color-text-muted)] text-xs mb-1">
            <ArrowUp size={12} className="text-[var(--color-success)]" />
            Direct/DC
          </div>
          <div className="text-xl font-bold text-[var(--color-success)] tabular-nums">
            {formatBytesPerSec(directBPS > 0 ? directBPS : 0)}
          </div>
        </div>
      </div>
      <div className="mt-3 pt-3 border-t border-[var(--color-border)]">
        <div className="flex justify-between text-xs text-[var(--color-text-muted)]">
          <span>Total throughput</span>
          <span className="text-[var(--color-text-secondary)] font-medium tabular-nums">{formatBytesPerSec(bytesPerSecond)}</span>
        </div>
      </div>
    </div>
  )
})

function formatChartBytes(bytes: number): string {
  if (bytes === 0) return '0'
  if (bytes < 1024) return `${bytes}B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)}K`
  return `${(bytes / (1024 * 1024)).toFixed(1)}M`
}

const BandwidthChart = memo(function BandwidthChart() {
  const speedHistory = useBandwidthStore((s) => s.speedHistory)

  return (
    <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl p-4">
      <div className="flex items-center justify-between mb-4">
        <h2 className="text-sm font-medium text-[var(--color-text-secondary)]">Bandwidth (Live)</h2>
        <span className="text-xs text-[var(--color-text-muted)]">Last {speedHistory.length}s</span>
      </div>
      <div className="h-56">
        {speedHistory.length > 2 ? (
          <ChartContent data={speedHistory} />
        ) : (
          <div className="h-full flex items-center justify-center text-[var(--color-text-muted)] text-sm">
            Đang chờ dữ liệu... Hãy start proxy và gửi traffic qua localhost:8888
          </div>
        )}
      </div>
      <div className="flex gap-4 mt-2 text-xs text-[var(--color-text-muted)]">
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-0.5 bg-[var(--color-primary)] rounded" />
          Total
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-0.5 bg-[var(--color-danger)] rounded" />
          Residential
        </div>
      </div>
    </div>
  )
})

const ChartContent = memo(function ChartContent({ data }: { data: { time: string; total: number; residential: number }[] }) {
  return (
    <ResponsiveContainer width="100%" height="100%">
      <AreaChart data={data} syncId="bw">
        <defs>
          <linearGradient id="colorRes" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--color-danger)" stopOpacity={0.3} />
            <stop offset="95%" stopColor="var(--color-danger)" stopOpacity={0} />
          </linearGradient>
          <linearGradient id="colorTotal" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor="var(--color-primary)" stopOpacity={0.3} />
            <stop offset="95%" stopColor="var(--color-primary)" stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid strokeDasharray="3 3" stroke="var(--color-chart-grid)" />
        <XAxis
          dataKey="time"
          tick={{ fontSize: 10, fill: 'var(--color-chart-axis)' }}
          interval="preserveStartEnd"
          minTickGap={40}
        />
        <YAxis
          tick={{ fontSize: 10, fill: 'var(--color-chart-axis)' }}
          tickFormatter={formatChartBytes}
          width={50}
        />
        <Tooltip
          contentStyle={{
            backgroundColor: 'var(--color-chart-tooltip-bg)',
            border: '1px solid var(--color-chart-tooltip-border)',
            borderRadius: '8px',
            fontSize: '12px',
          }}
          labelStyle={{ color: 'var(--color-text-muted)' }}
          formatter={(value, name) => [
            formatBytesPerSec(Number(value ?? 0)),
            name === 'total' ? 'Total' : 'Residential',
          ]}
        />
        <Area
          type="monotone"
          dataKey="total"
          stroke="var(--color-primary)"
          fill="url(#colorTotal)"
          strokeWidth={2}
          animationDuration={0}
          dot={false}
        />
        <Area
          type="monotone"
          dataKey="residential"
          stroke="var(--color-danger)"
          fill="url(#colorRes)"
          strokeWidth={2}
          animationDuration={0}
          dot={false}
        />
      </AreaChart>
    </ResponsiveContainer>
  )
})

const ConnectionsInfo = memo(function ConnectionsInfo() {
  const activeConnections = useBandwidthStore((s) => s.activeConnections)

  return (
    <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl p-4">
      <h2 className="text-xs text-[var(--color-text-muted)] uppercase tracking-wide mb-3">Proxy Info</h2>
      <div className="space-y-2 text-sm">
        <div className="flex justify-between">
          <span className="text-[var(--color-text-muted)]">Output Ports</span>
          <span className="text-[var(--color-text-secondary)] font-mono text-xs">30000+</span>
        </div>
        <div className="flex justify-between">
          <span className="text-[var(--color-text-muted)]">Protocol</span>
          <span className="text-[var(--color-text-secondary)] text-xs">HTTP + SOCKS5</span>
        </div>
        <div className="flex justify-between">
          <span className="text-[var(--color-text-muted)]">Active Conns</span>
          <span className="text-[var(--color-text-secondary)] tabular-nums">{activeConnections}</span>
        </div>
      </div>
    </div>
  )
})

export default function DashboardPage() {
  const totalToday = useBandwidthStore((s) => s.totalToday)
  const residentialToday = useBandwidthStore((s) => s.residentialToday)
  const costToday = useBandwidthStore((s) => s.costToday)
  const cacheHitRatio = useBandwidthStore((s) => s.cacheHitRatio)
  const initialize = useBandwidthStore((s) => s.initialize)
  const updateFromEvent = useBandwidthStore((s) => s.updateFromEvent)

  useEffect(() => {
    initialize()
  }, [initialize])

  useWailsEvent('bandwidth:update', updateFromEvent)

  const savedToday = totalToday - residentialToday
  const savedPct = useMemo(
    () => totalToday > 0 ? savedToday / totalToday : 0,
    [savedToday, totalToday]
  )

  return (
    <div className="p-6 overflow-y-auto h-full space-y-4">
      <div className="grid grid-cols-4 gap-4">
        <div className="animate-fade-in-up stagger-1">
          <StatCard
            icon={Activity}
            label="Residential Used"
            value={formatBytes(residentialToday)}
            sub={`of ${formatBytes(totalToday)} total`}
            color="bg-[var(--color-danger-bg)] text-[var(--color-danger)]"
          />
        </div>
        <div className="animate-fade-in-up stagger-2">
          <StatCard
            icon={TrendingDown}
            label="Saved Today"
            value={formatBytes(savedToday)}
            sub={savedPct > 0 ? `${formatPercent(savedPct)} tiết kiệm` : undefined}
            color="bg-[var(--color-success-bg)] text-[var(--color-success)]"
          />
        </div>
        <div className="animate-fade-in-up stagger-3">
          <StatCard
            icon={DollarSign}
            label="Cost Today"
            value={formatCost(costToday)}
            color="bg-[var(--color-warning-bg)] text-[var(--color-warning)]"
          />
        </div>
        <div className="animate-fade-in-up stagger-4">
          <StatCard
            icon={Zap}
            label="Cache Hit Rate"
            value={formatPercent(cacheHitRatio)}
            color="bg-[var(--color-primary-bg)] text-[var(--color-primary)]"
          />
        </div>
      </div>

      <div className="grid grid-cols-3 gap-4 animate-fade-in-up stagger-5">
        <div className="col-span-2">
          <BandwidthChart />
        </div>
        <div className="space-y-4">
          <SpeedIndicator />
          <ConnectionsInfo />
        </div>
      </div>
    </div>
  )
}
