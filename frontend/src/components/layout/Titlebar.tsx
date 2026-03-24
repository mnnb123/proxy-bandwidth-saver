import { Minus, Square, X, Sun, Moon } from 'lucide-react'
import { useThemeStore } from '../../stores/themeStore'

const isWails = typeof (window as any).__wails_invoke !== 'undefined'

export default function Titlebar() {
  const resolved = useThemeStore((s) => s.resolved)
  const toggle = useThemeStore((s) => s.toggle)

  const handleMinimize = async () => {
    if (isWails) {
      const { WindowMinimise } = await import('../../../wailsjs/runtime/runtime')
      WindowMinimise()
    }
  }
  const handleMaximize = async () => {
    if (isWails) {
      const { WindowToggleMaximise } = await import('../../../wailsjs/runtime/runtime')
      WindowToggleMaximise()
    }
  }
  const handleClose = async () => {
    if (isWails) {
      const { Quit } = await import('../../../wailsjs/runtime/runtime')
      Quit()
    }
  }

  return (
    <div
      className="h-8 bg-[var(--color-sidebar-bg)] border-b border-[var(--color-border)] flex items-center justify-between select-none"
      style={isWails ? { WebkitAppRegion: 'drag' } as React.CSSProperties : undefined}
    >
      <div className="flex items-center gap-2 px-3">
        <div className="w-3 h-3 rounded-full bg-[var(--color-primary)]" />
        <span className="text-xs font-medium text-[var(--color-text-secondary)]">
          Proxy Bandwidth Saver
        </span>
      </div>

      <div
        className="flex items-center h-full"
        style={isWails ? { WebkitAppRegion: 'no-drag' } as React.CSSProperties : undefined}
      >
        {/* Theme Toggle */}
        <button
          onClick={toggle}
          className="h-full px-2.5 hover:bg-[var(--color-sidebar-hover)] transition-colors text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]"
          aria-label={resolved === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'}
          title={resolved === 'dark' ? 'Light mode' : 'Dark mode'}
        >
          {resolved === 'dark' ? <Sun size={14} /> : <Moon size={14} />}
        </button>

        {/* Window Controls */}
        {isWails && (
          <>
            <button onClick={handleMinimize} className="h-full px-3 hover:bg-[var(--color-sidebar-hover)] transition-colors text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]" aria-label="Minimize">
              <Minus size={14} />
            </button>
            <button onClick={handleMaximize} className="h-full px-3 hover:bg-[var(--color-sidebar-hover)] transition-colors text-[var(--color-text-muted)] hover:text-[var(--color-text-primary)]" aria-label="Maximize">
              <Square size={12} />
            </button>
            <button onClick={handleClose} className="h-full px-3 hover:bg-[var(--color-danger)] transition-colors text-[var(--color-text-muted)] hover:text-white" aria-label="Close">
              <X size={14} />
            </button>
          </>
        )}
      </div>
    </div>
  )
}
