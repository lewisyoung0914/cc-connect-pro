import { HashRouter, Routes, Route } from 'react-router-dom'
import { SidebarLayout } from './layouts/SidebarLayout'
import { Dashboard } from './pages/Dashboard'
import { Welcome } from './pages/Welcome'
import { useServiceStatus } from './hooks/useServiceStatus'
import { useState, useEffect } from 'react'

// Placeholder Wails bindings - will be replaced by auto-generated ones
const HasConfig = async (): Promise<boolean> => { return false }

export default function App() {
  const [hasConfig, setHasConfig] = useState<boolean | null>(null)
  const status = useServiceStatus()

  useEffect(() => {
    HasConfig().then(setHasConfig)
  }, [])

  if (hasConfig === null) {
    return (
      <div className="flex items-center justify-center h-screen bg-white">
        <p className="text-body text-secondary">正在检查配置...</p>
      </div>
    )
  }

  if (!hasConfig) {
    return <Welcome onConfigCreated={() => setHasConfig(true)} />
  }

  return (
    <HashRouter>
      <SidebarLayout status={status}>
        <Routes>
          <Route path="/" element={<Dashboard />} />
          <Route path="/feishu" element={<PlaceholderPage title="飞书配置" desc="子项目 3 将实现完整的飞书凭证与工作区管理页面" />} />
          <Route path="/agents" element={<PlaceholderPage title="Agent 管理" desc="子项目 4 将实现完整的 Agent 状态与任务队列页面" />} />
          <Route path="/monitor" element={<PlaceholderPage title="监控" desc="子项目 5 将实现完整的健康监控页面" />} />
        </Routes>
      </SidebarLayout>
    </HashRouter>
  )
}

function PlaceholderPage({ title, desc }: { title: string; desc: string }) {
  return (
    <div className="bg-surface rounded-lg p-6">
      <h2 className="text-title text-primary">{title}</h2>
      <p className="text-body text-secondary mt-2">{desc}</p>
    </div>
  )
}
