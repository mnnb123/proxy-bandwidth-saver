/** Base skeleton pulse block */
function Bone({ className = '', style }: { className?: string; style?: React.CSSProperties }) {
  return <div className={`bg-[var(--color-bg-elevated)] rounded animate-pulse ${className}`} style={style} />
}

/** Card wrapper matching real card layout */
function SkeletonCard({ children, className = '' }: { children: React.ReactNode; className?: string }) {
  return (
    <div className={`bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl p-4 ${className}`}>
      {children}
    </div>
  )
}

/** Dashboard stat card skeleton */
export function StatCardSkeleton() {
  return (
    <SkeletonCard>
      <div className="flex items-center gap-3 mb-2">
        <Bone className="w-9 h-9 rounded-lg" />
        <Bone className="h-3 w-24" />
      </div>
      <Bone className="h-7 w-20 mb-1" />
      <Bone className="h-3 w-16" />
    </SkeletonCard>
  )
}

/** Dashboard chart skeleton */
export function ChartSkeleton() {
  return (
    <SkeletonCard>
      <div className="flex items-center justify-between mb-4">
        <Bone className="h-4 w-28" />
        <Bone className="h-3 w-16" />
      </div>
      <div className="h-56 flex items-end gap-1 px-2">
        {[40, 65, 45, 80, 55, 70, 90, 60, 75, 50, 85, 65].map((h, i) => (
          <Bone key={i} className="flex-1 rounded-t" style={{ height: `${h}%` }} />
        ))}
      </div>
      <div className="flex gap-4 mt-2">
        <Bone className="h-3 w-12" />
        <Bone className="h-3 w-16" />
      </div>
    </SkeletonCard>
  )
}

/** Speed indicator skeleton */
export function SpeedSkeleton() {
  return (
    <SkeletonCard>
      <div className="flex items-center gap-2 mb-3">
        <Bone className="w-4 h-4 rounded" />
        <Bone className="h-3 w-24" />
      </div>
      <div className="grid grid-cols-2 gap-4">
        <div>
          <Bone className="h-3 w-16 mb-1" />
          <Bone className="h-6 w-20" />
        </div>
        <div>
          <Bone className="h-3 w-16 mb-1" />
          <Bone className="h-6 w-20" />
        </div>
      </div>
    </SkeletonCard>
  )
}

/** Connections info skeleton */
export function ConnectionsSkeleton() {
  return (
    <SkeletonCard>
      <Bone className="h-3 w-16 mb-3" />
      <div className="space-y-2">
        {[...Array(4)].map((_, i) => (
          <div key={i} className="flex justify-between">
            <Bone className="h-3 w-20" />
            <Bone className="h-3 w-24" />
          </div>
        ))}
      </div>
    </SkeletonCard>
  )
}

/** Table row skeleton for rules/proxies */
export function TableRowSkeleton({ cols = 5 }: { cols?: number }) {
  return (
    <div className="flex items-center gap-3 px-3 py-2.5">
      {[...Array(cols)].map((_, i) => (
        <Bone key={i} className="h-3 flex-1" style={{ maxWidth: i === 0 ? '40%' : '20%' }} />
      ))}
    </div>
  )
}

/** Full table skeleton */
export function TableSkeleton({ rows = 5, cols = 5 }: { rows?: number; cols?: number }) {
  return (
    <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl overflow-hidden">
      <div className="border-b border-[var(--color-border)] px-3 py-2.5 flex gap-3">
        {[...Array(cols)].map((_, i) => (
          <Bone key={i} className="h-3 flex-1" style={{ maxWidth: i === 0 ? '40%' : '20%' }} />
        ))}
      </div>
      {[...Array(rows)].map((_, i) => (
        <div key={i} className="border-b border-[var(--color-border-subtle)]">
          <TableRowSkeleton cols={cols} />
        </div>
      ))}
    </div>
  )
}

/** Settings section skeleton */
export function SettingsSkeleton() {
  return (
    <SkeletonCard className="p-5">
      <div className="flex items-center gap-2 mb-1">
        <Bone className="w-4 h-4 rounded" />
        <Bone className="h-4 w-28" />
      </div>
      <Bone className="h-3 w-48 mb-4" />
      <div className="space-y-4">
        {[...Array(3)].map((_, i) => (
          <div key={i} className="flex items-center justify-between">
            <Bone className="h-3 w-24" />
            <Bone className="h-7 w-48 rounded-[var(--radius-lg)]" />
          </div>
        ))}
      </div>
    </SkeletonCard>
  )
}

/** Full dashboard skeleton */
export function DashboardSkeleton() {
  return (
    <div className="p-6 space-y-4">
      <div className="grid grid-cols-4 gap-4">
        {[...Array(4)].map((_, i) => <StatCardSkeleton key={i} />)}
      </div>
      <div className="grid grid-cols-3 gap-4">
        <div className="col-span-2"><ChartSkeleton /></div>
        <div className="space-y-4">
          <SpeedSkeleton />
          <ConnectionsSkeleton />
        </div>
      </div>
    </div>
  )
}
