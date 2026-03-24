import { create } from 'zustand'
import { GetRealtimeStats, GetCostSummary, GetBudgetStatus } from '../lib/api'

interface BandwidthUpdate {
  bytesPerSecond: number
  residentialBps: number
  totalToday: number
  residentialToday: number
  cacheHitRatio: number
  activeConnections: number
}

interface SpeedPoint {
  time: string
  total: number
  residential: number
}

interface BandwidthState {
  bytesPerSecond: number
  residentialBPS: number
  totalToday: number
  residentialToday: number
  costToday: number
  cacheHitRatio: number
  activeConnections: number
  speedHistory: SpeedPoint[]
  costSummary: {
    costToday: number
    costWeek: number
    costMonth: number
    costTotal: number
    savedBytes: number
    savedCost: number
  }
  budget: {
    monthlyBudgetGb: number
    usedGb: number
    usedPercent: number
    remainingGb: number
    costPerGb: number
  }
  initialize: () => Promise<void>
  updateFromEvent: (data: BandwidthUpdate) => void
}

const MAX_SPEED_POINTS = 120 // 2 minutes of 1s data

export const useBandwidthStore = create<BandwidthState>((set, get) => ({
  bytesPerSecond: 0,
  residentialBPS: 0,
  totalToday: 0,
  residentialToday: 0,
  costToday: 0,
  cacheHitRatio: 0,
  activeConnections: 0,
  speedHistory: [],
  costSummary: {
    costToday: 0,
    costWeek: 0,
    costMonth: 0,
    costTotal: 0,
    savedBytes: 0,
    savedCost: 0,
  },
  budget: {
    monthlyBudgetGb: 0,
    usedGb: 0,
    usedPercent: 0,
    remainingGb: 0,
    costPerGb: 0,
  },

  initialize: async () => {
    try {
      const [stats, cost, budget] = await Promise.all([
        GetRealtimeStats(),
        GetCostSummary(),
        GetBudgetStatus(),
      ])
      set({
        bytesPerSecond: stats.bytesPerSecond,
        residentialBPS: stats.residentialBps,
        totalToday: stats.totalToday,
        residentialToday: stats.residentialToday,
        costToday: stats.costToday,
        cacheHitRatio: stats.cacheHitRatio,
        activeConnections: stats.activeConnections,
        costSummary: cost,
        budget: {
          monthlyBudgetGb: budget.monthlyBudgetGb,
          usedGb: budget.usedGb,
          usedPercent: budget.usedPercent,
          remainingGb: budget.remainingGb,
          costPerGb: budget.costPerGb,
        },
      })
    } catch (e) {
      console.error('Failed to load bandwidth stats:', e)
    }
  },

  updateFromEvent: (data: BandwidthUpdate) => {
    const now = new Date()
    const timeStr = now.toLocaleTimeString('vi-VN', { hour: '2-digit', minute: '2-digit', second: '2-digit' })
    const prev = get().speedHistory
    const newHistory = [
      ...prev.slice(-(MAX_SPEED_POINTS - 1)),
      { time: timeStr, total: data.bytesPerSecond, residential: data.residentialBps },
    ]

    set({
      bytesPerSecond: data.bytesPerSecond,
      residentialBPS: data.residentialBps,
      totalToday: data.totalToday,
      residentialToday: data.residentialToday,
      cacheHitRatio: data.cacheHitRatio,
      activeConnections: data.activeConnections,
      speedHistory: newHistory,
    })
  },
}))
