# 多飞书实例路由改造设计

**日期**: 2026-06-02
**状态**: 待审阅
**范围**: 仅飞书平台；架构上为其他平台留有扩展空间

## 问题

当一个项目在 `[[projects.platforms]]` 中配置多个飞书应用时，Engine 无法区分它们用于主动消息（cron、heartbeat、webhook、send）。所有路由使用 `Platform.Name()`，它对每个实例都返回 `"feishu"`，导致 first-match 始终命中第一个实例。Session key 也共享 `"feishu:"` 前缀，无法识别会话属于哪个实例。

受影响的路由路径（11 处）：

| 路径 | 文件 | 函数 | 行号 |
|------|------|------|------|
| Cron | core/engine.go | ExecuteCronJob | 1054, 1064 |
| Heartbeat | core/engine.go | ExecuteHeartbeat | 1479, 1489 |
| Provider 切换 | core/engine.go | tryProviderAddPreset | 9005, 9014 |
| cc-connect send | core/engine.go | SendToSessionWithAttachments | 8997-9021 |
| 删除模式卡片 | core/engine.go | pushDeleteModeResultCard | 10139 |
| 模型切换卡片 | core/engine.go | pushModelSwitchResultCard | 10200 |
| 重启通知 | core/engine.go | SendRestartNotification | 109 |
| Webhook prompt | core/webhook.go | executePrompt | 210 |
| Webhook shell | core/webhook.go | executeShell | 269 |
| Relay 群消息 | core/relay.go | sendToGroup | 253 |
| Bridge 状态 | core/management.go | bridge status | 1620 |

注：`initPlatformCapabilities`（行 1689）中的 `p.Name()` 仅用于类型判断和日志（如 `strings.EqualFold(p.Name(), "telegram")`），不是实例路由——保留使用 `Name()`。

## 设计决策

- **范围**：仅飞书平台，但 `Tag()` 接口机制足够通用，未来其他平台可复用
- **向后兼容**：完全向后兼容。未配置 `name` 时行为不变
- **实例标识**：在 `Platform` 接口新增 `Tag() string` 方法；`Name()` 仍为类型名
- **WebSocket 共享**：不变——`sharedWSGroup` 继续按 `app_id + domain` 分组
- **配置字段**：`name` 是 `PlatformConfig` 的顶层字段，不在 `options` 内

## §1 — 接口层

在 `core/interfaces.go` 的 `Platform` 接口新增 `Tag() string`：

```go
type Platform interface {
    Name() string   // 平台类型名 — "feishu"、"telegram" — 用于显示、日志、类型判断
    Tag() string    // 路由标识符 — "feishu" 或 "feishu-teamA" — 用于 session key 前缀和平台查找
    Start(handler MessageHandler) error
    Reply(ctx context.Context, replyCtx any, content string) error
    Send(ctx context.Context, replyCtx any, content string) error
    Stop() error
}
```

所有现有平台实现添加默认 `Tag()`，返回与 `Name()` 相同的值：

```go
// platform/telegram/telegram.go
func (p *Platform) Tag() string { return p.Name() }
```

确保现有部署零影响——`Tag()` 默认等于 `Name()`。

**保持使用 `Name()` 的场景**（类型级操作）：
- `strings.EqualFold(p.Name(), "telegram")` — 类型判断
- `menuCommandsForPlatform(p.Name())` — 命令菜单过滤
- `displayCommandForPlatform(p.Name(), ...)` — 命令显示格式
- 日志、Hook 发射、Doctor 输出 — 显示类型名 `"feishu"` 而非 `"feishu-teamA"`

## §2 — 飞书平台变更

### 新增结构体字段

```go
type Platform struct {
    platformName string   // 不变 — Name() 返回此值（"feishu" 或 "lark"）
    instanceTag  string   // 新增 — Tag() 返回此值（"feishu" 或 "feishu-teamA"）
    ...
}
```

### newPlatform 变更

```go
func newPlatform(name string, baseURL string, opts map[string]any) (*Platform, error) {
    instanceName, _ := opts["cc_platform_name"].(string)
    if instanceName == "" {
        instanceName = name  // 默认值 = 工厂注册名（"feishu" 或 "lark"）
    }

    // 校验：自定义 name 必须以平台类型名开头加 "-"
    if instanceName != name && !strings.HasPrefix(instanceName, name+"-") {
        return nil, fmt.Errorf("feishu: invalid name %q: must start with %q followed by '-'",
            instanceName, name)
    }

    p := &Platform{
        platformName: name,
        instanceTag:  instanceName,
        ...
    }
}
```

### 方法变更

| 方法 | 改造前 | 改造后 |
|------|--------|--------|
| `Name()` | `p.platformName` | **不变** |
| `Tag()` | 无 | `p.instanceTag` |
| `tag()`（内部） | `p.platformName` | `p.instanceTag` |

### Session key 格式

- 未配置 `name`：`feishu:oc_xxx:ou_xxx` — 与当前完全一致
- 配置 `name = "feishu-teamA"`：`feishu-teamA:oc_xxx:ou_xxx` — 前缀改变

### WebSocket 共享 — 不变

`ws_shared.go` 中的 `sharedWSKey` 按 `app_id + domain` 分组，与 `Tag()` 无关。

### 配置示例

```toml
# 第一个飞书应用
[[projects.platforms]]
type = "feishu"
name = "feishu-teamA"    # 可选，默认为 "feishu"

[projects.platforms.options]
app_id = "cli_appA"
app_secret = "secretA"

# 第二个飞书应用
[[projects.platforms]]
type = "feishu"
name = "feishu-teamB"    # 项目内必须唯一

[projects.platforms.options]
app_id = "cli_appB"
app_secret = "secretB"
```

### name 校验规则

`name` 字段必须等于 `type`（默认情况）或以 `type + "-"` 开头。这确保 `Tag()` 值始终可被类型判断逻辑识别。无效示例：`"myapp"`、`"teamA-feishu"`。

## §3 — Engine 路由层变更

### extractPlatformName → extractPlatformTag

重命名辅助函数，逻辑不变：

```go
func extractPlatformTag(sessionKey string) string {
    if i := strings.IndexByte(sessionKey, ':'); i >= 0 {
        return sessionKey[:i]
    }
    return sessionKey
}
```

### 所有路由查找：Name() → Tag()

11 处路由位置将 `p.Name()` 替换为 `p.Tag()`。变量命名同步更新：路由上下文中的 `platformName` → `platformTag`。

| 路径 | 改造前 | 改造后 |
|------|--------|--------|
| Cron | `p.Name() == platformName` | `p.Tag() == platformTag` |
| Heartbeat | `p.Name() == platformName` | `p.Tag() == platformTag` |
| Provider 切换 | `p.Name() == platformName` | `p.Tag() == platformTag` |
| cc-connect send | `p.Name() == platformName` | `p.Tag() == platformTag` |
| 删除模式卡片 | `p.Name() == platformName` | `p.Tag() == platformTag` |
| 模型切换卡片 | `p.Name() == platformName` | `p.Tag() == platformTag` |
| 重启通知 | `p.Name() != platformName` | `p.Tag() != platformTag` |
| Webhook prompt | `p.Name() == platformName` | `p.Tag() == platformTag` |
| Webhook shell | `p.Name() == platformName` | `p.Tag() == platformTag` |
| Relay 群消息 | `p.Name() != platform` | `p.Tag() != platform` |
| Bridge 状态 | `p.Name() == name` | `p.Tag() == name` |
| Needle fallback | `":" + p.Name() + ":"` | `":" + p.Tag() + ":"` |

### 保持使用 Name() 的部分

以下场景保留 `Name()`，因为它们判断**类型**而非**实例**：
- `strings.EqualFold(p.Name(), "telegram")` — 类型判断
- `menuCommandsForPlatform(p.Name())` — 命令菜单过滤
- `displayCommandForPlatform(p.Name(), ...)` — 命令显示
- 日志、Hook、Doctor — 显示类型名

## §4 — 配置层与注册层变更

### PlatformConfig 结构体

```go
type PlatformConfig struct {
    Type    string         `toml:"type"`
    Name    string         `toml:"name"`    // 可选，实例标识；默认等于 Type
    Options map[string]any `toml:"options"`
}
```

### validate() — name 规则

- 空 `name` 合法（运行时默认等于 `type`）
- 若提供 `name`，必须等于 `type` 或以 `type + "-"` 开头
- 不符合规则的 `name` 产生清晰的错误提示

### opts 注入

在 `cmd/cc-connect/main.go` 的平台创建循环中注入 `name`：

```go
opts["cc_platform_name"] = pc.Name
```

### config.example.toml

```toml
[[projects.platforms]]
type = "feishu"
# name = "feishu"  # 可选，默认等于 type。多飞书实例时需唯一指定

[projects.platforms.options]
app_id = "cli_xxx"
app_secret = "xxx"
```

### feishu setup CLI

现有的 `--platform-index` 参数已支持多飞书配置，无需改动。但当存在多个飞书平台时，生成的配置应包含 `name` 字段。

## §5 — 测试策略

### 飞书平台测试（platform/feishu/platform_test.go）

1. **Tag() 默认行为**：未配置 `name` → `Tag()` 返回 `"feishu"` = `Name()`
2. **Tag() 自定义行为**：`name = "feishu-teamA"` → `Tag()` 返回 `"feishu-teamA"`，`Name()` 返回 `"feishu"`
3. **name 校验**：`name = "myapp"` → `newPlatform` 返回错误
4. **lark 变体**：`name = "lark-intl"` 合法，`name = "myapp"` 报错
5. **Session key 格式**：自定义 name → 前缀 `"feishu-teamA:"`

### Engine 测试（core/engine_test.go）

1. **stubPlatform 增加 Tag()**：返回与 `Name()` 相同值，现有测试不变
2. **多实例路由测试**：构造两个 `Tag()` 不同但 `Name()` 相同的 stub platform，验证 cron/heartbeat/webhook 路由到正确实例
3. **向后兼容测试**：`Tag() == Name()` 时，所有路由行为与改造前一致

### 配置测试

1. **空 name 合法**：不提供 `name` → 校验通过
2. **name 格式校验**：`"feishu-teamA"` 通过，`"myapp"` 报错，`"feishu"` 通过（等于 type）

### 端到端验证

- `go test ./...` — 全量通过
- `go build ./...` — 编译无错误
- 手动测试：配置两个飞书应用，验证消息路由到正确实例

## 实现阶段

### Phase 1 — 接口层 + 所有平台 Tag() 默认实现

**目标**：让 `Tag()` 成为 `Platform` 接口的正式方法，所有平台都能编译通过。

**变更文件**：
- `core/interfaces.go` — 在 `Platform` 接口新增 `Tag() string`
- 所有 `platform/*.go` — 添加默认 `func (p *Platform) Tag() string { return p.Name() }`
- `core/engine_test.go` — stubPlatform 增加 `Tag()` 实现
- `core/cron.go` — mutePlatform 增加 `Tag()` 委托
- `core/bridge.go` — bridge adapter 如有 `Name()` 委托则同步增加 `Tag()`

**验证**：`go build ./...` 编译通过，`go test ./...` 全量通过。此时 `Tag()` == `Name()`，行为完全不变。

### Phase 2 — 配置层（PlatformConfig.Name、校验、opts 注入）

**目标**：配置结构体支持 `name` 字段，并在创建平台时注入 opts。

**变更文件**：
- `config/config.go` — `PlatformConfig` 新增 `Name string` 字段；`validate()` 增加 name 格式校验
- `cmd/cc-connect/main.go` — 平台创建循环中注入 `opts["cc_platform_name"] = pc.Name`
- `config.example.toml` — 新增 `name` 字段注释

**验证**：`go build ./...` 编译通过。手动测试：配置中加 `name = "feishu-teamA"` 不报错；`name = "myapp"` 报错。

### Phase 3 — 飞书平台（instanceTag、Tag()/tag()、name 解析）

**目标**：飞书平台从 opts 读取 `cc_platform_name`，`Tag()` 返回自定义值。

**变更文件**：
- `platform/feishu/feishu.go` — 新增 `instanceTag` 字段；`Tag()` 返回 `p.instanceTag`；`tag()` 返回 `p.instanceTag`；`newPlatform` 从 opts 读取并校验 name
- `platform/feishu/platform_test.go` — 新增 Tag 默认/自定义/校验/lark 变体/session key 测试

**验证**：`go test ./platform/feishu/ -v` 通过。此时飞书 `Tag()` 可返回 `"feishu-teamA"`，但 Engine 路由仍用 `Name()`（下一阶段才改）。

### Phase 4 — Engine 路由（11 处 Name() → Tag()）

**目标**：所有路由查找从 `Name()` 切换到 `Tag()`，实现多实例正确路由。

**变更文件**：
- `core/engine.go` — 重命名 `extractPlatformName` → `extractPlatformTag`；11 处路由 `p.Name()` → `p.Tag()`；变量名 `platformName` → `platformTag`（仅路由上下文）
- `core/webhook.go` — 路由查找 `Name()` → `Tag()`
- `core/relay.go` — 路由查找 `Name()` → `Tag()`
- `core/management.go` — 路由查找 `Name()` → `Tag()`

**验证**：`go build ./...` 编译通过，`go test ./...` 全量通过。配置两个不同 `name` 的飞书平台，验证 cron/heartbeat/webhook 消息路由到正确实例。

### Phase 5 — 测试补全

**目标**：为 Phase 4 的路由变更增加专门的测试覆盖。

**变更文件**：
- `core/engine_test.go` — 新增多实例路由测试（两个 Tag 不同但 Name 相同的 stub platform）
- `config/config_test.go`（或现有测试文件）— 新增 name 校验测试

**验证**：`go test ./...` 全量通过。

## 变更文件汇总

| 文件 | 变更 |
|------|------|
| core/interfaces.go | 在 `Platform` 接口新增 `Tag() string` |
| core/engine.go | 重命名 `extractPlatformName` → `extractPlatformTag`；11 处 `Name()` → `Tag()`；`Name()` 用于类型判断/显示 |
| core/engine_test.go | stubPlatform 增加 `Tag()`；新增多实例路由测试 |
| core/webhook.go | 路由查找 `Name()` → `Tag()` |
| core/relay.go | 路由查找 `Name()` → `Tag()` |
| core/management.go | 路由查找 `Name()` → `Tag()` |
| core/bridge.go | 更新适配器 key 逻辑（如有需要） |
| core/cron.go（mutePlatform） | 增加 `Tag()` 委托 |
| config/config.go | `PlatformConfig` 新增 `Name` 字段；新增校验 |
| cmd/cc-connect/main.go | 将 `cc_platform_name` 注入 opts |
| platform/feishu/feishu.go | 新增 `instanceTag` 字段；更新 `Tag()`、`tag()`、`newPlatform` |
| platform/feishu/platform_test.go | 新增 Tag/默认/自定义/校验测试 |
| 所有其他 platform/*.go | 新增默认 `Tag() string { return p.Name() }` |
| config.example.toml | 新增 `name` 字段注释 |