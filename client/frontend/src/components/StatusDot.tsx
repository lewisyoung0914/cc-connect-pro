interface StatusDotProps {
  status: 'active' | 'idle' | 'error' | 'warning'
  label?: string
  size?: number
}

const colorMap = {
  active: 'bg-success',
  idle: 'bg-secondary',
  error: 'bg-red-500',
  warning: 'bg-warning',
}

export function StatusDot({ status, label, size = 6 }: StatusDotProps) {
  return (
    <span className="inline-flex items-center gap-2">
      <span className={`${colorMap[status]} rounded-full`} style={{ width: size, height: size }} />
      {label && <span className="text-caption text-secondary">{label}</span>}
    </span>
  )
}
