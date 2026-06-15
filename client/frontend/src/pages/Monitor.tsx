import { useState, useEffect, useRef, useCallback } from 'react'
import { StatusDot } from '../components/StatusDot'

// Placeholder Wails bindings - will be replaced by auto-generated ones
const GetProcessInfo = async (): Promise<ProcessInfo> => ({
  pid: 0, uptime: '', memoryMB: 0, goroutines: 0
})
const GetPlatformHealth = async (): Promise<PlatformHealth[]> => []
const GetRecentLogs = async (_count: number): Promise<LogEntry[]> => []
const RunDoctorChecks = async (): Promise<DoctorCheckResult[]> => []

interface ProcessInfo {
  pid: number
  uptime: string
  memoryMB: number
  goroutines: number
}

interface PlatformHealth {
  platformName: string
  projectName: string
  connected: boolean
  reconnectCount: number
  messagesSent: number
  messagesReceived: number
}

interface LogEntry {
  level: string
  message: string
  time: string
  source?: string
}

interface DoctorCheckResult {
  name: string
  passed: boolean
  detail: string
}

const levelColors: Record<string, string> = {
  ERROR: 'bg-red-500',
  WARN: 'bg-warning',
  INFO: 'bg-secondary',
  DEBUG: 'bg-gray-300',
}

export function Monitor() {
  const [processInfo, setProcessInfo] = useState<ProcessInfo>({
    pid: 0, uptime: '', memoryMB: 0, goroutines: 0
  })
  const [platforms, setPlatforms] = useState<PlatformHealth[]>([])
  const [logs, setLogs] = useState<LogEntry[]>([])
  const [logCount, setLogCount] = useState(100)
  const [autoScroll, setAutoScroll] = useState(true)
  const [doctorResults, setDoctorResults] = useState<DoctorCheckResult[]>([])
  const [doctorRunning, setDoctorRunning] = useState(false)
  const logContainerRef = useRef<HTMLDivElement>(null)

  // Fetch process info periodically
  useEffect(() => {
    const fetchProcessInfo = async () => {
      try {
        const info = await GetProcessInfo()
        setProcessInfo(info)
      } catch {
        // ignore errors
      }
    }
    fetchProcessInfo()
    const interval = setInterval(fetchProcessInfo, 3000)
    return () => clearInterval(interval)
  }, [])

  // Fetch platform health periodically
  useEffect(() => {
    const fetchHealth = async () => {
      try {
        const health = await GetPlatformHealth()
        setPlatforms(health || [])
      } catch {
        // ignore errors
      }
    }
    fetchHealth()
    const interval = setInterval(fetchHealth, 5000)
    return () => clearInterval(interval)
  }, [])

  // Fetch logs periodically
  useEffect(() => {
    const fetchLogs = async () => {
      try {
        const entries = await GetRecentLogs(logCount)
        setLogs(entries || [])
      } catch {
        // ignore errors
      }
    }
    fetchLogs()
    const interval = setInterval(fetchLogs, 1000)
    return () => clearInterval(interval)
  }, [logCount])

  // Auto-scroll to bottom when new logs arrive
  useEffect(() => {
    if (autoScroll && logContainerRef.current) {
      logContainerRef.current.scrollTop = logContainerRef.current.scrollHeight
    }
  }, [logs, autoScroll])

  // Handle user scroll to pause/resume auto-scroll
  const handleScroll = useCallback(() => {
    if (!logContainerRef.current) return
    const { scrollTop, scrollHeight, clientHeight } = logContainerRef.current
    const isNearBottom = scrollHeight - scrollTop - clientHeight < 30
    setAutoScroll(isNearBottom)
  }, [])

  const handleRunDoctor = async () => {
    setDoctorRunning(true)
    try {
      const results = await RunDoctorChecks()
      setDoctorResults(results || [])
    } catch {
      // ignore errors
    } finally {
      setDoctorRunning(false)
    }
  }

  return (
    <div className="space-y-6">
      {/* Process Info */}
      <div className="bg-surface rounded-lg p-card">
        <h3 className="text-headline text-primary mb-4">进程信息</h3>
        <div className="grid grid-cols-2 gap-x-8 gap-y-3">
          <div className="flex items-center justify-between">
            <span className="text-caption text-secondary">PID</span>
            <span className="font-mono text-body text-primary">{processInfo.pid}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-caption text-secondary">运行时间</span>
            <span className="font-mono text-body text-primary">{processInfo.uptime || '--'}</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-caption text-secondary">内存</span>
            <span className="font-mono text-body text-primary">{processInfo.memoryMB.toFixed(1)} MB</span>
          </div>
          <div className="flex items-center justify-between">
            <span className="text-caption text-secondary">Goroutines</span>
            <span className="font-mono text-body text-primary">{processInfo.goroutines}</span>
          </div>
        </div>
      </div>

      {/* Platform Health */}
      <div className="bg-surface rounded-lg p-card">
        <h3 className="text-headline text-primary mb-4">平台连接状态</h3>
        {platforms.length === 0 ? (
          <p className="text-caption text-secondary">服务启动后显示</p>
        ) : (
          <div className="space-y-3">
            {platforms.map((p) => (
              <div key={`${p.projectName}-${p.platformName}`} className="flex items-center justify-between">
                <div className="flex items-center gap-3">
                  <StatusDot
                    status={p.connected ? 'active' : 'error'}
                    size={8}
                  />
                  <span className="text-body text-primary">{p.platformName}</span>
                  <span className="text-caption text-secondary">({p.projectName})</span>
                </div>
                <div className="flex items-center gap-4 text-caption text-secondary">
                  <span>重连: {p.reconnectCount}</span>
                  <span>发送: {p.messagesSent}</span>
                  <span>接收: {p.messagesReceived}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Log Viewer */}
      <div className="bg-surface rounded-lg p-card">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-headline text-primary">日志</h3>
          <div className="flex items-center gap-2">
            {logCount === 100 ? (
              <button
                onClick={() => setLogCount(500)}
                className="px-3 py-1 rounded-sm text-caption bg-accent/10 text-accent transition-colors"
              >
                加载更多
              </button>
            ) : (
              <span className="text-caption text-secondary">显示 500 条</span>
            )}
            {!autoScroll && (
              <button
                onClick={() => setAutoScroll(true)}
                className="px-3 py-1 rounded-sm text-caption bg-accent/10 text-accent transition-colors"
              >
                回到底部
              </button>
            )}
          </div>
        </div>
        <div
          ref={logContainerRef}
          onScroll={handleScroll}
          className="h-[300px] overflow-y-auto rounded-sm bg-white border border-gray-200/50"
        >
          {logs.length === 0 ? (
            <p className="text-caption text-secondary p-4">暂无日志</p>
          ) : (
            <div className="p-2 space-y-0.5">
              {logs.map((log, i) => (
                <div key={i} className="flex items-start gap-2 text-mini leading-tight">
                  <span className="font-mono text-secondary shrink-0 w-[52px]">{log.time}</span>
                  <span
                    className={`${levelColors[log.level] || 'bg-secondary'} rounded-full shrink-0 mt-[3px]`}
                    style={{ width: 6, height: 6 }}
                  />
                  <span className={
                    log.level === 'ERROR' ? 'text-red-500' :
                    log.level === 'WARN' ? 'text-warning' :
                    'text-primary'
                  }>
                    {log.message}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>

      {/* Doctor Checks */}
      <div className="bg-surface rounded-lg p-card">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-headline text-primary">诊断检查</h3>
          <button
            onClick={handleRunDoctor}
            disabled={doctorRunning}
            className="px-4 py-2 rounded-sm text-body bg-accent text-white disabled:opacity-50 transition-colors"
          >
            {doctorRunning ? '检查中...' : '运行检查'}
          </button>
        </div>
        {doctorResults.length === 0 ? (
          <p className="text-caption text-secondary">点击"运行检查"开始诊断</p>
        ) : (
          <div className="space-y-2">
            {doctorResults.map((r, i) => (
              <div key={i} className="flex items-center gap-3">
                <span className={`text-body ${r.passed ? 'text-success' : 'text-red-500'}`}>
                  {r.passed ? '\u2713' : '\u2717'}
                </span>
                <span className="text-body text-primary">{r.name}</span>
                <span className="text-caption text-secondary ml-auto">{r.detail}</span>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  )
}
