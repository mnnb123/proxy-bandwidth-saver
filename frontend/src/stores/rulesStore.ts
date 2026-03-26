import { create } from 'zustand'
import { GetRules, AddRule, AddBulkRules, UpdateRuleById, DeleteRule, ToggleRule, TestRule, ImportRules, ExportRules } from '../lib/api'
import { useToastStore } from './toastStore'

interface Rule {
  id: number; ruleType: string; pattern: string; action: string
  priority: number; enabled: boolean; hitCount: number; bytesSaved: number; createdAt: any
}

interface RulesState {
  rules: Rule[]
  loading: boolean
  testResult: string | null
  fetchRules: () => Promise<void>
  createRule: (ruleType: string, pattern: string, action: string, priority: number) => Promise<void>
  createBulkRules: (patterns: string[], action: string, priority: number) => Promise<number>
  updateRule: (id: number, ruleType: string, pattern: string, action: string, priority: number, enabled: boolean) => Promise<void>
  deleteRule: (id: number) => Promise<void>
  toggleRule: (id: number, enabled: boolean) => Promise<void>
  testRule: (domain: string, url: string, contentType: string) => Promise<void>
  importRules: (json: string) => Promise<number>
  exportRules: () => Promise<string>
}

const toast = () => useToastStore.getState()

export const useRulesStore = create<RulesState>((set) => ({
  rules: [],
  loading: false,
  testResult: null,

  fetchRules: async () => {
    set({ loading: true })
    try {
      const rules = await GetRules()
      set({ rules: rules || [] })
    } catch (e) {
      toast().addToast('error', `Failed to load rules: ${e}`)
    } finally {
      set({ loading: false })
    }
  },

  createRule: async (ruleType, pattern, action, priority) => {
    try {
      await AddRule(ruleType, pattern, action, priority)
      toast().addToast('success', 'Rule created')
      const rules = await GetRules()
      set({ rules: rules || [] })
    } catch (e) {
      toast().addToast('error', `Failed to create rule: ${e}`)
    }
  },

  createBulkRules: async (patterns, action, priority) => {
    try {
      const count = await AddBulkRules(patterns, action, priority)
      toast().addToast('success', `Added ${count} rules`)
      const rules = await GetRules()
      set({ rules: rules || [] })
      return count
    } catch (e) {
      toast().addToast('error', `Failed to create rules: ${e}`)
      return 0
    }
  },

  updateRule: async (id, ruleType, pattern, action, priority, enabled) => {
    try {
      await UpdateRuleById(id, ruleType, pattern, action, priority, enabled)
      toast().addToast('success', 'Rule updated')
      const rules = await GetRules()
      set({ rules: rules || [] })
    } catch (e) {
      toast().addToast('error', `Failed to update rule: ${e}`)
    }
  },

  deleteRule: async (id) => {
    try {
      await DeleteRule(id)
      toast().addToast('success', 'Rule deleted')
      set((s) => ({ rules: s.rules.filter((r) => r.id !== id) }))
    } catch (e) {
      toast().addToast('error', `Failed to delete rule: ${e}`)
    }
  },

  toggleRule: async (id, enabled) => {
    try {
      await ToggleRule(id, enabled)
      set((s) => ({
        rules: s.rules.map((r) => r.id === id ? { ...r, enabled } : r),
      }))
    } catch (e) {
      toast().addToast('error', `Failed to toggle rule: ${e}`)
    }
  },

  testRule: async (domain, url, contentType) => {
    try {
      const result = await TestRule(domain, url, contentType)
      set({ testResult: result })
    } catch (e) {
      toast().addToast('error', `Test failed: ${e}`)
    }
  },

  importRules: async (json) => {
    try {
      const count = await ImportRules(json)
      toast().addToast('success', `Imported ${count} rules`)
      const rules = await GetRules()
      set({ rules: rules || [] })
      return count
    } catch (e) {
      toast().addToast('error', `Import failed: ${e}`)
      return 0
    }
  },

  exportRules: async () => {
    try {
      return await ExportRules()
    } catch (e) {
      toast().addToast('error', `Export failed: ${e}`)
      return ''
    }
  },
}))
