import { useState, useEffect, useCallback } from 'react'
import { ConnectionBridge } from '../components/ConnectionBridge'
import { StatusDot } from '../components/StatusDot'
import { useServiceStatus } from '../hooks/useServiceStatus'
import { App, AgentStatusInfo } from '../../bindings/github.com/chenhg5/cc-connect/client'
import type { PlatformHealth } from '../lib/types'

export function Dashboard() {
  const { status, errorMessage } = useServiceStatus()
  const [loading, setLoading] = useState(false)
  const [platformHealth, setPlatformHealth] = useState<PlatformHealth[]>([])
  const [agentStatus, setAgentStatus] = useState<AgentStatusInfo[]>([])

  // Poll platform health and agent status when service is running
  useEffect(() => {
    if (status !== 'running') {
      setPlatformHealth([])
      setAgentStatus([])
      return
    }

    const fetchData = async () => {
      try {
        const health = await App.GetPlatformHealth() as PlatformHealth[]
        const agents = await App.GetAgentStatus() as AgentStatusInfo[]
        setPlatformHealth(health || [])
        setAgentStatus(agents || [])
      } catch {
        // ignore fetch errors
      }
    }

    fetchData()
    const interval = setInterval(fetchData, 5000)
    return () => clearInterval(interval)
  }, [status])

  const handleAction = useCallback(async (action: () => Promise<void>) => {
    setLoading(true)
    try {
      await action()
    } catch (err) {
      console.error('Service action failed:', err)
    } finally {
      setLoading(false)
    }
  }, [])

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
          {errorMessage && (
            <span className="text-caption text-warning">{errorMessage}</span>
          )}
        </div>
        <div className="flex items-center gap-2">
          {status === 'idle' || status === 'error' ? (
            <button
              onClick={() => handleAction(App.StartService)}
              disabled={loading}
              className="px-4 py-2 rounded-sm text-body bg-accent text-white disabled:opacity-50 transition-colors"
            >
              {loading ? '启动中...' : '启动'}
            </button>
          ) : status === 'running' ? (
            <>
              <button
                onClick={() => handleAction(App.StopService)}
                disabled={loading}
                className="px-4 py-2 rounded-sm text-body bg-surface text-secondary border border-gray-200/50 disabled:opacity-50 transition-colors"
              >
                {loading ? '停止中...' : '停止'}
              </button>
              <button
                onClick={() => handleAction(App.RestartService)}
                disabled={loading}
                className="px-4 py-2 rounded-sm text-body bg-surface text-secondary border border-gray-200/50 disabled:opacity-50 transition-colors"
              >
                {loading ? '重启中...' : '重启'}
              </button>
            </>
          ) : null}
        </div>
      </div>

      <div className="grid grid-cols-2 gap-4">
        {/* Platform Health Card */}
        <div className="bg-surface rounded-lg p-card">
          <h3 className="text-headline text-primary mb-3">平台连接</h3>
          {status !== 'running' ? (
            <p className="text-caption text-secondary">服务启动后显示</p>
          ) : platformHealth.length === 0 ? (
            <p className="text-caption text-secondary">暂无平台数据</p>
          ) : (
            <div className="space-y-2">
              {platformHealth.map((p, i) => (
                <div key={i} className="flex items-center justify-between text-body">
                  <div className="flex items-center gap-2">
                    <StatusDot
                      status={p.connected ? 'active' : 'error'}
                      size={4}
                    />
                    <span className="text-primary">{p.platformName}</span>
                  </div>
                  <span className="text-secondary text-caption">
                    {p.connected ? '已连接' : '未连接'}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Agent Status Card */}
        <div className="bg-surface rounded-lg p-card">
          <h3 className="text-headline text-primary mb-3">Agent 状态</h3>
          {status !== 'running' ? (
            <p className="text-caption text-secondary">服务启动后显示</p>
          ) : agentStatus.length === 0 ? (
            <p className="text-caption text-secondary">暂无 Agent 数据</p>
          ) : (
            <div className="space-y-2">
              {agentStatus.map((a, i) => (
                <div key={i} className="flex items-center justify-between text-body">
                  <div className="flex items-center gap-2">
                    <StatusDot
                      status={a.status === 'running' ? 'active' : 'idle'}
                      size={4}
                    />
                    <span className="text-primary">{a.agentType}</span>
                    <span className="text-secondary text-caption">{a.projectName}</span>
                  </div>
                  <span className="text-secondary text-caption">
                    {a.status === 'running' ? '运行中' : '空闲'}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>
      </div>
    </div>
  )
}
