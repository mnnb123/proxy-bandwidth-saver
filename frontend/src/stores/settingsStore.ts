import { create } from 'zustand'
import { GetSettings, UpdateSetting } from '../lib/api'
import { useToastStore } from './toastStore'

interface SettingsState {
  settings: Record<string, string>
  loading: boolean
  initialize: () => Promise<void>
  get: (key: string, fallback?: string) => string
  update: (key: string, value: string) => Promise<void>
  updateBatch: (entries: Record<string, string>) => Promise<void>
}

const toast = () => useToastStore.getState()

export const useSettingsStore = create<SettingsState>((set, getState) => ({
  settings: {},
  loading: false,

  initialize: async () => {
    set({ loading: true })
    try {
      const settings = await GetSettings()
      set({ settings: settings || {} })
    } catch (e) {
      toast().addToast('error', `Failed to load settings: ${e}`)
    } finally {
      set({ loading: false })
    }
  },

  get: (key, fallback = '') => {
    return getState().settings[key] ?? fallback
  },

  update: async (key, value) => {
    try {
      await UpdateSetting(key, value)
      set((s) => ({ settings: { ...s.settings, [key]: value } }))
    } catch (e) {
      toast().addToast('error', `Failed to save setting: ${e}`)
    }
  },

  updateBatch: async (entries) => {
    try {
      for (const [key, value] of Object.entries(entries)) {
        await UpdateSetting(key, value)
      }
      set((s) => ({ settings: { ...s.settings, ...entries } }))
      toast().addToast('success', 'Settings saved')
    } catch (e) {
      toast().addToast('error', `Failed to save settings: ${e}`)
    }
  },
}))
