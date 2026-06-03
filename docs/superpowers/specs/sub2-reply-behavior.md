# 子项目 2：回复行为 — replyContext + shouldReplyInThread

**日期**: 2026-06-02
**状态**: 待实施
**依赖**: 子项目 1（makeSessionKey 路由变更）

## 目标

确保私聊线程中 bot 的回复行为正确：使用普通 reply API 形成可视回复链，而非群聊的 `ReplyInThread=true` 话题模式。

## 变更

### 2a. replyContext 结构体新增 chatType 字段

**文件**: `platform/feishu/feishu.go`

```go
type replyContext struct {
    messageID   string
    chatID      string
    sessionKey  string
    chatType    string  // 新增："group" 或 "p2p"
}
```

需在所有构造 `replyContext` 的位置传播 `chatType`。主要来源是 `onMessage` 处理流程中已知的 `msg.ChatType`。

### 2b. shouldReplyInThread 方法修改

**当前代码**（约第 2961 行）：

```go
func (p *Platform) shouldReplyInThread(rc replyContext) bool {
    if rc.messageID == "" {
        return false
    }
    return p.threadIsolation && isThreadSessionKey(rc.sessionKey)
}
```

**变更代码**：

```go
func (p *Platform) shouldReplyInThread(rc replyContext) bool {
    if rc.messageID == "" || !p.threadIsolation || !isThreadSessionKey(rc.sessionKey) {
        return false
    }
    // ReplyInThread 仅适用于群聊线程。
    // 私聊线程中使用普通回复 API 形成可视回复链。
    return rc.chatType == "group"
}
```

## 行为变化

| 场景 | 变更前 | 变更后 |
|------|--------|--------|
| 群聊线程回复 | `ReplyInThread=true` | `ReplyInThread=true`（不变） |
| 私聊线程回复 | `ReplyInThread=true`（错误） | 普通 reply（正确） |
| 私聊非线程回复 | 普通 reply | 普通 reply（不变） |

## 需排查的位置

需要确认所有构造 `replyContext` 的代码路径，确保 `chatType` 正确传入。主要位置：

1. `onMessage` → `replyContext` 构造
2. `onCardAction` → `replyContext` 构造
3. 其他回复场景（reaction 回复、permission 回复等）

## 验证方法

手动测试：配置 `thread_isolation = true`，在私聊线程中 bot 回复不应创建话题线程，而是普通回复链。