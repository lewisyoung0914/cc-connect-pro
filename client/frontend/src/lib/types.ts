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
