import { useState, useEffect, lazy, Suspense } from 'react'
import Titlebar from './components/layout/Titlebar'
import Sidebar, { type Page } from './components/layout/Sidebar'
import { ToastContainer } from './components/ui/Toast'
import { useProxyStore } from './stores/proxyStore'
import { DashboardSkeleton, TableSkeleton, SettingsSkeleton } from './components/ui/Skeleton'

const DashboardPage = lazy(() => import('./pages/DashboardPage'))
const RulesPage = lazy(() => import('./pages/RulesPage'))
const ProxiesPage = lazy(() => import('./pages/ProxiesPage'))
const SettingsPage = lazy(() => import('./pages/SettingsPage'))

const pages: Record<Page, React.ComponentType> = {
  dashboard: DashboardPage,
  rules: RulesPage,
  proxies: ProxiesPage,
  settings: SettingsPage,
}

const fallbacks: Record<Page, React.ReactNode> = {
  dashboard: <DashboardSkeleton />,
  rules: <div className="p-6 space-y-4"><TableSkeleton rows={6} cols={6} /></div>,
  proxies: <div className="p-6 space-y-4"><TableSkeleton rows={4} cols={4} /></div>,
  settings: <div className="p-6 space-y-4">{[...Array(3)].map((_, i) => <SettingsSkeleton key={i} />)}</div>,
}

function App() {
  const [activePage, setActivePage] = useState<Page>('dashboard')
  const initialize = useProxyStore((s) => s.initialize)

  useEffect(() => {
    initialize()
  }, [initialize])

  const ActivePage = pages[activePage]

  return (
    <div className="h-screen flex flex-col bg-[var(--color-bg-base)] text-[var(--color-text-primary)]">
      <Titlebar />
      <div className="flex flex-1 overflow-hidden">
        <Sidebar activePage={activePage} onNavigate={setActivePage} />
        <main className="flex-1 overflow-hidden">
          <Suspense fallback={fallbacks[activePage]}>
            <ActivePage />
          </Suspense>
        </main>
      </div>
      <ToastContainer />
    </div>
  )
}

export default App
