import { forwardRef, type InputHTMLAttributes, type TextareaHTMLAttributes, type SelectHTMLAttributes, type ReactNode } from 'react'

/* ============================================
   TEXT INPUT
   ============================================ */
interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  label?: string
  error?: string
}

const inputBase = [
  'w-full rounded-[var(--radius-lg)] px-3 py-1.5 text-xs',
  'bg-[var(--color-input-bg)] text-[var(--color-text-primary)]',
  'border border-[var(--color-input-border)]',
  'outline-none transition-colors duration-150',
  'focus:border-[var(--color-input-focus)] focus:ring-1 focus:ring-[var(--color-input-focus)]',
  'placeholder:text-[var(--color-text-muted)]',
  'disabled:opacity-50 disabled:cursor-not-allowed',
].join(' ')

export const Input = forwardRef<HTMLInputElement, InputProps>(
  ({ label, error, id, className = '', ...props }, ref) => {
    const inputId = id || label?.toLowerCase().replace(/\s+/g, '-')
    return (
      <div className="w-full">
        {label && (
          <label htmlFor={inputId} className="block text-xs text-[var(--color-text-secondary)] mb-1">
            {label}
          </label>
        )}
        <input
          ref={ref}
          id={inputId}
          className={`${inputBase} ${error ? 'border-[var(--color-danger)]' : ''} ${className}`}
          aria-invalid={!!error}
          aria-describedby={error ? `${inputId}-error` : undefined}
          {...props}
        />
        {error && (
          <p id={`${inputId}-error`} className="text-[10px] text-[var(--color-danger-text)] mt-1" role="alert">
            {error}
          </p>
        )}
      </div>
    )
  }
)
Input.displayName = 'Input'

/* ============================================
   NUMBER INPUT
   ============================================ */
interface NumberInputProps {
  value: number
  onChange: (v: number) => void
  min?: number
  max?: number
  step?: number
  label?: string
}

export function NumberInput({ value, onChange, min, max, step, label }: NumberInputProps) {
  const id = label?.toLowerCase().replace(/\s+/g, '-')
  return (
    <div className="w-full">
      {label && (
        <label htmlFor={id} className="block text-xs text-[var(--color-text-secondary)] mb-1">
          {label}
        </label>
      )}
      <input
        id={id}
        type="number"
        value={value}
        onChange={(e) => onChange(Number(e.target.value))}
        min={min}
        max={max}
        step={step}
        className={`${inputBase} tabular-nums`}
      />
    </div>
  )
}

/* ============================================
   SELECT INPUT
   ============================================ */
interface SelectInputProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, 'onChange'> {
  value: string
  onChange: (v: string) => void
  options: { value: string; label: string }[]
  label?: string
}

export function SelectInput({ value, onChange, options, label, ...props }: SelectInputProps) {
  const id = label?.toLowerCase().replace(/\s+/g, '-')
  return (
    <div className="w-full">
      {label && (
        <label htmlFor={id} className="block text-xs text-[var(--color-text-secondary)] mb-1">
          {label}
        </label>
      )}
      <select
        id={id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={inputBase}
        {...props}
      >
        {options.map((o) => <option key={o.value} value={o.value}>{o.label}</option>)}
      </select>
    </div>
  )
}

/* ============================================
   TEXTAREA INPUT
   ============================================ */
interface TextAreaProps extends Omit<TextareaHTMLAttributes<HTMLTextAreaElement>, 'onChange'> {
  value: string
  onChange: (v: string) => void
  label?: string
}

export function TextArea({ value, onChange, label, className = '', ...props }: TextAreaProps) {
  const id = label?.toLowerCase().replace(/\s+/g, '-')
  return (
    <div className="w-full">
      {label && (
        <label htmlFor={id} className="block text-xs text-[var(--color-text-secondary)] mb-1">
          {label}
        </label>
      )}
      <textarea
        id={id}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className={`${inputBase} resize-none ${className}`}
        {...props}
      />
    </div>
  )
}
