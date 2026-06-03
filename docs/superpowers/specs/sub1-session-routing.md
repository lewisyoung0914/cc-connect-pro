# 子项目 1：Session 路由 — makeSessionKey 条件扩展

**日期**: 2026-06-02
**状态**: 待实施

## 目标

扩展 `makeSessionKey` 的条件判断，使 `thread_isolation` 在私聊（P2P）中生效。这是整个功能的**核心路由变更**，决定消息如何分发到不同的 agent session。

## 变更

**文件**: `platform/feishu/feishu.go`

**当前代码**（约第 2933 行）：

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

**变更代码**：

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

**唯一变更点**：`stringValue(msg.ChatType) == "group"` → `chatType == "group" || chatType == "p2p"`

## 行为变化

| 场景 | `thread_isolation=false`（默认） | `thread_isolation=true`（变更后） |
|------|----------------------------------|----------------------------------|
| 私聊新消息（无 root_id） | `feishu:ou_xxx:ou_xxx` | `feishu:ou_xxx:ou_xxx:root:om_msg123` |
| 私聊回复消息（有 root_id） | `feishu:ou_xxx:ou_xxx` | `feishu:ou_xxx:ou_xxx:root:om_parent` |
| 群聊消息 | 不变 | 不变 |

## 配置优先级

当 `threadIsolation` 和 `shareSessionInChannel` 同时为 true 时，`threadIsolation` 在私聊中优先（因为它在 `makeSessionKey` 中先检查）。这与群聊行为一致。

## 向后兼容

- `thread_isolation = false`：私聊行为完全不变
- `thread_isolation = true`：群聊行为完全不变（已有实现），私聊新增线程并发

## 验证方法

手动测试：配置飞书 bot 的 `thread_isolation = true`，在私聊中：
1. 发送新消息 → 应创建新线程 session
2. 回复 bot 的消息 → 应路由到同一线程 session
3. 发送第二条新消息 → 应创建第二个线程 session，与第一个并发运行