import { useState, useEffect } from 'react'
import { LayoutDashboard, Shield, Globe, Settings, Play, Square, Menu, X } from 'lucide-react'
import { useProxyStore } from '../../stores/proxyStore'

type Page = 'dashboard' | 'rules' | 'proxies' | 'settings'

interface SidebarProps {
  activePage: Page
  onNavigate: (page: Page) => void
}

const navItems: { id: Page; label: string; icon: typeof LayoutDashboard }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'rules', label: 'Rules', icon: Shield },
  { id: 'proxies', label: 'Proxies', icon: Globe },
  { id: 'settings', label: 'Settings', icon: Settings },
]

const COLLAPSE_WIDTH = 720

export default function Sidebar({ activePage, onNavigate }: SidebarProps) {
  const running = useProxyStore((s) => s.running)
  const startProxy = useProxyStore((s) => s.startProxy)
  const stopProxy = useProxyStore((s) => s.stopProxy)

  const [collapsed, setCollapsed] = useState(() => window.innerWidth < COLLAPSE_WIDTH)
  const [mobileOpen, setMobileOpen] = useState(false)

  useEffect(() => {
    const onResize = () => {
      const narrow = window.innerWidth < COLLAPSE_WIDTH
      setCollapsed(narrow)
      if (!narrow) setMobileOpen(false)
    }
    window.addEventListener('resize', onResize)
    return () => window.removeEventListener('resize', onResize)
  }, [])

  const handleNav = (page: Page) => {
    onNavigate(page)
    if (collapsed) setMobileOpen(false)
  }

  // Mobile overlay
  if (collapsed && mobileOpen) {
    return (
      <>
        {/* Backdrop */}
        <div
          className="fixed inset-0 z-40 bg-[var(--color-bg-overlay)]"
          onClick={() => setMobileOpen(false)}
        />
        {/* Slide-out sidebar */}
        <div className="fixed inset-y-0 left-0 z-50 w-52 bg-[var(--color-sidebar-bg)] border-r border-[var(--color-border)] flex flex-col animate-slide-in-left">
          <div className="flex items-center justify-between px-4 py-3 border-b border-[var(--color-border)]">
            <span className="text-xs font-medium text-[var(--color-text-secondary)]">Menu</span>
            <button
              onClick={() => setMobileOpen(false)}
              className="p-1 rounded hover:bg-[var(--color-sidebar-hover)] text-[var(--color-text-muted)]"
              aria-label="Close menu"
            >
              <X size={16} />
            </button>
          </div>
          <SidebarContent activePage={activePage} onNavigate={handleNav} running={running} startProxy={startProxy} stopProxy={stopProxy} />
        </div>
      </>
    )
  }

  // Collapsed: show hamburger button only
  if (collapsed) {
    return (
      <div className="w-12 bg-[var(--color-sidebar-bg)] border-r border-[var(--color-border)] flex flex-col items-center py-3 h-full">
        <button
          onClick={() => setMobileOpen(true)}
          className="p-2 rounded-[var(--radius-lg)] hover:bg-[var(--color-sidebar-hover)] text-[var(--color-text-secondary)] transition-colors"
          aria-label="Open menu"
        >
          <Menu size={18} />
        </button>
        <nav className="flex-1 flex flex-col items-center gap-1 mt-3">
          {navItems.map((item) => {
            const Icon = item.icon
            const isActive = activePage === item.id
            return (
              <button
                key={item.id}
                onClick={() => handleNav(item.id)}
                title={item.label}
                className={`p-2 rounded-[var(--radius-lg)] transition-all duration-150 active:scale-[0.97]
                  ${isActive
                    ? 'bg-[var(--color-sidebar-active)] text-[var(--color-primary)]'
                    : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-sidebar-hover)]'
                  }`}
              >
                <Icon size={18} />
              </button>
            )
          })}
        </nav>
        <div className="pb-3">
          <div className={`w-2 h-2 rounded-full mx-auto mb-2 ${running ? 'bg-[var(--color-success)]' : 'bg-[var(--color-danger)]'}`} />
          <button
            onClick={running ? stopProxy : startProxy}
            title={running ? 'Stop Proxy' : 'Start Proxy'}
            className={`p-2 rounded-[var(--radius-lg)] transition-all duration-150 active:scale-[0.97]
              ${running
                ? 'text-[var(--color-danger)] hover:bg-[var(--color-danger-bg)]'
                : 'text-[var(--color-success)] hover:bg-[var(--color-success-bg)]'
              }`}
          >
            {running ? <Square size={16} /> : <Play size={16} />}
          </button>
        </div>
      </div>
    )
  }

  // Full sidebar
  return (
    <div className="w-52 bg-[var(--color-sidebar-bg)] border-r border-[var(--color-border)] flex flex-col h-full">
      <SidebarContent activePage={activePage} onNavigate={handleNav} running={running} startProxy={startProxy} stopProxy={stopProxy} />
    </div>
  )
}

function SidebarContent({ activePage, onNavigate, running, startProxy, stopProxy }: {
  activePage: Page; onNavigate: (page: Page) => void; running: boolean; startProxy: () => void; stopProxy: () => void
}) {
  return (
    <>
      <nav className="flex-1 py-3">
        {navItems.map((item) => {
          const Icon = item.icon
          const isActive = activePage === item.id
          return (
            <button
              key={item.id}
              onClick={() => onNavigate(item.id)}
              className={`w-full flex items-center gap-3 px-4 py-2.5 text-sm transition-all duration-150 active:scale-[0.97]
                ${isActive
                  ? 'bg-[var(--color-sidebar-active)] text-[var(--color-primary)] border-l-2 border-[var(--color-primary)]'
                  : 'text-[var(--color-text-secondary)] hover:text-[var(--color-text-primary)] hover:bg-[var(--color-sidebar-hover)] border-l-2 border-transparent'
                }`}
            >
              <Icon size={18} />
              {item.label}
            </button>
          )
        })}
      </nav>

      <div className="p-4 border-t border-[var(--color-border)]">
        <div className="flex items-center gap-2 mb-3">
          <div className={`w-2 h-2 rounded-full ${running ? 'bg-[var(--color-success)]' : 'bg-[var(--color-danger)]'}`} />
          <span className="text-xs text-[var(--color-text-muted)]">
            {running ? 'Running' : 'Stopped'}
          </span>
        </div>
        <button
          onClick={running ? stopProxy : startProxy}
          className={`w-full flex items-center justify-center gap-2 py-2 rounded-[var(--radius-lg)] text-sm font-medium transition-all duration-150 cursor-pointer active:scale-[0.97]
            ${running
              ? 'bg-[var(--color-danger-bg)] text-[var(--color-danger-text)] hover:bg-[var(--color-danger)] hover:text-white'
              : 'bg-[var(--color-success-bg)] text-[var(--color-success-text)] hover:bg-[var(--color-success)] hover:text-white'
            }`}
        >
          {running ? <Square size={14} /> : <Play size={14} />}
          {running ? 'Stop Proxy' : 'Start Proxy'}
        </button>
      </div>
    </>
  )
}

export type { Page }
