import { create } from 'zustand'

type Theme = 'light' | 'dark' | 'system'

interface ThemeState {
  theme: Theme
  resolved: 'light' | 'dark'
  setTheme: (theme: Theme) => void
  toggle: () => void
}

function getSystemTheme(): 'light' | 'dark' {
  if (typeof window === 'undefined') return 'dark'
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
}

function resolveTheme(theme: Theme): 'light' | 'dark' {
  return theme === 'system' ? getSystemTheme() : theme
}

function applyTheme(resolved: 'light' | 'dark') {
  const root = document.documentElement
  // Add transitioning class for smooth switch
  root.classList.add('transitioning')
  if (resolved === 'dark') {
    root.classList.add('dark')
  } else {
    root.classList.remove('dark')
  }
  // Remove transitioning class after animation
  setTimeout(() => root.classList.remove('transitioning'), 350)
}

function loadSavedTheme(): Theme {
  try {
    const saved = localStorage.getItem('pbs-theme')
    if (saved === 'light' || saved === 'dark' || saved === 'system') return saved
  } catch {}
  return 'system'
}

const initialTheme = loadSavedTheme()
const initialResolved = resolveTheme(initialTheme)

// Apply on module load (before React renders)
if (typeof document !== 'undefined') {
  if (initialResolved === 'dark') {
    document.documentElement.classList.add('dark')
  }
}

export const useThemeStore = create<ThemeState>((set, get) => ({
  theme: initialTheme,
  resolved: initialResolved,

  setTheme: (theme: Theme) => {
    const resolved = resolveTheme(theme)
    localStorage.setItem('pbs-theme', theme)
    applyTheme(resolved)
    set({ theme, resolved })
  },

  toggle: () => {
    const { resolved } = get()
    const newTheme: Theme = resolved === 'dark' ? 'light' : 'dark'
    const newResolved = resolveTheme(newTheme)
    localStorage.setItem('pbs-theme', newTheme)
    applyTheme(newResolved)
    set({ theme: newTheme, resolved: newResolved })
  },
}))

// Listen for OS theme changes when using 'system'
if (typeof window !== 'undefined') {
  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    const state = useThemeStore.getState()
    if (state.theme === 'system') {
      const resolved = getSystemTheme()
      applyTheme(resolved)
      useThemeStore.setState({ resolved })
    }
  })
}
