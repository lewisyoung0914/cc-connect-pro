import { useState, useEffect } from 'react'

// Placeholder - will be replaced by Wails auto-generated bindings
const ListRegisteredAgents = async (): Promise<string[]> => { console.log('ListRegisteredAgents called'); return [] }
const CreateProjectWithFeishu = async (opts: {
  projectName: string
  agentType: string
  appId: string
  appSecret: string
  workDir: string
}): Promise<void> => { console.log('CreateProjectWithFeishu called', opts) }
const StartService = async () => { console.log('StartService called') }

interface WelcomeProps {
  onConfigCreated: () => void
}

export function Welcome({ onConfigCreated }: WelcomeProps) {
  const [appId, setAppId] = useState('')
  const [appSecret, setAppSecret] = useState('')
  const [projectName, setProjectName] = useState('')
  const [agentType, setAgentType] = useState('')
  const [workDir, setWorkDir] = useState('')
  const [agents, setAgents] = useState<string[]>([])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    ListRegisteredAgents().then(setAgents).catch(() => setAgents([]))
  }, [])

  const canSubmit = projectName.trim() && appId.trim() && appSecret.trim() && agentType

  const handleSubmit = async () => {
    setError('')
    setLoading(true)
    try {
      await CreateProjectWithFeishu({
        projectName,
        agentType,
        appId,
        appSecret,
        workDir,
      })
      await StartService()
      onConfigCreated()
    } catch (err: any) {
      setError(err?.message || '创建失败')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-white">
      <div className="w-[400px] space-y-6">
        <div className="text-center">
          <h1 className="text-large-title font-semibold text-primary">欢迎使用 cc-connect</h1>
          <p className="text-body text-secondary mt-2">配置飞书应用凭证，即可开始使用</p>
        </div>

        <div className="space-y-4">
          <div>
            <label className="text-caption text-secondary block mb-1">App ID</label>
            <input
              type="text"
              value={appId}
              onChange={(e) => setAppId(e.target.value)}
              placeholder="cli_xxxxxxxx"
              className="w-full px-3 py-2 rounded-sm border border-gray-200 text-body text-primary focus:border-accent focus:outline-none transition-colors"
            />
          </div>

          <div>
            <label className="text-caption text-secondary block mb-1">App Secret</label>
            <input
              type="password"
              value={appSecret}
              onChange={(e) => setAppSecret(e.target.value)}
              placeholder="飞书应用密钥"
              className="w-full px-3 py-2 rounded-sm border border-gray-200 text-body text-primary focus:border-accent focus:outline-none transition-colors"
            />
          </div>

          <div>
            <label className="text-caption text-secondary block mb-1">项目名称</label>
            <input
              type="text"
              value={projectName}
              onChange={(e) => setProjectName(e.target.value)}
              placeholder="my-project"
              className="w-full px-3 py-2 rounded-sm border border-gray-200 text-body text-primary focus:border-accent focus:outline-none transition-colors"
            />
          </div>

          <div>
            <label className="text-caption text-secondary block mb-1">Agent 类型</label>
            <select
              value={agentType}
              onChange={(e) => setAgentType(e.target.value)}
              className="w-full px-3 py-2 rounded-sm border border-gray-200 text-body text-primary focus:border-accent focus:outline-none transition-colors bg-white"
            >
              <option value="">选择 Agent 类型</option>
              {agents.map((agent) => (
                <option key={agent} value={agent}>{agent}</option>
              ))}
            </select>
          </div>

          <div>
            <label className="text-caption text-secondary block mb-1">工作目录（可选）</label>
            <input
              type="text"
              value={workDir}
              onChange={(e) => setWorkDir(e.target.value)}
              placeholder="自定义工作目录路径"
              className="w-full px-3 py-2 rounded-sm border border-gray-200 text-body text-primary focus:border-accent focus:outline-none transition-colors"
            />
          </div>
        </div>

        {error && (
          <p className="text-caption text-warning">{error}</p>
        )}

        <button
          onClick={handleSubmit}
          disabled={!canSubmit || loading}
          className="w-full py-2.5 rounded-sm text-body bg-accent text-white disabled:opacity-50 transition-colors"
        >
          {loading ? '创建中...' : '创建并启动'}
        </button>
      </div>
    </div>
  )
}
