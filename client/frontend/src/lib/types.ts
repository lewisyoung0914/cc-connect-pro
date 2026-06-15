export type ServiceStatus = 'idle' | 'starting' | 'running' | 'stopping' | 'error'

export interface ConfigSummary {
  configPath: string
  dataDir: string
  language: string
  projects: ProjectInfo[]
}

export interface ProjectInfo {
  name: string
  agentType: string
  workDir: string
  hasFeishu: boolean
}

export interface ServiceError {
  message: string
  timestamp: number
}

export interface ProcessInfo {
  pid: number
  uptime: string
  memoryMB: number
  goroutines: number
}

export interface PlatformHealth {
  platformName: string
  projectName: string
  connected: boolean
  reconnectCount: number
  messagesSent: number
  messagesReceived: number
}

export interface LogEntry {
  level: string
  message: string
  time: string
  source?: string
}

export interface DoctorCheckResult {
  name: string
  passed: boolean
  detail: string
}
