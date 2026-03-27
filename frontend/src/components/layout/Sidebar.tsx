import { useState, useEffect } from 'react'
import { LayoutDashboard, Shield, Globe, Settings, Menu, X, BarChart3 } from 'lucide-react'
import { GetVersion } from '../../lib/api'

type Page = 'dashboard' | 'rules' | 'proxies' | 'domains' | 'settings'

interface SidebarProps {
  activePage: Page
  onNavigate: (page: Page) => void
}

const navItems: { id: Page; label: string; icon: typeof LayoutDashboard }[] = [
  { id: 'dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { id: 'rules', label: 'Rules', icon: Shield },
  { id: 'proxies', label: 'Proxies', icon: Globe },
  { id: 'domains', label: 'Domains', icon: BarChart3 },
  { id: 'settings', label: 'Settings', icon: Settings },
]

const COLLAPSE_WIDTH = 720

export default function Sidebar({ activePage, onNavigate }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(() => window.innerWidth < COLLAPSE_WIDTH)
  const [mobileOpen, setMobileOpen] = useState(false)
  const [version, setVersion] = useState('')

  useEffect(() => {
    GetVersion().then(setVersion)
  }, [])

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
        <div
          className="fixed inset-0 z-40 bg-[var(--color-bg-overlay)]"
          onClick={() => setMobileOpen(false)}
        />
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
          <SidebarContent activePage={activePage} onNavigate={handleNav} version={version} />
        </div>
      </>
    )
  }

  // Collapsed icon bar
  if (collapsed) {
    return (
      <div className="flex flex-col items-center">
        <button
          onClick={() => setMobileOpen(true)}
          className="p-2 m-2 rounded-[var(--radius-lg)] hover:bg-[var(--color-sidebar-hover)] text-[var(--color-text-secondary)]"
          aria-label="Open menu"
        >
          <Menu size={20} />
        </button>
        <nav className="flex flex-col items-center gap-1 mt-2">
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
      </div>
    )
  }

  // Full sidebar
  return (
    <div className="w-52 bg-[var(--color-sidebar-bg)] border-r border-[var(--color-border)] flex flex-col h-full">
      <SidebarContent activePage={activePage} onNavigate={handleNav} version={version} />
    </div>
  )
}

function SidebarContent({ activePage, onNavigate, version }: {
  activePage: Page; onNavigate: (page: Page) => void; version: string
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

      {version && (
        <div className="p-4 border-t border-[var(--color-border)]">
          <p className="text-[10px] text-[var(--color-text-muted)] text-center">v{version}</p>
        </div>
      )}
    </>
  )
}

export type { Page }
