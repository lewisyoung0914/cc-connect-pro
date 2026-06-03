# 飞书私聊线程隔离 — 总体设计

**日期**: 2026-06-02
**状态**: 草稿

## 问题

私聊中所有消息共享一个 `SessionKey`（`feishu:{chatID}:{userID}`），通过 `Session.TryLock()` 串行处理。用户无法在 agent 处理任务 A 的同时启动任务 B。

## 方案

将群聊已有的 `thread_isolation` 扩展到私聊。每个飞书"回复链"成为独立的 agent session，多个回复链并发运行。

详细设计见：[2026-06-02-feishu-p2p-thread-isolation-design.md](2026-06-02-feishu-p2p-thread-isolation-design.md)

## 子项目拆分

| # | 子项目 | 核心变更 | 依赖 | 规格文件 |
|---|--------|----------|------|----------|
| 1 | Session 路由 | `makeSessionKey` 条件扩展 | 无 | `sub1-session-routing.md` |
| 2 | 回复行为 | `replyContext` + `shouldReplyInThread` | 子项目 1 | `sub2-reply-behavior.md` |
| 3 | 单元测试 | 测试用例覆盖私聊线程隔离 | 子项目 1、2 | `sub3-testing.md` |

每个子项目可独立验证，按序实施。