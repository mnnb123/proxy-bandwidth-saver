import { create } from 'zustand'
import { GetProxyStatus, StartProxy, StopProxy } from '../lib/api'
import { useToastStore } from './toastStore'

const toast = () => useToastStore.getState()

interface ProxyState {
  running: boolean
  httpPort: number
  socks5Port: number
  uptime: number
  connections: number
  initialize: () => Promise<void>
  startProxy: () => Promise<void>
  stopProxy: () => Promise<void>
  refresh: () => Promise<void>
}

export const useProxyStore = create<ProxyState>((set) => ({
  running: false,
  httpPort: 8888,
  socks5Port: 8889,
  uptime: 0,
  connections: 0,

  initialize: async () => {
    try {
      const status = await GetProxyStatus()
      set({
        running: status.running,
        httpPort: status.httpPort,
        socks5Port: status.socks5Port,
        uptime: status.uptime,
        connections: status.connections,
      })
    } catch (e) {
      console.error('Failed to get proxy status:', e)
    }
  },

  startProxy: async () => {
    try {
      await StartProxy()
      const status = await GetProxyStatus()
      set({ running: status.running })
      toast().addToast('success', 'Proxy started')
    } catch (e) {
      toast().addToast('error', `Failed to start proxy: ${e}`)
    }
  },

  stopProxy: async () => {
    try {
      await StopProxy()
      set({ running: false })
      toast().addToast('success', 'Proxy stopped')
    } catch (e) {
      toast().addToast('error', `Failed to stop proxy: ${e}`)
    }
  },

  refresh: async () => {
    try {
      const status = await GetProxyStatus()
      set({
        running: status.running,
        uptime: status.uptime,
        connections: status.connections,
      })
    } catch (e) {
      console.error('Failed to refresh proxy status:', e)
    }
  },
}))
