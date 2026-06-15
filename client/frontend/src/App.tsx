import { HashRouter, Routes, Route } from 'react-router-dom'
import { SidebarLayout } from './layouts/SidebarLayout'
import { Dashboard } from './pages/Dashboard'
import { Welcome } from './pages/Welcome'
import { FeishuConfig } from './pages/FeishuConfig'
import { Agents } from './pages/Agents'
import { Monitor } from './pages/Monitor'
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
          <Route path="/feishu" element={<FeishuConfig />} />
          <Route path="/agents" element={<Agents />} />
          <Route path="/monitor" element={<Monitor />} />
        </Routes>
      </SidebarLayout>
    </HashRouter>
  )
}
