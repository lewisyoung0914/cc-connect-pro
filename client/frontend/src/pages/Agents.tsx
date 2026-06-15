import { useState, useEffect, useCallback } from 'react'
import { StatusDot } from '../components/StatusDot'

// Placeholder Wails bindings - will be replaced by auto-generated ones
const ListSessions = async (): Promise<SessionInfo[]> => []
const GetTaskQueue = async (_projectName: string): Promise<TaskInfo[]> => []
const StopSession = async (_sessionKey: string): Promise<void> => {}
const GetAgentStatus = async (): Promise<AgentStatusInfo[]> => []

interface SessionInfo {
  sessionKey: string
  agentSessionId: string
  agentType: string
  projectName: string
  platform: string
  chatName: string
  userName: string
  status: string
  createdAt: string
  updatedAt: string
  queueDepth: number
}

interface TaskInfo {
  platform: string
  userName: string
  content: string
  queueTime: string
}

interface AgentStatusInfo {
  projectName: string
  agentType: string
  status: string
}

function relativeTime(dateStr: string): string {
  if (!dateStr) return ''
  const d = new Date(dateStr)
  const now = new Date()
  const diffMs = now.getTime() - d.getTime()
  const diffMin = Math.floor(diffMs / 60000)
  if (diffMin < 1) return '刚刚'
  if (diffMin < 60) return `${diffMin}分钟前`
  const diffH = Math.floor(diffMin / 60)
  if (diffH < 24) return `${diffH}小时前`
  const diffD = Math.floor(diffH / 24)
  return `${diffD}天前`
}

function truncate(s: string, max: number): string {
  if (!s) return ''
  return s.length > max ? s.slice(0, max) + '...' : s
}

export function Agents() {
  const [sessions, setSessions] = useState<SessionInfo[]>([])
  const [agentStatuses, setAgentStatuses] = useState<AgentStatusInfo[]>([])
  const [selectedSession, setSelectedSession] = useState<SessionInfo | null>(null)
  const [selectedProjectTasks, setSelectedProjectTasks] = useState<TaskInfo[]>([])
  const [stopping, setStopping] = useState(false)
  const [showStopConfirm, setShowStopConfirm] = useState(false)

  const refresh = useCallback(async () => {
    try {
      const sess = await ListSessions()
      setSessions(sess)
      const statuses = await GetAgentStatus()
      setAgentStatuses(statuses)
      // If a session is selected, refresh its project's task queue
      if (selectedSession) {
        const tasks = await GetTaskQueue(selectedSession.projectName)
        setSelectedProjectTasks(tasks)
      }
    } catch (err) {
      console.error('Failed to refresh sessions:', err)
    }
  }, [selectedSession])

  useEffect(() => {
    refresh()
    const interval = setInterval(refresh, 3000)
    return () => clearInterval(interval)
  }, [refresh])

  const handleSelectSession = async (session: SessionInfo) => {
    setSelectedSession(session)
    try {
      const tasks = await GetTaskQueue(session.projectName)
      setSelectedProjectTasks(tasks)
    } catch (err) {
      console.error('Failed to load task queue:', err)
    }
  }

  const handleStopSession = async () => {
    if (!selectedSession) return
    setStopping(true)
    try {
      await StopSession(selectedSession.sessionKey)
      setShowStopConfirm(false)
      setSelectedSession(null)
      setSelectedProjectTasks([])
      await refresh()
    } catch (err) {
      console.error('Failed to stop session:', err)
    } finally {
      setStopping(false)
    }
  }

  // Aggregate tasks for selected project
  const projectTasks = selectedProjectTasks

  return (
    <div className="flex h-full">
      {/* Main content area */}
      <div className="flex-1 overflow-auto">
        <div className="space-y-6">
          {/* Agent Status Summary */}
          <div className="bg-surface rounded-lg p-card">
            <h2 className="text-title text-primary mb-3">Agent 状态</h2>
            {agentStatuses.length === 0 ? (
              <p className="text-caption text-secondary">服务未运行或无项目配置</p>
            ) : (
              <div className="flex flex-wrap gap-3">
                {agentStatuses.map((a) => (
                  <div key={a.projectName} className="bg-white rounded-sm px-3 py-2 flex items-center gap-2">
                    <StatusDot
                      status={a.status === 'running' ? 'active' : a.status === 'error' ? 'error' : 'idle'}
                    />
                    <span className="text-body text-primary">{a.projectName}</span>
                    <span className="text-caption text-secondary">{a.agentType}</span>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Session List */}
          <div className="bg-surface rounded-lg p-card">
            <h2 className="text-title text-primary mb-3">活跃会话</h2>
            {sessions.length === 0 ? (
              <p className="text-caption text-secondary">无活跃交互会话</p>
            ) : (
              <div className="divide-y divide-gray-200/50">
                {sessions.map((s) => (
                  <div
                    key={s.sessionKey}
                    onClick={() => handleSelectSession(s)}
                    className={`flex items-center gap-3 py-3 px-2 cursor-pointer transition-colors ${
                      selectedSession?.sessionKey === s.sessionKey
                        ? 'bg-accent/5'
                        : 'hover:bg-white'
                    }`}
                  >
                    {/* Status dot */}
                    <StatusDot
                      status={s.status === 'active' ? 'active' : s.status === 'error' ? 'error' : 'idle'}
                      size={6}
                    />

                    {/* Session key (mono) */}
                    <span className="font-mono text-caption text-secondary w-[180px] truncate" title={s.sessionKey}>
                      {truncate(s.sessionKey, 24)}
                    </span>

                    {/* Agent type label */}
                    <span className="text-caption text-primary bg-white rounded-sm px-2 py-0.5 border border-gray-200/50">
                      {s.agentType}
                    </span>

                    {/* Platform + Chat name */}
                    <span className="text-body text-secondary w-[120px] truncate">
                      {s.platform}{s.chatName ? ` / ${s.chatName}` : ''}
                    </span>

                    {/* User name */}
                    {s.userName && (
                      <span className="text-caption text-secondary w-[80px] truncate">
                        {s.userName}
                      </span>
                    )}

                    {/* Project name */}
                    <span className="text-caption text-secondary">
                      {s.projectName}
                    </span>

                    {/* Queue depth badge */}
                    {s.queueDepth > 0 && (
                      <span className="ml-1 bg-warning text-white text-mini rounded-full px-1.5 py-0.5 min-w-[20px] text-center">
                        {s.queueDepth}
                      </span>
                    )}

                    {/* Created time */}
                    <span className="text-caption text-secondary ml-auto">
                      {relativeTime(s.createdAt)}
                    </span>
                  </div>
                ))}
              </div>
            )}
          </div>

          {/* Task Queue Section (only visible when there are tasks) */}
          {projectTasks.length > 0 && selectedSession && (
            <div className="bg-surface rounded-lg p-card">
              <h2 className="text-title text-primary mb-3">
                任务队列 <span className="text-caption text-secondary">- {selectedSession.projectName}</span>
              </h2>
              <div className="divide-y divide-gray-200/50">
                {projectTasks.map((t, i) => (
                  <div key={i} className="flex items-center gap-3 py-2 px-2">
                    <span className="text-caption text-secondary bg-white rounded-sm px-2 py-0.5 border border-gray-200/50">
                      {t.platform}
                    </span>
                    <span className="text-body text-primary w-[80px] truncate">
                      {t.userName}
                    </span>
                    <span className="text-body text-secondary flex-1 truncate">
                      {t.content}
                    </span>
                    <span className="text-caption text-secondary">
                      {t.queueTime}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}
        </div>
      </div>

      {/* Session Detail Panel (Apple Mail style slide-in) */}
      {selectedSession && (
        <div className="w-[280px] bg-surface border-l border-gray-200/50 p-card flex flex-col overflow-auto">
          <h3 className="text-headline text-primary mb-4">会话详情</h3>

          <div className="space-y-3">
            <div>
              <span className="text-caption text-secondary">Session Key</span>
              <p className="font-mono text-caption text-primary mt-0.5 break-all">{selectedSession.sessionKey}</p>
            </div>

            <div>
              <span className="text-caption text-secondary">Agent Session ID</span>
              <p className="font-mono text-caption text-primary mt-0.5">{selectedSession.agentSessionId || '-'}</p>
            </div>

            <div>
              <span className="text-caption text-secondary">Agent 类型</span>
              <p className="text-body text-primary mt-0.5">{selectedSession.agentType}</p>
            </div>

            <div>
              <span className="text-caption text-secondary">平台</span>
              <p className="text-body text-primary mt-0.5">{selectedSession.platform}</p>
            </div>

            {selectedSession.chatName && (
              <div>
                <span className="text-caption text-secondary">对话名称</span>
                <p className="text-body text-primary mt-0.5">{selectedSession.chatName}</p>
              </div>
            )}

            {selectedSession.userName && (
              <div>
                <span className="text-caption text-secondary">用户</span>
                <p className="text-body text-primary mt-0.5">{selectedSession.userName}</p>
              </div>
            )}

            <div>
              <span className="text-caption text-secondary">项目</span>
              <p className="text-body text-primary mt-0.5">{selectedSession.projectName}</p>
            </div>

            <div>
              <span className="text-caption text-secondary">状态</span>
              <p className="text-body text-primary mt-0.5">
                <StatusDot
                  status={selectedSession.status === 'active' ? 'active' : selectedSession.status === 'error' ? 'error' : 'idle'}
                  label={selectedSession.status === 'active' ? '活跃' : selectedSession.status === 'error' ? '错误' : '空闲'}
                />
              </p>
            </div>

            <div>
              <span className="text-caption text-secondary">队列深度</span>
              <p className="text-body text-primary mt-0.5">{selectedSession.queueDepth}</p>
            </div>

            <div>
              <span className="text-caption text-secondary">创建时间</span>
              <p className="text-caption text-primary mt-0.5">{selectedSession.createdAt || '-'}</p>
            </div>

            <div>
              <span className="text-caption text-secondary">更新时间</span>
              <p className="text-caption text-primary mt-0.5">{selectedSession.updatedAt || '-'}</p>
            </div>
          </div>

          {/* Actions Toolbar */}
          <div className="mt-6 space-y-2">
            {showStopConfirm ? (
              <div className="bg-white rounded-lg p-3 border border-red-200">
                <p className="text-body text-primary mb-2">确认停止此会话？</p>
                <div className="flex gap-2">
                  <button
                    onClick={handleStopSession}
                    disabled={stopping}
                    className="px-3 py-1.5 rounded-sm text-caption bg-red-500 text-white disabled:opacity-50"
                  >
                    {stopping ? '停止中...' : '确认停止'}
                  </button>
                  <button
                    onClick={() => setShowStopConfirm(false)}
                    className="px-3 py-1.5 rounded-sm text-caption bg-surface text-secondary border border-gray-200/50"
                  >
                    取消
                  </button>
                </div>
              </div>
            ) : (
              <button
                onClick={() => setShowStopConfirm(true)}
                className="w-full px-3 py-2 rounded-sm text-body bg-surface text-secondary border border-gray-200/50 hover:border-red-300 hover:text-red-500 transition-colors"
              >
                停止会话
              </button>
            )}
          </div>
        </div>
      )}
    </div>
  )
}
