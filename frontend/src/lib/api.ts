// API adapter: works with both Wails (desktop) and REST API (web/VPS).
// Auto-detects mode based on whether Wails runtime is available.

const isWails = typeof (window as any).__wails_invoke !== 'undefined'
const API_BASE = ''

async function fetchAPI<T>(path: string, options?: RequestInit): Promise<T> {
  const res = await fetch(`${API_BASE}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    ...options,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || res.statusText)
  }
  return res.json()
}

// Helper: call Wails binding by name, or fall back to REST API.
async function wailsOrAPI<T>(bindingName: string, apiPath: string, apiOptions?: RequestInit, ...wailsArgs: any[]): Promise<T> {
  if (isWails) {
    const mod = await import('../../wailsjs/go/main/App')
    return (mod as any)[bindingName](...wailsArgs)
  }
  return fetchAPI<T>(apiPath, apiOptions)
}

function postJSON(_path: string, body: object): RequestInit {
  return { method: 'POST', body: JSON.stringify(body) }
}

function putJSON(_path: string, body: object): RequestInit {
  return { method: 'PUT', body: JSON.stringify(body) }
}

// === Proxy control ===

export const StartProxy = () =>
  wailsOrAPI<void>('StartProxy', '/api/proxy/start', { method: 'POST' })

export const StopProxy = () =>
  wailsOrAPI<void>('StopProxy', '/api/proxy/stop', { method: 'POST' })

export const GetProxyStatus = () =>
  wailsOrAPI<any>('GetProxyStatus', '/api/proxy/status')

// === Rules ===

export const GetRules = () =>
  wailsOrAPI<any[]>('GetRules', '/api/rules')

export const AddRule = (ruleType: string, pattern: string, action: string, priority: number) =>
  wailsOrAPI<void>('AddRule', '/api/rules', postJSON('/api/rules', { ruleType, pattern, action, priority }), ruleType, pattern, action, priority)

export const UpdateRuleById = (id: number, ruleType: string, pattern: string, action: string, priority: number, enabled: boolean) =>
  wailsOrAPI<void>('UpdateRuleById', `/api/rules/${id}`, putJSON(`/api/rules/${id}`, { ruleType, pattern, action, priority, enabled }), id, ruleType, pattern, action, priority, enabled)

export const DeleteRule = (id: number) =>
  wailsOrAPI<void>('DeleteRule', `/api/rules/${id}`, { method: 'DELETE' }, id)

export const ToggleRule = (id: number, enabled: boolean) =>
  wailsOrAPI<void>('ToggleRule', '/api/rules/toggle', postJSON('/api/rules/toggle', { id, enabled }), id, enabled)

export const ClearAllRules = () =>
  wailsOrAPI<void>('ClearAllRules', '/api/rules/clear', { method: 'POST' })

export async function TestRule(domain: string, url: string, contentType: string): Promise<string> {
  if (isWails) {
    const { TestRule } = await import('../../wailsjs/go/main/App')
    return TestRule(domain, url, contentType)
  }
  const res = await fetchAPI<{ result: string }>('/api/rules/test', postJSON('/api/rules/test', { domain, url, contentType }))
  return res.result
}

export async function AddBulkRules(patterns: string[], action: string, priority: number): Promise<number> {
  if (isWails) {
    // Wails doesn't have AddBulkRules binding — add rules one by one
    const mod = await import('../../wailsjs/go/main/App')
    let count = 0
    for (const p of patterns) {
      try { await mod.AddRule('domain', p.trim(), action, priority); count++ } catch {}
    }
    return count
  }
  const res = await fetchAPI<{ count: number }>('/api/rules/bulk', postJSON('/api/rules/bulk', { patterns, action, priority }))
  return res.count
}

export async function ImportRules(data: string): Promise<number> {
  if (isWails) {
    const { ImportRules } = await import('../../wailsjs/go/main/App')
    return ImportRules(data)
  }
  const res = await fetchAPI<{ count: number }>('/api/rules/import', postJSON('/api/rules/import', { data }))
  return res.count
}

export async function ExportRules(): Promise<string> {
  if (isWails) {
    const { ExportRules } = await import('../../wailsjs/go/main/App')
    return ExportRules()
  }
  const res = await fetchAPI<{ data: string }>('/api/rules/export')
  return res.data
}

// === Proxies ===

export const GetProxies = () =>
  wailsOrAPI<any[]>('GetProxies', '/api/proxies')

export async function AddProxyAPI(address: string, username: string, password: string, type_: string, category: string): Promise<void> {
  if (isWails) {
    const { AddProxy } = await import('../../wailsjs/go/main/App')
    return AddProxy(address, username, password, type_, category)
  }
  await fetchAPI('/api/proxies', postJSON('/api/proxies', { address, username, password, type: type_, category }))
}

export const DeleteProxy = (id: number) =>
  wailsOrAPI<void>('DeleteProxy', `/api/proxies/${id}`, { method: 'DELETE' }, id)

export const ClearAllProxies = () =>
  wailsOrAPI<void>('ClearAllProxies', '/api/proxies/clear', { method: 'DELETE' })

export async function ImportProxies(text: string): Promise<number> {
  if (isWails) {
    const { ImportProxies } = await import('../../wailsjs/go/main/App')
    return ImportProxies(text)
  }
  const res = await fetchAPI<{ count: number }>('/api/proxies/import', postJSON('/api/proxies/import', { data: text }))
  return res.count
}

export const GetOutputProxies = () =>
  wailsOrAPI<any[]>('GetOutputProxies', '/api/proxies/output')

// === Stats ===

export const GetRealtimeStats = () =>
  wailsOrAPI<any>('GetRealtimeStats', '/api/stats/realtime')

export const GetCostSummary = () =>
  wailsOrAPI<any>('GetCostSummary', '/api/stats/cost')

export const GetBudgetStatus = () =>
  wailsOrAPI<any>('GetBudgetStatus', '/api/stats/budget')

// === Domain Stats ===

export const GetDomainStats = (period: string = '24h', proxyId: number = 0) =>
  fetchAPI<any[]>(`/api/stats/domains?period=${period}${proxyId > 0 ? `&proxyId=${proxyId}` : ''}`)

export const ClearDomainStats = () =>
  fetchAPI<void>('/api/stats/domains/clear', { method: 'POST' })

// === Cache ===

export const GetCacheStats = () =>
  wailsOrAPI<any>('GetCacheStats', '/api/cache/stats')

export const ClearCache = () =>
  wailsOrAPI<void>('ClearCache', '/api/cache/clear', { method: 'POST' })

// === Settings ===

export const GetSettings = () =>
  wailsOrAPI<Record<string, string>>('GetSettings', '/api/settings')

export const UpdateSetting = (key: string, value: string) =>
  wailsOrAPI<void>('UpdateSetting', '/api/settings', putJSON('/api/settings', { key, value }), key, value)

export async function GetVersion(): Promise<string> {
  try {
    const res = await fetchAPI<{ version: string }>('/api/version')
    return res.version
  } catch {
    return 'unknown'
  }
}

export async function GetCACertPath(): Promise<string> {
  if (isWails) {
    const { GetCACertPath } = await import('../../wailsjs/go/main/App')
    return GetCACertPath()
  }
  const res = await fetchAPI<{ path: string }>('/api/cert/path')
  return res.path
}
