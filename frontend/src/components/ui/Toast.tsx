import { Toaster } from 'sonner'
import { useThemeStore } from '../../stores/themeStore'

export function ToastContainer() {
  const resolved = useThemeStore((s) => s.resolved)

  return (
    <Toaster
      theme={resolved}
      position="bottom-right"
      toastOptions={{
        style: {
          background: 'var(--color-bg-elevated)',
          border: '1px solid var(--color-border)',
          color: 'var(--color-text-primary)',
          fontSize: '13px',
        },
      }}
      gap={8}
      closeButton
    />
  )
}
