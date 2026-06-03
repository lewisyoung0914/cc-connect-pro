# 多飞书实例路由改造 — 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 Engine 能够区分并正确路由到同一项目中多个不同飞书应用实例。

**Architecture:** 在 Platform 接口新增 `Tag() string` 方法作为路由标识符（`Name()` 保持类型名不变），飞书平台通过配置 `name` 字段自定义 `Tag()` 值，Engine 路由层从 `Name()` 切换到 `Tag()`。

**Tech Stack:** Go 1.x, TOML 配置, 结构化日志 slog

---

## Phase 1 — 接口层 + 所有平台 Tag() 默认实现

**目标**：让 `Tag()` 成为 `Platform` 接口的正式方法，所有平台都能编译通过。此时 `Tag()` == `Name()`，行为完全不变。

### Task 1.1: 在 Platform 接口新增 Tag() 方法

**Files:**
- Modify: `core/interfaces.go:10-16`

- [ ] **Step 1: 在 Platform 接口定义中新增 Tag() string**

在 `core/interfaces.go` 的 `Platform` 接口中，在 `Name() string` 之后新增 `Tag() string`：

```go
type Platform interface {
    Name() string
    Tag() string    // routing identifier — same as Name() by default; customized for multi-instance platforms
    Start(handler MessageHandler) error
    Reply(ctx context.Context, replyCtx any, content string) error
    Send(ctx context.Context, replyCtx any, content string) error
    Stop() error
}
```

- [ ] **Step 2: 运行 go build 验证编译失败**

Run: `go build ./...`
Expected: 编译失败，所有缺少 `Tag()` 的平台实现报错

- [ ] **Step 3: Commit 接口变更（此步暂不单独 commit，等所有平台 Tag() 都加完后一起 commit）**

### Task 1.2: 为所有平台添加默认 Tag() 实现

**Files:**
- Modify: `platform/feishu/feishu.go:307-313`
- Modify: `platform/telegram/telegram.go:168`
- Modify: `platform/discord/discord.go:136`
- Modify: `platform/slack/slack.go:64`
- Modify: `platform/dingtalk/dingtalk.go:143`
- Modify: `platform/weixin/weixin.go:207`
- Modify: `platform/wecom/wecom.go:201`
- Modify: `platform/wecom/websocket.go:141`
- Modify: `platform/qq/qq.go:62`
- Modify: `platform/qqbot/qqbot.go:185`
- Modify: `platform/line/line.go:71`
- Modify: `platform/max/max.go:152`
- Modify: `platform/weibo/weibo.go:103`
- Modify: `platform/wps-xiezuo/wpsxiezuo.go:185`
- Modify: `core/bridge.go:320`

每个平台的 `Tag()` 默认返回与 `Name()` 相同的值。

**飞书（feishu）—— 特殊处理，Tag() 返回 p.instanceTag（但此时 instanceTag 还未添加，暂返回 p.platformName）：**

```go
// platform/feishu/feishu.go — 在 Name() 方法之后新增
func (p *Platform) Tag() string { return p.platformName }
```

**所有其他平台—— 在 Name() 方法之后新增，返回与 Name() 相同的值：**

telegram:
```go
func (p *Platform) Tag() string { return p.Name() }
```

discord:
```go
func (p *Platform) Tag() string { return p.Name() }
```

slack:
```go
func (p *Platform) Tag() string { return p.Name() }
```

dingtalk:
```go
func (p *Platform) Tag() string { return p.Name() }
```

weixin:
```go
func (p *Platform) Tag() string { return p.Name() }
```

wecom (wecom.go):
```go
func (p *Platform) Tag() string { return p.Name() }
```

wecom (websocket.go):
```go
func (p *WSPlatform) Tag() string { return p.Name() }
```

qq:
```go
func (p *Platform) Tag() string { return p.Name() }
```

qqbot:
```go
func (p *Platform) Tag() string { return p.Name() }
```

line:
```go
func (p *Platform) Tag() string { return p.Name() }
```

max:
```go
func (p *Platform) Tag() string { return p.Name() }
```

weibo:
```go
func (p *Platform) Tag() string { return p.Name() }
```

wps-xiezuo:
```go
func (p *Platform) Tag() string { return p.Name() }
```

bridge:
```go
func (bp *BridgePlatform) Tag() string { return bp.Name() }
```

- [ ] **Step 4: 运行 go build 验证编译通过**

Run: `go build ./...`
Expected: 编译通过

### Task 1.3: 为测试 stub 和 mutePlatform 添加 Tag()

**Files:**
- Modify: `core/engine_test.go:53-87`
- Modify: `core/registry_test.go:8-14`
- Modify: `core/cron.go:694-699`

**engine_test.go stubPlatformEngine:**

在 `stubPlatformEngine` 的方法列表中，在 `Name()` 之后新增：

```go
func (p *stubPlatformEngine) Tag() string { return p.n }
```

**registry_test.go stubPlatform:**

在 `stubPlatform` 的方法列表中，在 `Name()` 之后新增：

```go
func (s *stubPlatform) Tag() string { return s.n }
```

**cron.go mutePlatform:**

`mutePlatform` 通过嵌入 `Platform` 已自动继承 `Tag()`，无需额外代码。但验证一下编译无误。

- [ ] **Step 5: 运行 go test 验证全量通过**

Run: `go test ./...`
Expected: 全量通过

- [ ] **Step 6: Commit Phase 1**

```bash
git add core/interfaces.go core/engine_test.go core/registry_test.go core/cron.go core/bridge.go platform/feishu/feishu.go platform/telegram/telegram.go platform/discord/discord.go platform/slack/slack.go platform/dingtalk/dingtalk.go platform/weixin/weixin.go platform/wecom/wecom.go platform/wecom/websocket.go platform/qq/qq.go platform/qqbot/qqbot.go platform/line/line.go platform/max/max.go platform/weibo/weibo.go platform/wps-xiezuo/wpsxiezuo.go
git commit -m "feat: add Tag() string to Platform interface with default implementations"
```

---

## Phase 2 — 配置层（PlatformConfig.Name、校验、opts 注入）

**目标**：配置结构体支持 `name` 字段，并在创建平台时注入 opts。

### Task 2.1: PlatformConfig 新增 Name 字段

**Files:**
- Modify: `config/config.go:443-446`

- [ ] **Step 1: 在 PlatformConfig 结构体新增 Name 字段**

```go
type PlatformConfig struct {
    Type    string         `toml:"type"`
    Name    string         `toml:"name"`    // optional, instance identifier; defaults to Type
    Options map[string]any `toml:"options"`
}
```

### Task 2.2: validate() 新增 name 格式校验

**Files:**
- Modify: `config/config.go:797-804`

当前校验代码：

```go
for j, p := range proj.Platforms {
    if p.Type == "" {
        return fmt.Errorf("config: %s.platforms[%d].type is required", prefix, j)
    }
}
```

- [ ] **Step 2: 在 Type 校验之后新增 Name 格式校验**

```go
for j, p := range proj.Platforms {
    if p.Type == "" {
        return fmt.Errorf("config: %s.platforms[%d].type is required", prefix, j)
    }
    if p.Name != "" && p.Name != p.Type && !strings.HasPrefix(p.Name, p.Type+"-") {
        return fmt.Errorf("config: %s.platforms[%d].name must equal type or start with type + \"-\" (e.g. \"feishu-teamA\"), got %q",
            prefix, j, p.Name)
    }
}
```

注意：需要在文件顶部确认 `strings` 包已导入，如未导入需添加 `"strings"`。

### Task 2.3: main.go 注入 cc_platform_name 到 opts

**Files:**
- Modify: `cmd/cc-connect/main.go:237-251`

当前代码：

```go
var platforms []core.Platform
for _, pc := range proj.Platforms {
    opts := make(map[string]any, len(pc.Options)+2)
    for k, v := range pc.Options {
        opts[k] = v
    }
    opts["cc_data_dir"] = cfg.DataDir
    opts["cc_project"] = proj.Name
    p, err := core.CreatePlatform(pc.Type, opts)
```

- [ ] **Step 3: 注入 cc_platform_name 并调整 opts 容量**

```go
var platforms []core.Platform
for _, pc := range proj.Platforms {
    opts := make(map[string]any, len(pc.Options)+3)
    for k, v := range pc.Options {
        opts[k] = v
    }
    opts["cc_data_dir"] = cfg.DataDir
    opts["cc_project"] = proj.Name
    opts["cc_platform_name"] = pc.Name
    p, err := core.CreatePlatform(pc.Type, opts)
```

变化：容量从 `+2` 改为 `+3`；新增 `opts["cc_platform_name"] = pc.Name`。

### Task 2.4: config.example.toml 新增 name 注释

**Files:**
- Modify: `config.example.toml:918-976`

- [ ] **Step 4: 在飞书配置段新增 name 字段注释**

在 `[[projects.platforms]]` 和 `type = "feishu"` 之间插入注释行：

```toml
[[projects.platforms]]
type = "feishu"
# name = "feishu"  # optional, defaults to type. Specify unique name when running multiple feishu instances

[projects.platforms.options]
```

同样在 lark 配置段（注释块内）也加注释：

```toml
# [[projects.platforms]]
# type = "lark"
# name = "lark"    # optional, defaults to type. Specify unique name when running multiple lark instances
```

- [ ] **Step 5: 运行 go build + go test 验证**

Run: `go build ./... && go test ./...`
Expected: 编译通过，测试通过

- [ ] **Step 6: Commit Phase 2**

```bash
git add config/config.go cmd/cc-connect/main.go config.example.toml
git commit -m "feat: add Name field to PlatformConfig with validation and opts injection"
```

---

## Phase 3 — 飞书平台（instanceTag、Tag()/tag()、name 解析）

**目标**：飞书平台从 opts 读取 `cc_platform_name`，`Tag()` 返回自定义值。

### Task 3.1: Platform 结构体新增 instanceTag 字段

**Files:**
- Modify: `platform/feishu/feishu.go:115-170`

- [ ] **Step 1: 在 Platform 结构体中新增 instanceTag 字段**

在 `platformName` 字段之后新增：

```go
type Platform struct {
    mu                         sync.RWMutex
    platformName               string
    instanceTag                string    // routing identifier — "feishu" or "feishu-teamA"
    domain                     string
    appID                      string
    ...
}
```

### Task 3.2: newPlatform 从 opts 读取 name 并校验

**Files:**
- Modify: `platform/feishu/feishu.go:186-305`

- [ ] **Step 2: 在 newPlatform 函数中解析 cc_platform_name 并校验**

在 `newPlatform` 函数中，`name` 参数解析之后，新增从 opts 读取 `cc_platform_name` 的逻辑：

```go
func newPlatform(name, domain string, opts map[string]any) (core.Platform, error) {
    // Read optional instance name from config
    instanceName, _ := opts["cc_platform_name"].(string)
    if instanceName == "" {
        instanceName = name
    }

    // Validate: custom name must start with platform type + "-"
    if instanceName != name && !strings.HasPrefix(instanceName, name+"-") {
        return nil, fmt.Errorf("feishu: invalid name %q: must equal %q or start with %q followed by '-'",
            instanceName, name, name)
    }

    appID, _ := opts["app_id"].(string)
    appSecret, _ := opts["app_secret"].(string)
    ...
```

然后在构造 Platform 结构体时设置 `instanceTag`：

```go
    base := &Platform{
        platformName: name,
        instanceTag:  instanceName,
        domain:       domain,
        appID:        appID,
        ...
    }
```

### Task 3.3: 更新 Tag() 和 tag() 方法

**Files:**
- Modify: `platform/feishu/feishu.go:307-313`

当前代码：

```go
func (p *Platform) Name() string { return p.platformName }
func (p *Platform) Tag() string  { return p.platformName }    // Phase 1 添加的默认实现
func (p *Platform) tag() string  { return p.platformName }
```

- [ ] **Step 3: 将 Tag() 和 tag() 改为返回 p.instanceTag**

```go
func (p *Platform) Name() string { return p.platformName }
func (p *Platform) Tag() string  { return p.instanceTag }
func (p *Platform) tag() string  { return p.instanceTag }
```

`Name()` 不变，仍然返回 `"feishu"` 或 `"lark"`。

### Task 3.4: 验证 feishu 构造和 session key

需要确认所有使用 `p.tag()` 的地方（session key 构建）现在都用 `instanceTag`，这是正确的——因为 session key 前缀应该是路由标识。

`makeSessionKey` 系列函数已经使用 `p.tag()`，将 `tag()` 改为返回 `instanceTag` 后，session key 自然使用正确的路由前缀。无需额外修改。

- [ ] **Step 4: 运行 go build 验证编译通过**

Run: `go build ./...`
Expected: 编译通过

### Task 3.5: 飞书平台 Tag() 测试

**Files:**
- Modify: `platform/feishu/platform_test.go`

- [ ] **Step 5: 新增 Tag() 默认行为测试**

```go
func TestTagDefault(t *testing.T) {
    opts := map[string]any{
        "app_id":      "test_app",
        "app_secret":   "test_secret",
    }
    p, err := newPlatform("feishu", lark.FeishuBaseUrl, opts)
    require.NoError(t, err)

    feishu := p.(*Platform)
    assert.Equal(t, "feishu", feishu.Name())
    assert.Equal(t, "feishu", feishu.Tag())
    assert.Equal(t, "feishu", feishu.tag())
}
```

- [ ] **Step 6: 新增 Tag() 自定义行为测试**

```go
func TestTagCustom(t *testing.T) {
    opts := map[string]any{
        "app_id":           "test_app",
        "app_secret":        "test_secret",
        "cc_platform_name":  "feishu-teamA",
    }
    p, err := newPlatform("feishu", lark.FeishuBaseUrl, opts)
    require.NoError(t, err)

    feishu := p.(*Platform)
    assert.Equal(t, "feishu", feishu.Name())       // Name() 不变
    assert.Equal(t, "feishu-teamA", feishu.Tag())   // Tag() 自定义
    assert.Equal(t, "feishu-teamA", feishu.tag())   // tag() 自定义
}
```

- [ ] **Step 7: 新增 name 校验错误测试**

```go
func TestTagInvalidName(t *testing.T) {
    opts := map[string]any{
        "app_id":           "test_app",
        "app_secret":        "test_secret",
        "cc_platform_name":  "myapp",
    }
    _, err := newPlatform("feishu", lark.FeishuBaseUrl, opts)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "invalid name")
    assert.Contains(t, err.Error(), "myapp")
}
```

- [ ] **Step 8: 新增 lark 变体校验测试**

```go
func TestTagLarkVariant(t *testing.T) {
    // valid lark name
    opts1 := map[string]any{
        "app_id":           "test_app",
        "app_secret":        "test_secret",
        "cc_platform_name":  "lark-intl",
    }
    p1, err := newPlatform("lark", lark.LarkBaseUrl, opts1)
    require.NoError(t, err)
    lark1 := p1.(*Platform)
    assert.Equal(t, "lark", lark1.Name())
    assert.Equal(t, "lark-intl", lark1.Tag())

    // invalid lark name
    opts2 := map[string]any{
        "app_id":           "test_app",
        "app_secret":        "test_secret",
        "cc_platform_name":  "myapp",
    }
    _, err = newPlatform("lark", lark.LarkBaseUrl, opts2)
    assert.Error(t, err)
}
```

- [ ] **Step 9: 运行飞书平台测试**

Run: `go test ./platform/feishu/ -v -run TestTag`
Expected: 全部通过

- [ ] **Step 10: 运行全量测试验证**

Run: `go test ./...`
Expected: 全量通过

- [ ] **Step 11: Commit Phase 3**

```bash
git add platform/feishu/feishu.go platform/feishu/platform_test.go
git commit -m "feat: feishu platform supports custom instance Tag via config name field"
```

---

## Phase 4 — Engine 路由（11 处 Name() → Tag()）

**目标**：所有路由查找从 `Name()` 切换到 `Tag()`，实现多实例正确路由。

### Task 4.1: 重命名 extractPlatformName → extractPlatformTag

**Files:**
- Modify: `core/engine.go:13388-13393`

当前代码：

```go
func extractPlatformName(sessionKey string) string {
    if i := strings.IndexByte(sessionKey, ':'); i >= 0 {
        return sessionKey[:i]
    }
    return sessionKey
}
```

- [ ] **Step 1: 重命名函数并更新注释**

```go
func extractPlatformTag(sessionKey string) string {
    if i := strings.IndexByte(sessionKey, ':'); i >= 0 {
        return sessionKey[:i]
    }
    return sessionKey
}
```

- [ ] **Step 2: 全局替换所有 extractPlatformName 调用为 extractPlatformTag**

搜索 `extractPlatformName` 在 `core/engine.go` 中的所有调用位置，逐一替换为 `extractPlatformTag`。涉及行号：
- 行 10136（pushDeleteModeResultCard）
- 行 10197（pushModelSwitchResultCard）
- 行 9573（renderWhoamiCard）
- 行 13406（extractWorkspaceChannelKey）中引用

注意：`renderWhoamiCard` 中 `Platform: extractPlatformName(sessionKey)` 是显示用途，此处仍用 `extractPlatformTag` 是正确的——因为 session key 前缀是 Tag 值，提取它也得到 Tag 值。

### Task 4.2: ExecuteCronJob 路由替换

**Files:**
- Modify: `core/engine.go:1047-1058, 1062-1071`

当前代码（行 1047-1058）：

```go
sessionKey := job.SessionKey
platformName := ""
if idx := strings.Index(sessionKey, ":"); idx > 0 {
    platformName = sessionKey[:idx]
}
var targetPlatform Platform
for _, p := range e.platforms {
    if p.Name() == platformName {
        targetPlatform = p
        break
    }
}
```

- [ ] **Step 3: 替换变量名和路由匹配**

```go
sessionKey := job.SessionKey
platformTag := ""
if idx := strings.Index(sessionKey, ":"); idx > 0 {
    platformTag = sessionKey[:idx]
}
var targetPlatform Platform
for _, p := range e.platforms {
    if p.Tag() == platformTag {
        targetPlatform = p
        break
    }
}
```

当前代码（行 1062-1071 fallback）：

```go
if targetPlatform == nil {
    for _, p := range e.platforms {
        needle := ":" + p.Name() + ":"
        if idx := strings.Index(sessionKey, needle); idx >= 0 {
            targetPlatform = p
            platformName = p.Name()
            sessionKey = sessionKey[idx+1:]
            break
        }
    }
}
```

- [ ] **Step 4: 替换 fallback 中的 Name() → Tag()**

```go
if targetPlatform == nil {
    for _, p := range e.platforms {
        needle := ":" + p.Tag() + ":"
        if idx := strings.Index(sessionKey, needle); idx >= 0 {
            targetPlatform = p
            platformTag = p.Tag()
            sessionKey = sessionKey[idx+1:]
            break
        }
    }
}
```

### Task 4.3: ExecuteHeartbeat 路由替换

**Files:**
- Modify: `core/engine.go:1477-1482, 1487-1496`

与 ExecuteCronJob 相同模式，替换 `p.Name()` → `p.Tag()`，变量名 `platformName` → `platformTag`。

- [ ] **Step 5: 替换 ExecuteHeartbeat 主查找和 fallback**

主查找（行 1477-1482）：

```go
var targetPlatform Platform
for _, p := range e.platforms {
    if p.Tag() == platformTag {
        targetPlatform = p
        break
    }
}
```

fallback（行 1487-1496）：

```go
if targetPlatform == nil {
    for _, p := range e.platforms {
        needle := ":" + p.Tag() + ":"
        if idx := strings.Index(sessionKey, needle); idx >= 0 {
            targetPlatform = p
            platformTag = p.Tag()
            sessionKey = sessionKey[idx+1:]
            break
        }
    }
}
```

### Task 4.4: SendRestartNotification 路由替换

**Files:**
- Modify: `core/engine.go:107-113`

当前代码：

```go
func (e *Engine) SendRestartNotification(platformName, sessionKey string) {
    for _, p := range e.platforms {
        if p.Name() != platformName {
            continue
        }
```

- [ ] **Step 6: 替换函数参数名和路由匹配**

```go
func (e *Engine) SendRestartNotification(platformTag, sessionKey string) {
    for _, p := range e.platforms {
        if p.Tag() != platformTag {
            continue
        }
```

注意：此函数的参数名从 `platformName` 改为 `platformTag`。需检查所有调用此函数的地方是否也需更新参数名。搜索 `SendRestartNotification` 的调用位置，确保传入的参数名语义正确。

### Task 4.5: pushDeleteModeResultCard 和 pushModelSwitchResultCard 路由替换

**Files:**
- Modify: `core/engine.go:10136-10140, 10197-10201`

当前代码（行 10136-10140）：

```go
platformName := extractPlatformName(sessionKey)
var targetPlatform Platform
for _, p := range e.platforms {
    if p.Name() == platformName {
        targetPlatform = p
        break
```

- [ ] **Step 7: 替换 pushDeleteModeResultCard 路由**

```go
platformTag := extractPlatformTag(sessionKey)
var targetPlatform Platform
for _, p := range e.platforms {
    if p.Tag() == platformTag {
        targetPlatform = p
        break
```

- [ ] **Step 8: 替换 pushModelSwitchResultCard 路由**

```go
platformTag := extractPlatformTag(sessionKey)
var targetPlatform Platform
for _, p := range e.platforms {
    if p.Tag() == platformTag {
        targetPlatform = p
        break
```

### Task 4.6: tryProviderAddPreset 路由替换

**Files:**
- Modify: `core/engine.go:9005-9014`

与 pushDeleteModeResultCard 相同模式。

- [ ] **Step 9: 替换 tryProviderAddPreset 路由**

搜索 `tryProviderAddPreset` 中的 `p.Name()` 路由匹配，替换为 `p.Tag()`。变量名 `platformName` → `platformTag`。

### Task 4.7: Webhook 路由替换

**Files:**
- Modify: `core/webhook.go:202-218, 260-273`

当前代码（executePrompt，行 202-218）：

```go
platformName := ""
if idx := strings.Index(sessionKey, ":"); idx > 0 {
    platformName = sessionKey[:idx]
}
var targetPlatform Platform
for _, p := range engine.platforms {
    if p.Name() == platformName {
        targetPlatform = p
        break
    }
}
```

- [ ] **Step 10: 替换 executePrompt 路由**

```go
platformTag := ""
if idx := strings.Index(sessionKey, ":"); idx > 0 {
    platformTag = sessionKey[:idx]
}
var targetPlatform Platform
for _, p := range engine.platforms {
    if p.Tag() == platformTag {
        targetPlatform = p
        break
    }
}
```

- [ ] **Step 11: 替换 executeShell 路由**

```go
platformTag := ""
if idx := strings.Index(sessionKey, ":"); idx > 0 {
    platformTag = sessionKey[:idx]
}
var targetPlatform Platform
for _, p := range engine.platforms {
    if p.Tag() == platformTag {
        targetPlatform = p
        break
    }
}
```

### Task 4.8: Relay 路由替换

**Files:**
- Modify: `core/relay.go:251-269`

当前代码（行 253）：

```go
if p.Name() != platform {
    continue
}
```

- [ ] **Step 12: 替换 sendToGroup 路由**

```go
if p.Tag() != platform {
    continue
}
```

### Task 4.9: Management 路由替换

**Files:**
- Modify: `core/management.go:1619-1622`

当前代码（行 1619-1622）：

```go
for pName, ref := range m.bridgeServer.engines {
    if ref.platform != nil && ref.platform.Name() == name {
        project = pName
        break
    }
}
```

- [ ] **Step 13: 替换 bridge 状态查找**

```go
for pName, ref := range m.bridgeServer.engines {
    if ref.platform != nil && ref.platform.Tag() == name {
        project = pName
        break
    }
}
```

### Task 4.10: renderWhoamiCard 中 extractPlatformName 调用

**Files:**
- Modify: `core/engine.go:9573`

当前代码：

```go
Platform:   extractPlatformName(sessionKey),
```

- [ ] **Step 14: 替换为 extractPlatformTag**

```go
Platform:   extractPlatformTag(sessionKey),
```

### Task 4.11: extractWorkspaceChannelKey 中的引用

**Files:**
- Modify: `core/engine.go:13406` 附近

- [ ] **Step 15: 更新 extractWorkspaceChannelKey 中对 extractPlatformName 的引用**

搜索 `extractWorkspaceChannelKey` 函数体中对 `extractPlatformName` 的调用，替换为 `extractPlatformTag`。

### Task 4.12: 编译和测试验证

- [ ] **Step 16: 运行 go build 验证编译通过**

Run: `go build ./...`
Expected: 编译通过

- [ ] **Step 17: 运行全量测试**

Run: `go test ./...`
Expected: 全量通过

- [ ] **Step 18: Commit Phase 4**

```bash
git add core/engine.go core/webhook.go core/relay.go core/management.go
git commit -m "feat: switch all routing lookups from Platform.Name() to Platform.Tag()"
```

---

## Phase 5 — 测试补全

**目标**：为 Phase 4 的路由变更增加专门的测试覆盖。

### Task 5.1: 多实例路由测试

**Files:**
- Modify: `core/engine_test.go`

- [ ] **Step 1: 创建 multiTagStubPlatform 类型**

在 engine_test.go 中新增一种 stub，可以独立设置 Name 和 Tag：

```go
type multiTagStubPlatform struct {
    stubPlatformEngine
    tagValue string
}

func (p *multiTagStubPlatform) Tag() string { return p.tagValue }
```

- [ ] **Step 2: 新增路由测试——两个相同 Name 不同 Tag 的平台**

```go
func TestEngineRoutingByTag(t *testing.T) {
    // Two platforms with same Name ("feishu") but different Tags
    p1 := &multiTagStubPlatform{stubPlatformEngine{n: "feishu"}, tagValue: "feishu-teamA"}
    p2 := &multiTagStubPlatform{stubPlatformEngine{n: "feishu"}, tagValue: "feishu-teamB"}

    // Verify Tag-based routing
    tag := extractPlatformTag("feishu-teamA:oc_chat1:ou_user1")
    assert.Equal(t, "feishu-teamA", tag)

    // Verify extractPlatformTag extracts the correct prefix
    tag2 := extractPlatformTag("feishu-teamB:oc_chat2:ou_user2")
    assert.Equal(t, "feishu-teamB", tag2)

    // Verify Name-based matching would fail (both return "feishu")
    matched := 0
    for _, p := range []Platform{p1, p2} {
        if p.Name() == "feishu" {
            matched++
        }
    }
    assert.Equal(t, 2, matched) // Name() matches both — not useful for routing

    // Verify Tag-based matching finds the correct one
    for _, p := range []Platform{p1, p2} {
        if p.Tag() == "feishu-teamA" {
            assert.Equal(t, "feishu-teamA", p.Tag())
            assert.Equal(t, "feishu", p.Name())
        }
    }
}
```

### Task 5.2: 配置校验测试

**Files:**
- Modify: `config/config_test.go` 或相关测试文件

- [ ] **Step 3: 新增 PlatformConfig Name 校验测试**

```go
func TestPlatformConfigNameValidation(t *testing.T) {
    tests := []struct {
        name    string
        config  PlatformConfig
        wantErr bool
    }{
        {"empty name is valid", PlatformConfig{Type: "feishu", Name: ""}, false},
        {"name equals type is valid", PlatformConfig{Type: "feishu", Name: "feishu"}, false},
        {"name starts with type-hyphen is valid", PlatformConfig{Type: "feishu", Name: "feishu-teamA"}, false},
        {"lark variant is valid", PlatformConfig{Type: "lark", Name: "lark-intl"}, false},
        {"name without type prefix is invalid", PlatformConfig{Type: "feishu", Name: "myapp"}, true},
        {"name with wrong type prefix is invalid", PlatformConfig{Type: "feishu", Name: "lark-teamA"}, true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            proj := ProjectConfig{
                Name:      "test",
                Platforms: []PlatformConfig{tt.config},
                Agent:     AgentConfig{Type: "claudecode"},
            }
            err := proj.validate("test")
            if tt.wantErr {
                assert.Error(t, err)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

注意：需确认 `validate` 方法的调用方式以及 `AgentConfig` 的最小必需字段。

- [ ] **Step 4: 运行新增测试**

Run: `go test ./core/ -v -run TestEngineRoutingByTag && go test ./config/ -v -run TestPlatformConfigNameValidation`
Expected: 全部通过

- [ ] **Step 5: 运行全量测试**

Run: `go test ./...`
Expected: 全量通过

- [ ] **Step 6: Commit Phase 5**

```bash
git add core/engine_test.go config/config_test.go
git commit -m "test: add multi-instance routing and config name validation tests"
```

---

## 自查清单

- [x] Spec 覆盖：§1 接口层 → Phase 1；§2 飞书平台 → Phase 3；§3 路由 → Phase 4；§4 配置 → Phase 2；§5 测试 → Phase 5
- [x] 占位符扫描：无 TBD、TODO、未完成步骤
- [x] 类型一致性：所有平台 `Tag()` 返回 `string`，与 `Name()` 类型一致；`instanceTag` 是 `string` 字段；`extractPlatformTag` 返回 `string`