# 子项目 3：单元测试 — 私聊线程隔离测试覆盖

**日期**: 2026-06-02
**状态**: 待实施
**依赖**: 子项目 1、2

## 目标

为私聊线程隔离功能添加单元测试，覆盖 session key 派生和回复行为的核心逻辑。

## 测试用例

### 3a. makeSessionKey 测试

**文件**: `platform/feishu/feishu_test.go`

| 测试场景 | 输入 | 期望 SessionKey |
|----------|------|-----------------|
| 私聊 + thread_isolation=true + 新消息（无 root_id） | chatType="p2p", rootId="", messageId="om_msg1" | `feishu:ou_chat:ou_chat:root:om_msg1` |
| 私聊 + thread_isolation=true + 回复消息 | chatType="p2p", rootId="om_parent", messageId="om_reply" | `feishu:ou_chat:ou_chat:root:om_parent` |
| 私聊 + thread_isolation=false | chatType="p2p" | `feishu:ou_chat:ou_user`（原有行为） |
| 群聊 + thread_isolation=true | chatType="group" | 原有行为不变 |
| 群聊 + thread_isolation=false | chatType="group" | 原有行为不变 |

### 3b. shouldReplyInThread 测试

| 测试场景 | 输入 | 期望结果 |
|----------|------|----------|
| 群聊线程 session + threadIsolation=true | chatType="group", thread session key | `true` |
| 私聊线程 session + threadIsolation=true | chatType="p2p", thread session key | `false` |
| 私聊非线程 session | chatType="p2p", 非线程 key | `false` |
| 群聊非线程 session | chatType="group", 非线程 key | `false` |

### 3c. shareSessionInChannel 优先级测试

| 测试场景 | 配置 | 期望行为 |
|----------|------|----------|
| threadIsolation=true + shareSessionInChannel=true + 私聊 | 两者均为 true | threadIsolation 优先（线程 key） |

## 验证方法

```bash
go test ./platform/feishu/ -run TestMakeSessionKey -v
go test ./platform/feishu/ -run TestShouldReplyInThread -v
```