interface FormFieldProps {
  label: string
  type: 'text' | 'password' | 'select' | 'toggle'
  value: string | boolean
  onChange: (value: string | boolean) => void
  placeholder?: string
  options?: { label: string; value: string }[]
  error?: string
  disabled?: boolean
}

export function FormField({
  label,
  type,
  value,
  onChange,
  placeholder,
  options,
  error,
  disabled,
}: FormFieldProps) {
  if (type === 'toggle') {
    const checked = typeof value === 'boolean' ? value : false
    return (
      <div className="flex items-center justify-between py-2">
        <label className="text-body text-primary">{label}</label>
        <button
          type="button"
          role="switch"
          aria-checked={checked}
          disabled={disabled}
          onClick={() => onChange(!checked)}
          className={`relative inline-flex h-[22px] w-[44px] items-center rounded-full transition-colors ${
            checked ? 'bg-accent' : 'bg-gray-300'
          } ${disabled ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}`}
        >
          <span
            className={`inline-block h-[18px] w-[18px] rounded-full bg-white transition-transform shadow-sm ${
              checked ? 'translate-x-[22px]' : 'translate-x-[2px]'
            }`}
          />
        </button>
        {error && <span className="text-caption text-warning ml-2">{error}</span>}
      </div>
    )
  }

  if (type === 'select') {
    return (
      <div className="py-1">
        <label className="text-caption text-secondary block mb-1">{label}</label>
        <select
          value={value as string}
          onChange={(e) => onChange(e.target.value)}
          disabled={disabled}
          className={`w-full px-3 py-2 rounded-[6px] border text-body text-primary transition-colors bg-white ${
            error
              ? 'border-warning focus:border-warning'
              : 'border-gray-200 focus:border-accent'
          } focus:outline-none ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
        >
          {options?.map((opt) => (
            <option key={opt.value} value={opt.value}>{opt.label}</option>
          ))}
        </select>
        {error && <span className="text-caption text-warning mt-1">{error}</span>}
      </div>
    )
  }

  return (
    <div className="py-1">
      <label className="text-caption text-secondary block mb-1">{label}</label>
      <input
        type={type}
        value={value as string}
        onChange={(e) => onChange(e.target.value)}
        placeholder={placeholder}
        disabled={disabled}
        className={`w-full px-3 py-2 rounded-[6px] border text-body text-primary transition-colors ${
          error
            ? 'border-warning focus:border-warning'
            : 'border-gray-200 focus:border-accent'
        } focus:outline-none ${disabled ? 'opacity-50 cursor-not-allowed' : ''}`}
      />
      {error && <span className="text-caption text-warning mt-1">{error}</span>}
    </div>
  )
}
