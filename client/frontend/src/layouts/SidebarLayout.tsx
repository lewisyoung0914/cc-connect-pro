import { NavLink } from 'react-router-dom'
import { Activity, MessageSquare, Bot, Monitor } from 'lucide-react'
import type { ServiceStatus } from '../lib/types'
import { StatusDot } from '../components/StatusDot'

interface SidebarLayoutProps {
  status: ServiceStatus
  children: React.ReactNode
}

const navItems = [
  { path: '/', label: '总览', icon: Activity },
  { path: '/feishu', label: '飞书配置', icon: MessageSquare },
  { path: '/agents', label: 'Agent 管理', icon: Bot },
  { path: '/monitor', label: '监控', icon: Monitor },
]

export function SidebarLayout({ status, children }: SidebarLayoutProps) {
  return (
    <div className="flex h-screen bg-white">
      <aside className="w-[220px] bg-surface flex flex-col border-r border-gray-200/50">
        <div className="px-5 pt-6 pb-4">
          <div className="text-large-title font-semibold text-primary">cc-connect</div>
          <div className="mt-2">
            <StatusDot
              status={status === 'running' ? 'active' : status === 'error' ? 'error' : 'idle'}
              label={status === 'running' ? '运行中' : status === 'error' ? '错误' : '已停止'}
            />
          </div>
        </div>
        <nav className="flex-1 px-2">
          {navItems.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                `flex items-center gap-3 px-3 py-2 rounded-sm my-0.5 text-body transition-colors ${
                  isActive
                    ? 'bg-accent/10 text-primary'
                    : 'text-secondary hover:text-primary hover:bg-surface'
                }`
              }
            >
              <item.icon size={18} />
              {item.label}
            </NavLink>
          ))}
        </nav>
        <div className="px-5 py-4 text-mini text-secondary">v0.1.0</div>
      </aside>
      <main className="flex-1 p-8 overflow-auto">{children}</main>
    </div>
  )
}
