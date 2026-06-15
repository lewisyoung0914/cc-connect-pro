interface SegmentedControlProps<T extends string> {
  options: { label: string; value: T }[]
  value: T
  onChange: (value: T) => void
}

export function SegmentedControl<T extends string>({ options, value, onChange }: SegmentedControlProps<T>) {
  return (
    <div className="inline-flex rounded-sm bg-surface p-1">
      {options.map((opt) => (
        <button
          key={opt.value}
          className={`px-4 py-1.5 rounded-sm text-body transition-colors ${
            opt.value === value ? 'bg-accent text-white' : 'text-secondary hover:text-primary'
          }`}
          onClick={() => onChange(opt.value)}
        >
          {opt.label}
        </button>
      ))}
    </div>
  )
}
