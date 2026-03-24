import { create } from 'zustand'
import { GetProxies, AddProxyAPI, DeleteProxy, ImportProxies, GetOutputProxies } from '../lib/api'
import { useToastStore } from './toastStore'

interface Proxy {
  id: number; address: string; username: string; password: string; type: string
  category: string; enabled: boolean; weight: number; totalBytesUp: number
  totalBytesDown: number; totalRequests: number; failCount: number; avgLatencyMs: number
  lastCheckAt: any; lastError: string; createdAt: any
}

interface OutputProxy {
  proxyId: number; localAddr: string; localPort: number; upstream: string; type: string
}

interface ProxiesState {
  proxies: Proxy[]
  outputProxies: OutputProxy[]
  loading: boolean
  fetchProxies: () => Promise<void>
  fetchOutputProxies: () => Promise<void>
  addProxy: (address: string, username: string, password: string, proxyType: string, category: string) => Promise<void>
  deleteProxy: (id: number) => Promise<void>
  importProxies: (text: string) => Promise<number>
}

const toast = () => useToastStore.getState()

export const useProxiesStore = create<ProxiesState>((set, get) => ({
  proxies: [],
  outputProxies: [],
  loading: false,

  fetchProxies: async () => {
    set({ loading: true })
    try {
      const proxies = await GetProxies()
      set({ proxies: proxies || [] })
    } catch (e) {
      toast().addToast('error', `Failed to load proxies: ${e}`)
    } finally {
      set({ loading: false })
    }
  },

  fetchOutputProxies: async () => {
    try {
      const out = await GetOutputProxies()
      set({ outputProxies: out || [] })
    } catch (e) {
      console.error('Failed to load output proxies:', e)
    }
  },

  addProxy: async (address, username, password, proxyType, category) => {
    try {
      await AddProxyAPI(address, username, password, proxyType, category)
      toast().addToast('success', 'Proxy added')
      const proxies = await GetProxies()
      set({ proxies: proxies || [] })
      get().fetchOutputProxies()
    } catch (e) {
      toast().addToast('error', `Failed to add proxy: ${e}`)
    }
  },

  deleteProxy: async (id) => {
    try {
      await DeleteProxy(id)
      toast().addToast('success', 'Proxy removed')
      set((s) => ({ proxies: s.proxies.filter((p) => p.id !== id) }))
      get().fetchOutputProxies()
    } catch (e) {
      toast().addToast('error', `Failed to remove proxy: ${e}`)
    }
  },

  importProxies: async (text) => {
    try {
      const count = await ImportProxies(text)
      toast().addToast('success', `Imported ${count} proxies`)
      const proxies = await GetProxies()
      set({ proxies: proxies || [] })
      get().fetchOutputProxies()
      return count
    } catch (e) {
      toast().addToast('error', `Import failed: ${e}`)
      return 0
    }
  },
}))
