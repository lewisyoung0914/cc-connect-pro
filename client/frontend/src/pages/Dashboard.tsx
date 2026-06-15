import { useState } from 'react'
import { ConnectionBridge } from '../components/ConnectionBridge'
import { StatusDot } from '../components/StatusDot'
import { useServiceStatus } from '../hooks/useServiceStatus'

// Placeholder - will be replaced by Wails auto-generated bindings
const StartService = async () => { console.log('StartService called') }
const StopService = async () => { console.log('StopService called') }
const RestartService = async () => { console.log('RestartService called') }

export function Dashboard() {
  const status = useServiceStatus()
  const [loading, setLoading] = useState(false)

  const handleAction = async (action: () => Promise<void>) => {
    setLoading(true)
    try {
      await action()
    } catch (err) {
      console.error('Service action failed:', err)
    } finally {
      setLoading(false)
    }
  }

  const statusLabel =
    status === 'running' ? '运行中' :
    status === 'starting' ? '启动中' :
    status === 'stopping' ? '停止中' :
    status === 'error' ? '错误' : '已停止'

  const dotStatus =
    status === 'running' ? 'active' :
    status === 'error' ? 'error' : 'idle'

  return (
    <div className="space-y-6">
      <div className="bg-surface rounded-lg p-card">
        <ConnectionBridge status={status} />
      </div>

      <div className="bg-surface rounded-lg p-card flex items-center justify-between">
        <div className="flex items-center gap-3">
          <StatusDot status={dotStatus} label={statusLabel} />
          <span className="font-mono text-body text-secondary">{status}</span>
        </div>
        <div className="flex items-center gap-2">
          {status === 'idle' || status === 'error' ? (
            <button
              onClick={() => handleAction(StartService)}
              disabled={loading}
              className="px-4 py-2 rounded-sm text-body bg-accent text-white disabled:opacity-50 transition-colors"
            >
              启动
            </button>
          ) : status === 'running' ? (
            <>
              <button
                onClick={() => handleAction(StopService)}
                disabled={loading}
                className="px-4 py-2 rounded-sm text-body bg-surface text-secondary border border-gray-200/50 disabled:opacity-50 transition-colors"
              >
                停止
              </button>
              <button
                onClick={() => handleAction(RestartService)}
                disabled={loading}
                className="px-4 py-2 rounded-sm text-body bg-surface text-secondary border border-gray-200/50 disabled:opacity-50 transition-colors"
              >
                重启
              </button>
            </>
          ) : null}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        <div className="bg-surface rounded-lg p-card">
          <h3 className="text-headline text-primary">平台连接</h3>
          <p className="text-caption text-secondary mt-2">服务启动后显示</p>
        </div>
        <div className="bg-surface rounded-lg p-card">
          <h3 className="text-headline text-primary">Agent 状态</h3>
          <p className="text-caption text-secondary mt-2">服务启动后显示</p>
        </div>
      </div>
    </div>
  )
}
