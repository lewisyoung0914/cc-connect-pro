import { useState, useEffect, useMemo } from 'react'
import { SegmentedControl } from '../components/SegmentedControl'
import { FormField } from '../components/FormField'
import { StatusDot } from '../components/StatusDot'
import { App, ConfigSummary, FeishuPlatformDetail, SaveFeishuConfigOpts, ProjectInfo } from '../../bindings/github.com/chenhg5/cc-connect/client'

interface FeishuConfigState {
  projectName: string
  platformType: string
  appId: string
  appSecret: string
  domain: string
  allowFrom: string
  allowChat: string
  groupOnly: boolean
  groupReplyAll: boolean
  shareSession: boolean
  threadIsolation: boolean
  reactionEmoji: string
  doneEmoji: string
  progressStyle: string
  enableCard: boolean
  resolveMentions: boolean
  port: string
  callbackPath: string
  encryptKey: string
  connectionMode: 'websocket' | 'webhook'
}

const defaultState: FeishuConfigState = {
  projectName: '',
  platformType: 'feishu',
  appId: '',
  appSecret: '',
  domain: '',
  allowFrom: '*',
  allowChat: '*',
  groupOnly: false,
  groupReplyAll: false,
  shareSession: false,
  threadIsolation: false,
  reactionEmoji: 'THINKING',
  doneEmoji: 'OK',
  progressStyle: 'legacy',
  enableCard: true,
  resolveMentions: false,
  port: '',
  callbackPath: '',
  encryptKey: '',
  connectionMode: 'websocket',
}

export function FeishuConfig() {
  const [projects, setProjects] = useState<ProjectInfo[]>([])
  const [selectedProject, setSelectedProject] = useState('')
  const [config, setConfig] = useState<FeishuConfigState>(defaultState)
  const [savedConfig, setSavedConfig] = useState<FeishuConfigState>(defaultState)
  const [validationState, setValidationState] = useState<'idle' | 'validating' | 'valid' | 'invalid'>('idle')
  const [validationError, setValidationError] = useState('')
  const [saving, setSaving] = useState(false)
  const [saveMessage, setSaveMessage] = useState<{ type: 'success' | 'error'; text: string } | null>(null)
  const [loading, setLoading] = useState(true)
  const [showValidationErrors, setShowValidationErrors] = useState(false)

  // Dirty state: compare current config vs saved config
  const isDirty = useMemo(() => {
    return JSON.stringify(config) !== JSON.stringify(savedConfig)
  }, [config, savedConfig])

  useEffect(() => {
    loadProjects()
  }, [])

  useEffect(() => {
    if (selectedProject) {
      loadConfigDetail(selectedProject)
    }
  }, [selectedProject])

  // Reset validation state when credential fields change
  useEffect(() => {
    setValidationState('idle')
    setValidationError('')
  }, [config.appId, config.appSecret])

  const loadProjects = async () => {
    setLoading(true)
    try {
      const summary = await App.GetConfigSummary() as ConfigSummary | null
      if (summary && summary.projects) {
        setProjects(summary.projects)
        const feishuProjects = summary.projects.filter((p: ProjectInfo) => p.hasFeishu)
        if (feishuProjects.length > 0) {
          setSelectedProject(feishuProjects[0].name)
        } else if (summary.projects.length > 0) {
          setSelectedProject(summary.projects[0].name)
        }
      }
    } catch (err) {
      console.error('Failed to load projects:', err)
    } finally {
      setLoading(false)
    }
  }

  const loadConfigDetail = async (projectName: string) => {
    try {
      const detail = await App.GetFeishuConfigDetail(projectName) as FeishuPlatformDetail | null
      if (detail) {
        const isWebhook = (detail.port && detail.port !== '') || (detail.encryptKey && detail.encryptKey !== '')
        const state: FeishuConfigState = {
          projectName: detail.projectName || projectName,
          platformType: detail.platformType || 'feishu',
          appId: detail.appId || '',
          appSecret: detail.appSecret || '',
          domain: detail.domain || '',
          allowFrom: detail.allowFrom || '*',
          allowChat: detail.allowChat || '*',
          groupOnly: detail.groupOnly || false,
          groupReplyAll: detail.groupReplyAll || false,
          shareSession: detail.shareSession || false,
          threadIsolation: detail.threadIsolation || false,
          reactionEmoji: detail.reactionEmoji || 'OnIt',
          doneEmoji: detail.doneEmoji || 'none',
          progressStyle: detail.progressStyle || 'legacy',
          enableCard: detail.enableCard !== undefined ? detail.enableCard : true,
          resolveMentions: detail.resolveMentions || false,
          port: detail.port || '',
          callbackPath: detail.callbackPath || '',
          encryptKey: detail.encryptKey || '',
          connectionMode: isWebhook ? 'webhook' : 'websocket',
        }
        setConfig(state)
        setSavedConfig(state)
      }
    } catch (err) {
      console.error('Failed to load config detail:', err)
    }
  }

  const handleValidateCredentials = async () => {
    setValidationState('validating')
    setValidationError('')
    try {
      const domain = config.domain || (config.platformType === 'lark' ? 'https://open.larksuite.com' : 'https://open.feishu.cn')
      await App.ValidateFeishuCredentials(config.appId, config.appSecret, domain)
      setValidationState('valid')
    } catch (err: any) {
      setValidationState('invalid')
      setValidationError(err?.message || '验证失败')
    }
  }

  const buildSaveOpts = (): SaveFeishuConfigOpts => {
    return new SaveFeishuConfigOpts({
      projectName: config.projectName,
      appId: config.appId,
      appSecret: config.appSecret,
      domain: config.domain,
      allowFrom: config.allowFrom,
      allowChat: config.allowChat,
      groupOnly: config.groupOnly,
      groupReplyAll: config.groupReplyAll,
      shareSession: config.shareSession,
      threadIsolation: config.threadIsolation,
      reactionEmoji: config.reactionEmoji,
      doneEmoji: config.doneEmoji,
      progressStyle: config.progressStyle,
      enableCard: config.enableCard,
      resolveMentions: config.resolveMentions,
      port: config.connectionMode === 'webhook' ? config.port : '',
      callbackPath: config.connectionMode === 'webhook' ? config.callbackPath : '',
      encryptKey: config.connectionMode === 'webhook' ? config.encryptKey : '',
    })
  }

  const handleSave = async () => {
    // Validate required fields before saving
    setShowValidationErrors(true)
    if (!config.appId.trim() || !config.appSecret.trim()) {
      setSaveMessage({ type: 'error', text: 'App ID 和 App Secret 为必填项' })
      return
    }
    if (config.connectionMode === 'webhook' && !config.port.trim()) {
      setSaveMessage({ type: 'error', text: 'Webhook 模式需要指定端口' })
      return
    }

    setSaving(true)
    setSaveMessage(null)
    try {
      await App.SaveFeishuConfig(buildSaveOpts())
      setSavedConfig({ ...config }) // Update saved state to match current
      setSaveMessage({ type: 'success', text: '配置已保存' })
      setShowValidationErrors(false)
    } catch (err: any) {
      setSaveMessage({ type: 'error', text: err?.message || '保存失败' })
    } finally {
      setSaving(false)
    }
  }

  const handleReset = async () => {
    setConfig({ ...savedConfig })
    setValidationState('idle')
    setValidationError('')
    setSaveMessage(null)
    setShowValidationErrors(false)
  }

  if (loading) {
    return (
      <div className="flex items-center justify-center h-64">
        <p className="text-body text-secondary">加载配置...</p>
      </div>
    )
  }

  if (projects.length === 0) {
    return (
      <div className="bg-surface rounded-lg p-card">
        <h2 className="text-title text-primary">飞书配置</h2>
        <p className="text-body text-secondary mt-2">暂无项目，请先创建项目</p>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      {/* Header: project selector + global actions */}
      <div className="bg-surface rounded-lg p-card">
        <div className="flex items-center justify-between">
          <div className="flex items-center gap-3">
            <h2 className="text-title text-primary">飞书配置</h2>
            {isDirty && (
              <span className="text-caption text-warning bg-warning/10 px-2 py-0.5 rounded-sm">未保存</span>
            )}
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={handleReset}
              disabled={!isDirty || saving}
              className="px-3 py-1.5 rounded-sm text-caption text-secondary border border-gray-200/50 disabled:opacity-50 transition-colors hover:text-primary"
            >
              重置
            </button>
            <button
              onClick={handleSave}
              disabled={saving || !isDirty}
              className="px-4 py-1.5 rounded-sm text-body bg-accent text-white disabled:opacity-50 transition-colors"
            >
              {saving ? '保存中...' : '保存配置'}
            </button>
          </div>
        </div>
        {saveMessage && (
          <p className={`text-caption mt-2 ${saveMessage.type === 'success' ? 'text-success' : 'text-warning'}`}>
            {saveMessage.text}
          </p>
        )}

        {/* Project selector - always visible */}
        <div className="mt-4 pt-3 border-t border-gray-200/50">
          <h3 className="text-headline text-primary mb-2">项目</h3>
          <div className="space-y-1">
            {projects.map((proj: ProjectInfo) => (
              <button
                key={proj.name}
                onClick={() => setSelectedProject(proj.name)}
                className={`flex items-center justify-between w-full px-3 py-2 rounded-sm text-body transition-colors ${
                  selectedProject === proj.name
                    ? 'bg-accent/10 text-primary'
                    : 'text-secondary hover:text-primary hover:bg-surface'
                }`}
              >
                <span>{proj.name} <span className="text-caption text-secondary">({proj.agentType})</span></span>
                <StatusDot
                  status={proj.hasFeishu ? 'active' : 'idle'}
                  label={proj.hasFeishu ? '已配置' : '未配置'}
                  size={8}
                />
              </button>
            ))}
          </div>
        </div>
      </div>

      {/* 飞书凭证 */}
      <div className="bg-surface rounded-lg p-card space-y-4">
        <h3 className="text-headline text-primary">飞书凭证</h3>

        <FormField
          label="平台类型"
          type="select"
          value={config.platformType}
          onChange={(v) => setConfig({ ...config, platformType: v as string })}
          options={[
            { label: '飞书 (Feishu)', value: 'feishu' },
            { label: 'Lark (国际版)', value: 'lark' },
          ]}
        />

        <FormField
          label="App ID"
          type="text"
          value={config.appId}
          onChange={(v) => setConfig({ ...config, appId: v as string })}
          placeholder="cli_xxxxxxxx"
          error={showValidationErrors && !config.appId.trim() ? '请输入 App ID' : undefined}
        />

        <FormField
          label="App Secret"
          type="password"
          value={config.appSecret}
          onChange={(v) => setConfig({ ...config, appSecret: v as string })}
          placeholder="飞书应用密钥"
          error={showValidationErrors && !config.appSecret.trim() ? '请输入 App Secret' : undefined}
        />

        <FormField
          label="Domain（可选）"
          type="select"
          value={config.domain || (config.platformType === 'lark' ? 'https://open.larksuite.com' : 'https://open.feishu.cn')}
          onChange={(v) => setConfig({ ...config, domain: v as string })}
          options={
            config.platformType === 'lark'
              ? [
                  { label: 'Lark 默认 (open.larksuite.com)', value: 'https://open.larksuite.com' },
                  { label: '自定义域名', value: '' },
                ]
              : [
                  { label: '飞书默认 (open.feishu.cn)', value: 'https://open.feishu.cn' },
                  { label: 'Lark (open.larksuite.com)', value: 'https://open.larksuite.com' },
                  { label: '自定义域名', value: '' },
                ]
          }
        />

        <div className="flex items-center gap-3">
          <button
            onClick={handleValidateCredentials}
            disabled={!config.appId || !config.appSecret || validationState === 'validating'}
            className="px-3 py-1.5 rounded-sm text-caption text-secondary border border-gray-200/50 disabled:opacity-50 transition-colors hover:text-primary"
          >
            {validationState === 'validating' ? '验证中...' : '验证凭证'}
          </button>
          {validationState === 'valid' && (
            <span className="text-caption text-success">✓ 验证成功</span>
          )}
          {validationState === 'invalid' && (
            <span className="text-caption text-warning">{validationError}</span>
          )}
        </div>
      </div>

      {/* 连接模式 */}
      <div className="bg-surface rounded-lg p-card space-y-4">
        <h3 className="text-headline text-primary">连接模式</h3>

        <SegmentedControl
          options={[
            { label: 'WebSocket', value: 'websocket' as const },
            { label: 'Webhook', value: 'webhook' as const },
          ]}
          value={config.connectionMode}
          onChange={(v) => setConfig({ ...config, connectionMode: v })}
        />

        {config.connectionMode === 'websocket' && (
          <p className="text-caption text-secondary">
            WebSocket 长连接模式，无需额外配置端口和回调路径。飞书 SDK 会自动建立连接。
          </p>
        )}

        {config.connectionMode === 'webhook' && (
          <div className="space-y-3 mt-3">
            <FormField
              label="端口"
              type="text"
              value={config.port}
              onChange={(v) => setConfig({ ...config, port: v as string })}
              placeholder="8080"
              error={showValidationErrors && config.connectionMode === 'webhook' && !config.port.trim() ? 'Webhook 需要指定端口' : undefined}
            />
            <FormField
              label="回调路径"
              type="text"
              value={config.callbackPath}
              onChange={(v) => setConfig({ ...config, callbackPath: v as string })}
              placeholder="/feishu/webhook"
            />
            <FormField
              label="Encrypt Key（可选）"
              type="password"
              value={config.encryptKey}
              onChange={(v) => setConfig({ ...config, encryptKey: v as string })}
              placeholder="事件加密密钥"
            />
          </div>
        )}
      </div>

      {/* 权限与行为 */}
      <div className="bg-surface rounded-lg p-card space-y-4">
        <h3 className="text-headline text-primary">权限与行为</h3>

        <FormField
          label="允许的用户 (allow_from)"
          type="text"
          value={config.allowFrom}
          onChange={(v) => setConfig({ ...config, allowFrom: v as string })}
          placeholder="* = 所有用户，或 open_id 逗号分隔"
        />

        <FormField
          label="允许的群聊 (allow_chat)"
          type="text"
          value={config.allowChat}
          onChange={(v) => setConfig({ ...config, allowChat: v as string })}
          placeholder="* = 所有群聊，或 chat_id 逗号分隔"
        />

        <div className="border-t border-gray-200/50 pt-3 space-y-1">
          <FormField
            label="仅响应群聊消息 (group_only)"
            type="toggle"
            value={config.groupOnly}
            onChange={(v) => setConfig({ ...config, groupOnly: v as boolean })}
          />
          <FormField
            label="群聊无需 @机器人 (group_reply_all)"
            type="toggle"
            value={config.groupReplyAll}
            onChange={(v) => setConfig({ ...config, groupReplyAll: v as boolean })}
          />
          <FormField
            label="群内共享会话 (share_session_in_channel)"
            type="toggle"
            value={config.shareSession}
            onChange={(v) => setConfig({ ...config, shareSession: v as boolean })}
          />
          <FormField
            label="话题隔离 (thread_isolation)"
            type="toggle"
            value={config.threadIsolation}
            onChange={(v) => setConfig({ ...config, threadIsolation: v as boolean })}
          />
          <FormField
            label="解析 @提及 (resolve_mentions)"
            type="toggle"
            value={config.resolveMentions}
            onChange={(v) => setConfig({ ...config, resolveMentions: v as boolean })}
          />
        </div>

        <div className="border-t border-gray-200/50 pt-3 space-y-3">
          <FormField
            label="收到消息表情 (reaction_emoji)"
            type="select"
            value={config.reactionEmoji}
            onChange={(v) => setConfig({ ...config, reactionEmoji: v as string })}
            options={[
              { label: '🤔 THINKING (默认)', value: 'THINKING' },
              { label: '👍 THUMBSUP', value: 'THUMBSUP' },
              { label: '❤️ HEART', value: 'HEART' },
              { label: '👏 CLAP', value: 'CLAP' },
              { label: '😂 LAUGH', value: 'LAUGH' },
              { label: '🎉 PARTY', value: 'PARTY' },
              { label: '👌 OK', value: 'OK' },
              { label: '😢 CRY', value: 'CRY' },
              { label: 'None (无表情)', value: 'none' },
            ]}
          />

          <FormField
            label="完成消息表情 (done_emoji)"
            type="select"
            value={config.doneEmoji}
            onChange={(v) => setConfig({ ...config, doneEmoji: v as string })}
            options={[
              { label: '👌 OK (默认)', value: 'OK' },
              { label: '👍 THUMBSUP', value: 'THUMBSUP' },
              { label: '❤️ HEART', value: 'HEART' },
              { label: '👏 CLAP', value: 'CLAP' },
              { label: '🎉 PARTY', value: 'PARTY' },
              { label: '😂 LAUGH', value: 'LAUGH' },
              { label: '🤔 THINKING', value: 'THINKING' },
              { label: '😢 CRY', value: 'CRY' },
              { label: 'None (无表情)', value: 'none' },
            ]}
          />

          <FormField
            label="进度展示风格 (progress_style)"
            type="select"
            value={config.progressStyle}
            onChange={(v) => setConfig({ ...config, progressStyle: v as string })}
            options={[
              { label: 'Legacy (逐条消息)', value: 'legacy' },
              { label: 'Compact (单消息更新)', value: 'compact' },
              { label: 'Card (结构化卡片)', value: 'card' },
            ]}
          />

          <FormField
            label="启用飞书卡片 (enable_feishu_card)"
            type="toggle"
            value={config.enableCard}
            onChange={(v) => setConfig({ ...config, enableCard: v as boolean })}
          />
        </div>
      </div>
    </div>
  )
}
