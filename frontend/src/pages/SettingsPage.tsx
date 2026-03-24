import { useEffect, useState, cloneElement, isValidElement } from 'react'
import { Server, Database, DollarSign, Zap, FileText, Shield, Cog, RotateCcw, Lock } from 'lucide-react'
import { useSettingsStore } from '../stores/settingsStore'
import { useToastStore } from '../stores/toastStore'
import { Toggle } from '../components/ui/Toggle'
import { ConfirmDialog } from '../components/ui/ConfirmDialog'
import { ClearCache, GetCACertPath, GetCacheStats } from '../lib/api'
import { SettingsSkeleton } from '../components/ui/Skeleton'

function Section({ icon: Icon, title, desc, children }: {
  icon: typeof Server; title: string; desc: string; children: React.ReactNode
}) {
  return (
    <div className="bg-[var(--color-bg-surface)] border border-[var(--color-border)] rounded-xl p-5">
      <div className="flex items-center gap-2 mb-1">
        <Icon size={16} className="text-[var(--color-primary)]" />
        <h2 className="text-sm font-semibold text-[var(--color-text-primary)]">{title}</h2>
      </div>
      <p className="text-xs text-[var(--color-text-muted)] mb-4">{desc}</p>
      <div className="space-y-4">{children}</div>
    </div>
  )
}

function toId(label: string) {
  return `setting-${label.toLowerCase().replace(/[^a-z0-9]+/g, '-')}`
}

function Field({ label, children }: { label: string; children: React.ReactNode }) {
  const id = toId(label)
  return (
    <div className="flex items-center justify-between">
      <label htmlFor={id} className="text-xs text-[var(--color-text-secondary)]">{label}</label>
      <div className="w-48">
        {isValidElement(children) ? cloneElement(children as React.ReactElement<any>, { id }) : children}
      </div>
    </div>
  )
}

function NumberInput({ value, onChange, min, max, step, id }: {
  value: number; onChange: (v: number) => void; min?: number; max?: number; step?: number; id?: string
}) {
  return (
    <input
      id={id}
      type="number"
      value={value}
      onChange={(e) => onChange(Number(e.target.value))}
      min={min}
      max={max}
      step={step}
      className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-1.5 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] tabular-nums"
    />
  )
}

function SelectInput({ value, onChange, options, id }: {
  value: string; onChange: (v: string) => void; options: { value: string; label: string }[]; id?: string
}) {
  return (
    <select
      id={id}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-1.5 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)]"
    >
      {options.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
    </select>
  )
}

function TextInput({ value, onChange, placeholder, type = 'text', id }: {
  value: string; onChange: (v: string) => void; placeholder?: string; type?: string; id?: string
}) {
  return (
    <input
      id={id}
      type={type}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-1.5 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)]"
    />
  )
}

function TextAreaInput({ value, onChange, placeholder, rows = 3, id }: {
  value: string; onChange: (v: string) => void; placeholder?: string; rows?: number; id?: string
}) {
  return (
    <textarea
      id={id}
      value={value}
      onChange={(e) => onChange(e.target.value)}
      placeholder={placeholder}
      rows={rows}
      className="w-full bg-[var(--color-input-bg)] border border-[var(--color-input-border)] rounded-[var(--radius-lg)] px-3 py-1.5 text-xs text-[var(--color-text-primary)] outline-none focus:border-[var(--color-input-focus)] resize-none"
    />
  )
}

export default function SettingsPage() {
  const { settings, loading, initialize, get, update } = useSettingsStore()
  const addToast = useToastStore((state) => state.addToast)
  const [clearConfirm, setClearConfirm] = useState(false)
  const [cacheInfo, setCacheInfo] = useState({ entries: 0, memoryUsedMb: 0, diskUsedMb: 0 })

  useEffect(() => {
    initialize()
    GetCacheStats().then((stats) => setCacheInfo(stats)).catch(() => {})
  }, [initialize])

  const getString = (key: string, fallback: string) => get(key, fallback)
  const getNumber = (key: string, fallback: number) => Number(get(key, String(fallback))) || fallback
  const getBool = (key: string, fallback: boolean) => get(key, String(fallback)) === 'true'

  const setSetting = (key: string, value: string | number | boolean) => update(key, String(value))

  const handleClearCache = async () => {
    try {
      await ClearCache()
      addToast('success', 'Cache cleared')
      setCacheInfo({ entries: 0, memoryUsedMb: 0, diskUsedMb: 0 })
    } catch (e) {
      addToast('error', `Failed to clear cache: ${e}`)
    }
  }

  const handleExportCA = async () => {
    try {
      const path = await GetCACertPath()
      addToast('success', `CA cert: ${path}`)
    } catch (e) {
      addToast('error', `${e}`)
    }
  }

  if (loading) {
    return (
      <div className="p-6 space-y-4">
        {[...Array(4)].map((_, i) => <SettingsSkeleton key={i} />)}
      </div>
    )
  }

  return (
    <div className="p-6 overflow-y-auto h-full space-y-4">
      <h1 className="text-lg font-semibold text-[var(--color-text-primary)]">Settings</h1>

      {/* Proxy Server */}
      <Section icon={Server} title="Proxy Server" desc="Local proxy server configuration. Changes require restart.">
        <Field label="HTTP Port">
          <NumberInput value={getNumber('http_port', 8888)} onChange={(v) => setSetting('http_port', v)} min={1024} max={65535} />
        </Field>
        <Field label="SOCKS5 Port">
          <NumberInput value={getNumber('socks5_port', 8889)} onChange={(v) => setSetting('socks5_port', v)} min={1024} max={65535} />
        </Field>
        <Field label="Output Base Port">
          <NumberInput value={getNumber('base_port', 30000)} onChange={(v) => setSetting('base_port', v)} min={1024} max={65535} />
        </Field>
        <Field label="Bind Address">
          <SelectInput
            value={getString('bind_address', '127.0.0.1')}
            onChange={(v) => setSetting('bind_address', v)}
            options={[
              { value: '127.0.0.1', label: 'localhost only' },
              { value: '0.0.0.0', label: 'All interfaces' },
            ]}
          />
        </Field>
      </Section>

      {/* Proxy Authentication */}
      <Section icon={Lock} title="Proxy Authentication" desc="Protect proxy ports with username/password and IP whitelist. Changes require proxy restart.">
        <div className="bg-[var(--color-danger-bg)] border border-[var(--color-danger)]/30 rounded-[var(--radius-lg)] p-3 text-xs text-[var(--color-danger-text)] mb-2">
          If binding to 0.0.0.0 (all interfaces), you MUST enable authentication or IP whitelist to prevent unauthorized access.
        </div>
        <Field label="Require Username/Password">
          <Toggle checked={getBool('proxy_auth_enabled', false)} onChange={(v) => setSetting('proxy_auth_enabled', v)} />
        </Field>
        {getBool('proxy_auth_enabled', false) && (
          <>
            <Field label="Username">
              <TextInput value={getString('proxy_username', '')} onChange={(v) => setSetting('proxy_username', v)} placeholder="proxy user" />
            </Field>
            <Field label="Password">
              <TextInput value={getString('proxy_password', '')} onChange={(v) => setSetting('proxy_password', v)} placeholder="proxy password" type="password" />
            </Field>
          </>
        )}
        <div className="pt-3 border-t border-[var(--color-border)]" />
        <Field label="Enable IP Whitelist">
          <Toggle checked={getBool('ip_whitelist_enabled', false)} onChange={(v) => setSetting('ip_whitelist_enabled', v)} />
        </Field>
        {getBool('ip_whitelist_enabled', false) && (
          <div>
            <label className="text-xs text-[var(--color-text-muted)] mb-1 block">Allowed IPs (comma-separated, supports CIDR e.g. 192.168.1.0/24)</label>
            <TextAreaInput
              value={getString('ip_whitelist', '')}
              onChange={(v) => setSetting('ip_whitelist', v)}
              placeholder="192.168.1.100, 10.0.0.0/8, 172.16.0.0/12"
              rows={3}
            />
            <p className="text-[10px] text-[var(--color-text-muted)] mt-1">127.0.0.1 and ::1 (localhost) are always allowed.</p>
          </div>
        )}
      </Section>

      {/* Cache */}
      <Section icon={Database} title="Cache" desc="In-memory and on-disk cache for reducing repeated requests.">
        <Field label="Memory Limit (MB)">
          <NumberInput value={getNumber('cache_memory_mb', 512)} onChange={(v) => setSetting('cache_memory_mb', v)} min={64} max={4096} step={64} />
        </Field>
        <Field label="Disk Limit (MB)">
          <NumberInput value={getNumber('cache_disk_mb', 2048)} onChange={(v) => setSetting('cache_disk_mb', v)} min={256} max={20480} step={256} />
        </Field>
        <Field label="Default TTL (minutes)">
          <NumberInput value={getNumber('cache_default_ttl', 60)} onChange={(v) => setSetting('cache_default_ttl', v)} min={1} max={10080} />
        </Field>
        <div className="pt-2 border-t border-[var(--color-border)]">
          <div className="flex items-center justify-between text-xs text-[var(--color-text-muted)] mb-3">
            <span>{cacheInfo.entries} entries | {cacheInfo.memoryUsedMb.toFixed(1)} MB memory | {cacheInfo.diskUsedMb.toFixed(1)} MB disk</span>
          </div>
          <button
            onClick={() => setClearConfirm(true)}
            className="px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-danger-bg)] text-[var(--color-danger-text)] border border-[var(--color-danger)]/30 hover:bg-[var(--color-danger)] hover:text-white transition-colors"
          >
            Clear All Cache
          </button>
        </div>
      </Section>

      {/* Budget */}
      <Section icon={DollarSign} title="Budget" desc="Monthly bandwidth budget and cost tracking.">
        <Field label="Monthly Budget (GB)">
          <NumberInput value={getNumber('monthly_budget_gb', 50)} onChange={(v) => setSetting('monthly_budget_gb', v)} min={0} max={10000} />
        </Field>
        <Field label="Cost per GB ($)">
          <NumberInput value={getNumber('cost_per_gb', 5)} onChange={(v) => setSetting('cost_per_gb', v)} min={0} max={100} step={0.01} />
        </Field>
        <Field label="Auto-pause at limit">
          <Toggle checked={getBool('auto_pause', false)} onChange={(v) => setSetting('auto_pause', v)} />
        </Field>
      </Section>

      {/* Optimization */}
      <Section icon={Zap} title="Optimization" desc="Request and response optimization settings.">
        <Field label="Strip unnecessary headers">
          <Toggle checked={getBool('strip_headers', true)} onChange={(v) => setSetting('strip_headers', v)} />
        </Field>
        <Field label="Enforce Accept-Encoding">
          <Toggle checked={getBool('enforce_encoding', true)} onChange={(v) => setSetting('enforce_encoding', v)} />
        </Field>
        <Field label="Recompress responses">
          <Toggle checked={getBool('recompress', false)} onChange={(v) => setSetting('recompress', v)} />
        </Field>
        <Field label="HTML Minification">
          <Toggle checked={getBool('minify_html', false)} onChange={(v) => setSetting('minify_html', v)} />
        </Field>
      </Section>

      {/* Logging */}
      <Section icon={FileText} title="Logging" desc="Log storage and retention settings.">
        <Field label="Log Level">
          <SelectInput
            value={getString('log_level', 'info')}
            onChange={(v) => setSetting('log_level', v)}
            options={[
              { value: 'debug', label: 'Debug' },
              { value: 'info', label: 'Info' },
              { value: 'warn', label: 'Warning' },
              { value: 'error', label: 'Error' },
            ]}
          />
        </Field>
        <Field label="Retention (days)">
          <NumberInput value={getNumber('log_retention_days', 7)} onChange={(v) => setSetting('log_retention_days', v)} min={1} max={90} />
        </Field>
      </Section>

      {/* TLS/MITM */}
      <Section icon={Shield} title="HTTPS Inspection" desc="Man-in-the-Middle for HTTPS traffic inspection and caching.">
        <div className="bg-[var(--color-warning-bg)] border border-[var(--color-warning)]/30 rounded-[var(--radius-lg)] p-3 text-xs text-[var(--color-warning-text)] mb-2">
          Enabling MITM allows full HTTPS inspection and caching. You must install and trust the generated CA certificate in your OS.
        </div>
        <Field label="Enable MITM">
          <Toggle checked={getBool('mitm_enabled', false)} onChange={(v) => setSetting('mitm_enabled', v)} />
        </Field>
        <div className="pt-2">
          <button
            onClick={handleExportCA}
            className="px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-secondary)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors"
          >
            Show CA Certificate Path
          </button>
        </div>
      </Section>

      {/* Advanced */}
      <Section icon={Cog} title="Advanced" desc="Connection limits and advanced settings.">
        <Field label="Max Connections">
          <NumberInput value={getNumber('max_connections', 500)} onChange={(v) => setSetting('max_connections', v)} min={10} max={10000} />
        </Field>
        <Field label="Connection Timeout (s)">
          <NumberInput value={getNumber('conn_timeout', 30)} onChange={(v) => setSetting('conn_timeout', v)} min={5} max={300} />
        </Field>
        <Field label="Idle Timeout (s)">
          <NumberInput value={getNumber('idle_timeout', 90)} onChange={(v) => setSetting('idle_timeout', v)} min={10} max={600} />
        </Field>
        <Field label="DNS via Proxy (prevent leaks)">
          <Toggle checked={getBool('dns_via_proxy', true)} onChange={(v) => setSetting('dns_via_proxy', v)} />
        </Field>
        <div className="pt-3 border-t border-[var(--color-border)]">
          <button
            onClick={() => addToast('warning', 'Reset all settings is not implemented yet')}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs rounded-[var(--radius-lg)] bg-[var(--color-bg-elevated)] text-[var(--color-text-muted)] hover:bg-[var(--color-sidebar-hover)] border border-[var(--color-border)] transition-colors"
          >
            <RotateCcw size={12} /> Reset All Settings
          </button>
        </div>
      </Section>

      {/* Clear Cache Confirm */}
      <ConfirmDialog
        open={clearConfirm}
        onClose={() => setClearConfirm(false)}
        onConfirm={handleClearCache}
        title="Clear All Cache"
        message="This will delete all cached responses from memory and disk. This cannot be undone."
        confirmText="Clear Cache"
        destructive
      />
    </div>
  )
}
