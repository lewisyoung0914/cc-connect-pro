# 飞书私聊线程隔离 — 私聊并发多任务

**日期**: 2026-06-02
**状态**: 草稿
**作者**: brainstorming skill

## 问题

当用户配置单个飞书 bot 并启动 `cc-connect` 后，私聊（P2P）会话被一个 agent 对话独占。同一私聊中的所有消息共享一个 `SessionKey`（`feishu:{chatID}:{userID}`），通过 `Session.TryLock()` 串行处理。用户在第一个任务运行期间无法启动第二个任务 — 第二条消息排队等待当前 turn 完成。

这限制了生产力：用户处理项目 A 时必须等 agent 完成后才能让它处理项目 B。

## 目标

将已有的 `thread_isolation` 功能从群聊扩展到私聊，实现私聊中的**并发多任务**。飞书中的每个"回复链"（通过 `root_id` / 回复功能串联的消息）成为独立的 agent session。多个回复链并发运行，各有自己的 Claude Code 子进程和 `TryLock`。

此外，项目已支持多飞书应用实例（多个 `[[projects]]` 各配置不同的 `app_id`/`app_secret`），每个项目有独立的 Engine 和 session 管理。此部分无需代码改动，仅需文档说明。

## 方案

将 `thread_isolation` 扩展到私聊，复用群聊已有的 session key 派生模式：

- **新顶级消息**（无 `root_id`）→ 创建新线程 session，key = `feishu:{chatID}:{userID}:root:{messageID}`
- **回复消息**（有 `root_id`）→ 路由到已有线程 session，key = `feishu:{chatID}:{userID}:root:{rootID}`
- **多个线程** → 各有独立的 `interactiveState` 和 `TryLock`，实现真正并发

此方案代码改动极小（`makeSessionKey` 中一个条件扩展），复用飞书原生回复 UI，完全向后兼容。

## 设计细节

### 1. Session Key 派生（`makeSessionKey`）

**当前代码**（[feishu.go:2933](platform/feishu/feishu.go#L2933)）：

```go
func (p *Platform) makeSessionKey(msg *larkim.EventMessage, chatID, userID string) string {
    if p.threadIsolation && msg != nil && stringValue(msg.ChatType) == "group" {
        rootID := stringValue(msg.RootId)
        if rootID == "" {
            rootID = stringValue(msg.MessageId)
        }
        if rootID != "" {
            return fmt.Sprintf("%s:%s:root:%s", p.tag(), chatID, rootID)
        }
    }
    if p.shareSessionInChannel {
        return fmt.Sprintf("%s:%s", p.tag(), chatID)
    }
    return fmt.Sprintf("%s:%s:%s", p.tag(), chatID, userID)
}
```

**变更代码** — 将条件扩展到私聊：

```go
func (p *Platform) makeSessionKey(msg *larkim.EventMessage, chatID, userID string) string {
    if p.threadIsolation && msg != nil {
        chatType := stringValue(msg.ChatType)
        if chatType == "group" || chatType == "p2p" {
            rootID := stringValue(msg.RootId)
            if rootID == "" {
                rootID = stringValue(msg.MessageId)
            }
            if rootID != "" {
                return fmt.Sprintf("%s:%s:root:%s", p.tag(), chatID, rootID)
            }
        }
    }
    if p.shareSessionInChannel {
        return fmt.Sprintf("%s:%s", p.tag(), chatID)
    }
    return fmt.Sprintf("%s:%s:%s", p.tag(), chatID, userID)
}
```

唯一变更：条件 `stringValue(msg.ChatType) == "group"` 改为 `chatType == "group" || chatType == "p2p"`。其余所有逻辑（root ID 派生、session key 格式、线程匹配）完全复用群聊已有实现。

### 2. Bot 回复行为（`shouldReplyInThread`）

**当前代码**（[feishu.go:2961](platform/feishu/feishu.go#L2961)）：

```go
func (p *Platform) shouldReplyInThread(rc replyContext) bool {
    if rc.messageID == "" {
        return false
    }
    return p.threadIsolation && isThreadSessionKey(rc.sessionKey)
}
```

`ReplyInThread=true` 是飞书**群聊专用**的 API 参数（创建"话题"线程）。私聊中 bot 应仍回复用户消息（使用 `Im.Message.Reply` API），但不应设置 `ReplyInThread=true`。普通回复在私聊中形成可视回复链，无需话题线程 UI。

**变更代码**：

`replyContext` 结构体新增 `chatType` 字段：

```go
type replyContext struct {
    messageID   string
    chatID      string
    sessionKey  string
    chatType    string  // "group" 或 "p2p"
}
```

最终 `shouldReplyInThread`：

```go
func (p *Platform) shouldReplyInThread(rc replyContext) bool {
    if rc.messageID == "" || !p.threadIsolation || !isThreadSessionKey(rc.sessionKey) {
        return false
    }
    // ReplyInThread 仅适用于群聊线程。
    // 私聊线程隔离中使用普通回复 API 形成可视回复链。
    return rc.chatType == "group"
}
```

私聊线程中 bot 仍回复用户消息（使用 `replyMessage`），形成可视回复链。`shouldUseThreadOrReplyAPI` 方法在 `noReplyToTrigger=false`（默认值）时已确保使用 reply API。

### 3. 交互卡片 Session 路由（`sessionKeyFromCardAction`）

交互卡片按钮点击需要路由到正确的线程 session。当前实现从卡片 payload 中提取 `session_key`，回退到 per-user key。

**线程 session 无需改动** — 卡片 payload 已携带渲染时嵌入的 `session_key` 值。当 bot 在私聊线程中渲染卡片时，嵌入线程的 `session_key`。用户点击该卡片按钮时，`sessionKeyFromCardAction` 返回嵌入的线程 key，正确路由。

唯一需确保：私聊线程 session 中的卡片正确嵌入线程 session key。卡片渲染管道已处理 — `replyContext.sessionKey` 会传播到 `renderCard`。

### 4. 并发模型

当 `thread_isolation` 在私聊中生效时，并发模型如下：

```
Engine.interactiveStates map:
├── "feishu:ou_userA:ou_userA:root:om_msg1" → interactiveState { AgentSession #1, TryLock }
├── "feishu:ou_userA:ou_userA:root:om_msg2" → interactiveState { AgentSession #2, TryLock }
└── "feishu:ou_userA:ou_userA:root:om_msg3" → interactiveState { AgentSession #3, TryLock }
```

每个 `interactiveKey` 有独立的 `TryLock`。多个线程可同时处理消息 — 各有自己的 Claude Code 子进程（`AgentSession`）。Engine 的 `handleMessage` 根据 `SessionKey` 路由到对应的 `interactiveState`。

单线程内的串行行为保持不变：消息排队等待当前 turn 完成。这是正确且预期的 — 单个 agent session 每次只处理一个 prompt。

### 5. 边界情况

**回复非 bot 消息**：用户可能回复自己之前发的消息（而非 bot 消息）。`root_id` 字段始终指向线程根消息。飞书回复链共享相同的 `root_id`，路由正确，无需特殊处理。

**连续多条新消息**：每条新顶级消息（无 `root_id`）创建新线程 session。这是预期行为 — 每条新消息启动一个新任务。要继续已有任务，用户需回复 bot 的响应。

**线程 session 空闲超时**：线程空闲超过 `reset_on_idle_mins` 后，Engine 回收 `AgentSession`。下次用户回复该线程时，创建新 `AgentSession`（若保存了旧 session ID，可使用 `--continue` 恢复上下文）。

**并发线程上限**：每个 agent session 是一个 Claude Code 子进程，消耗内存和 CPU。后续可增加 `max_concurrent_threads` 配置（默认 5），超出限制时拒绝新线程。初始版本不加硬限制 — 自然资源上限（进程内存、API 速率限制）提供实际边界。

**向后兼容**：`thread_isolation = false`（默认）时私聊行为不变 — 所有消息使用 `feishu:{chatID}:{userID}`，按用户串行。`thread_isolation = true` 时群聊行为不变（已有实现），私聊启用线程并发。

**配置交互**：`shareSessionInChannel` 和 `threadIsolation` 在私聊中存在优先级冲突。若两者均为 true，`shareSessionInChannel` 会将所有消息折叠为 `feishu:{chatID}`（私聊中每聊一个 key），破坏线程隔离。解决方案：`threadIsolation` 在私聊中优先于 `shareSessionInChannel`。这已是 `makeSessionKey` 的自然顺序 — 线程隔离分支先检查，`shareSessionInChannel` 后检查（作为回退）。扩展条件仅让线程隔离分支在私聊中生效，自然覆盖 `shareSessionInChannel`。

### 6. 配置

无需新增配置项。现有 `thread_isolation` 选项现在同时适用于群聊和私聊：

```toml
[[projects.platforms]]
type = "feishu"
[projects.platforms.options]
app_id = "cli_xxx"
app_secret = "xxx"
thread_isolation = true   # 现在也启用私聊线程并发
```

### 7. i18n

初始版本无需新增 i18n key。线程内的排队/繁忙行为与当前 per-user 串行相同 — 同样的消息、同样的用户体验。

可选优化：后续版本可将 `MsgPreviousProcessing` 的文案从"前一个请求仍在处理"改为"线程仍在处理中"，更准确反映线程场景。但初始版本直接复用现有 key。

### 8. 多飞书应用实例

当前架构已完全支持多飞书应用实例。每个 `[[projects]]` 条目在其 `[[projects.platforms]]` 中配置自己的 `app_id`/`app_secret`。多个项目各有独立的 Engine、Agent、SessionManager，在同一 Go 进程中并发运行。WebSocket 共享机制（[ws_shared.go](platform/feishu/ws_shared.go)）确保多个项目共享同一 `app_id` 时事件正确路由。

此部分无需代码改动。文档改进可更突出地展示多应用配置模式。

## 测试计划

1. **`makeSessionKey` 单元测试**：新增私聊线程隔离测试 — 新消息（无 root_id）、回复消息（有 root_id）、混合私聊/群聊场景。

2. **`shouldReplyInThread` 单元测试**：验证私聊线程 session 不设置 `ReplyInThread=true`，群聊线程 session 设置。

3. **集成测试**：模拟私聊事件（有无 `root_id`），验证 session 路由为不同线程创建独立 session、回复路由到正确线程。

4. **并发测试**：启动两个并发私聊线程 session，验证它们独立处理、互不阻塞。

## 代码变更汇总

| 文件 | 变更 |
|------|------|
| `platform/feishu/feishu.go` | `makeSessionKey`：条件从 `"group"` 扩展为 `"group" || "p2p"` |
| `platform/feishu/feishu.go` | `shouldReplyInThread`：检查 `chatType` 排除私聊的 `ReplyInThread=true` |
| `platform/feishu/feishu.go` | `replyContext`：新增 `chatType` 字段 |
| `platform/feishu/feishu.go` | 在 `replyContext` 构造处传播 `chatType` |
| `platform/feishu/feishu_test.go` | 新增私聊线程隔离测试用例 |