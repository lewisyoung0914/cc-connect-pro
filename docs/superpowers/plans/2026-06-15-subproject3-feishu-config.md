# 子项目 3：飞书凭证与工作区管理页面 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use subagent-driven-development to implement this plan.

**Goal:** 创建完整的飞书配置管理页面，包括工作区列表、飞书凭证编辑、权限与行为设置、连接模式切换、凭证验证、配置热更新。

**Architecture:** Go 侧新增 Feishu 配置读写和验证方法绑定到前端；前端创建 FeishuConfig.tsx 页面替换 PlaceholderPage。

**Tech Stack:** Wails v3 Go bindings, React, Tailwind CSS

---

## 文件结构

```
client/app.go                          ← 新增 GetFeishuConfigDetail, SaveFeishuConfig, ValidateFeishuCredentials 方法
client/frontend/src/pages/FeishuConfig.tsx   ← 新增飞书配置管理主页面
client/frontend/src/components/FormField.tsx  ← 新增表单字段组件（Apple 风格）
client/frontend/src/App.tsx            ← 替换 PlaceholderPage 为 FeishuConfig 页面
```

---

### Task 1: 在 app.go 中新增飞书配置相关方法和类型

**Files:**
- Modify: `client/app.go`

新增类型和方法：

```go
// FeishuPlatformDetail 返回给前端的飞书平台详细配置
type FeishuPlatformDetail struct {
    ProjectName       string `json:"projectName"`
    PlatformType      string `json:"platformType"`  // "feishu" or "lark"
    AppID             string `json:"appId"`
    AppSecret         string `json:"appSecret"`
    Domain            string `json:"domain"`
    AllowFrom         string `json:"allowFrom"`
    AllowChat         string `json:"allowChat"`
    GroupOnly         bool   `json:"groupOnly"`
    GroupReplyAll     bool   `json:"groupReplyAll"`
    ShareSession      bool   `json:"shareSession"`
    ThreadIsolation   bool   `json:"threadIsolation"`
    ReactionEmoji     string `json:"reactionEmoji"`
    DoneEmoji         string `json:"doneEmoji"`
    ProgressStyle     string `json:"progressStyle"`
    EnableCard        bool   `json:"enableCard"`
    ResolveMentions   bool   `json:"resolveMentions"`
    // Webhook 模式字段
    Port             string `json:"port"`
    CallbackPath     string `json:"callbackPath"`
    EncryptKey       string `json:"encryptKey"`
}

// SaveFeishuConfigOpts 保存飞书配置的选项
type SaveFeishuConfigOpts struct {
    ProjectName   string `json:"projectName"`
    AppID         string `json:"appId"`
    AppSecret     string `json:"appSecret"`
    Domain        string `json:"domain"`
    AllowFrom     string `json:"allowFrom"`
    AllowChat     string `json:"allowChat"`
    GroupOnly     bool   `json:"groupOnly"`
    ReactionEmoji string `json:"reactionEmoji"`
    // 其他字段按需...
    Port          string `json:"port"`
    CallbackPath  string `json:"callbackPath"`
    EncryptKey    string `json:"encryptKey"`
}

// GetFeishuConfigDetail 获取指定项目的飞书平台详细配置
func (a *App) GetFeishuConfigDetail(projectName string) (*FeishuPlatformDetail, error)

// SaveFeishuConfig 保存飞书配置（如果服务运行中则触发热更新）
func (a *App) SaveFeishuConfig(opts SaveFeishuConfigOpts) error

// ValidateFeishuCredentials 验证飞书 app_id/app_secret 是否有效
func (a *App) ValidateFeishuCredentials(appId, appSecret, domain string) error
```

实现要点：
- GetFeishuConfigDetail: 加载 config，找到指定项目，找飞书/lark platform，提取所有 options 字段
- SaveFeishuConfig: 用 config.SaveFeishuPlatformCredentials 写入凭证，然后用 config 包的其他方法写入权限/行为字段。如果服务运行中，调用 service.Restart() 触发热更新
- ValidateFeishuCredentials: 使用飞书 SDK 的 lark.Client 创建临时客户端，调用 `tenant.GetTenantInfoByAppID` 或类似 API 验证凭证有效性

---

### Task 2: 创建 FormField 组件

**Files:**
- Create: `client/frontend/src/components/FormField.tsx`

Apple 风格表单字段组件：
- label 在上方，caption 字体
- input: 1px border, 6px radius, focus → accent border, 15px body 字体
- 支持 text/password/select/checkbox 类型
- error 消息 inline 显示，warning 颜色

---

### Task 3: 创建 FeishuConfig 页面

**Files:**
- Create: `client/frontend/src/pages/FeishuConfig.tsx`

页面内容：
- **工作区列表区**: 用 GetConfigSummary 显示所有项目，每个项目卡片显示名称、Agent类型、飞书状态（StatusDot），点击选择后进入编辑
- **选中项目后的编辑区**:
  - 飞书凭证卡片: app_id (text), app_secret (password), domain (select: 飞书/Lark)
  - 验证按钮: 调用 ValidateFeishuCredentials，显示 ✓ 或 ✗ 结果
  - 连接模式: SegmentedControl (WebSocket / Webhook)，Webhook 时显示 port, callback_path, encrypt_key
  - 权限与行为卡片: allow_from, allow_chat (text), group_only, group_reply_all, share_session, thread_isolation (checkbox), reaction_emoji, done_emoji (select), progress_style (select: legacy/compact/card), enable_feishu_card (checkbox), resolve_mentions (checkbox)
- **新建工作区按钮**: 调用 CreateProjectWithFeishu + StartService
- **保存按钮**: 调用 SaveFeishuConfig

布局：苹果 Settings 风格，左侧列表选择项目，右侧编辑区。如果只有1个项目，直接显示编辑区不显示列表。

---

### Task 4: 替换 App.tsx 中的 PlaceholderPage

**Files:**
- Modify: `client/frontend/src/App.tsx`

将 `/feishu` 路由从 PlaceholderPage 替换为 FeishuConfig 页面。

---

### Task 5: 验证和 Commit

- go vet, go test 通过
- 前端 npm run build 通过
- git commit
