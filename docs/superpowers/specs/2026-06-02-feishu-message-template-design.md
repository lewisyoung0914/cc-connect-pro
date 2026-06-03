# 飞书消息模板与工具调用聚合设计

日期: 2026-06-02

## 问题

当前飞书消息展示存在两个问题：

1. **工具调用消息刷屏** — 每个 `EventToolUse` / `EventToolResult` 都单独发送一条消息，导致聊天界面被工具调用淹没
2. **排版不美观** — 回复内容缺乏视觉结构，无法区分代码、解释、错误等不同类型的内容

## 方案

**方案 A：纯飞书卡片模板**（已选定）

在现有 `Card` builder 上扩展 `CardTemplate` 机制。定义 5 种模板类型，每种有预设的飞书卡片 JSON 结构。将工具调用聚合到一张进度卡片中，替代逐条发送。引擎在 `EventResult` 时根据内容特征自动选择模板。

无需跨平台 DSL 或飞书模板 API 审批流程。仅飞书平台受益，其他平台保持现有行为不变。

---

## 设计

### 第一部分：CardTemplate 机制

#### `core/card.go` 新增类型

```go
type CardTemplate string

const (
    TemplateCode       CardTemplate = "code"       // 代码回复
    TemplateExplain    CardTemplate = "explain"    // 解释/分析
    TemplateProgress   CardTemplate = "progress"   // 工具调用进度
    TemplatePermission CardTemplate = "permission" // 权限请求
    TemplateError      CardTemplate = "error"      // 错误/警告
)
```

#### Card 结构体扩展

在 `Card` 结构体中新增 `Template CardTemplate` 字段。在 `CardBuilder` 中新增 `WithTemplate(t CardTemplate)` 方法。

#### 模板选择逻辑

引擎在 `EventResult` 时根据内容自动选择模板：

| 条件 | 模板 |
|---|---|
| 内容包含代码块且代码占比 > 50% | TemplateCode |
| 本轮对话包含 `EventToolUse` 步骤 | TemplateProgress |
| 权限请求事件 | TemplatePermission |
| 内容含 error/失败关键词 | TemplateError |
| 默认 | TemplateExplain |

#### 渲染流程

飞书适配器的 `renderCardMap()` 检查 `Card.Template` 字段，根据模板类型选择预设的飞书卡片 JSON 结构（颜色、布局、按钮配置）。未指定模板的卡片仍走原有渲染逻辑，保持不变。

### 第二部分：工具调用聚合

#### 当前行为

每个 `EventToolUse` / `EventToolResult` 都单独发送一条消息到飞书。

#### 新行为

工具调用在飞书平台上聚合到一张进度卡片中：

1. **开始阶段**：第一个 `EventToolUse` 创建并发送一张进度聚合卡片。卡片内容：
   - 标题：「正在执行...」
   - 折叠式 column_set：每个步骤显示工具名 + 状态图标（✓/⏳/✗）
   - 默认折叠，用户可展开查看每个步骤的简要输入摘要

2. **更新阶段**：后续的 `EventToolUse` / `EventToolResult` 通过飞书卡片 `update` API 更新同一张卡片，追加新步骤并更新已有步骤的状态。

3. **完成阶段**：`EventResult` 到来时，进度卡片更新为「执行完成」状态（所有步骤显示 ✓），然后引擎根据结果内容类型选择对应模板，发送**最终结果卡片**。

#### 关键改动点

- `engine.go` 事件循环：对飞书平台，`EventToolUse/EventToolResult` 不再调用 `p.Send()`，而是收集到 `toolSteps[]` 并更新进度卡片
- 飞书适配器：新增 `ProgressCardUpdater` 实现增量卡片更新
- 现有的 `CardSender` 和流式预览机制保持不变，进度卡片是在其之上新增的路径

#### 降级策略

如果飞书卡片更新 API 调用失败，回退到逐条发送的旧行为。

### 第三部分：模板类型定义

#### TemplateCode — 代码回复

- 主题色：深蓝 (`blue`)
- 标题：「代码结果」
- 内容区：Markdown 展示代码（飞书卡片原生语法高亮）
- 底部：Note 显示文件路径或上下文信息
- 按钮：无

#### TemplateExplain — 解释/分析

- 主题色：绿色 (`green`)
- 标题：「分析结果」
- 内容区：Markdown 段落文本，段落间用分隔线区分
- 按钮：无

#### TemplateProgress — 工具调用进度

- 主题色：橙色 (`orange`)
- 标题：「正在执行...」 → 完成后变为「执行完成」
- 内容区：折叠式 column_set — 每个工具步骤一行，显示 `✓ 工具名` / `⏳ 工具名` / `✗ 工具名`
- 默认折叠，展开后显示每个步骤的简要输入摘要
- 按钮：完成阶段可选「查看详情」按钮（展开所有步骤）
- 通过飞书卡片 update API 实时更新步骤状态

#### TemplatePermission — 权限请求

- 主题色：红色 (`red`)
- 标题：「需要授权」
- 内容区：请求操作的说明文字 + 被请求的命令/文件路径
- 按钮：两个交互按钮 — 「✓ 允许」和「✗ 拒绝」
- 交互：使用现有的 `handleCardAction()` 回调机制

#### TemplateError — 错误/警告

- 主题色：错误用红色 (`red`)，警告用橙色 (`orange`)
- 标题：「⚠ 出错了」 / 「⚠ 警告」
- 内容区：错误信息 + 可选的修复建议
- 按钮：可选「重新尝试」按钮

#### 飞书渲染细节

每种模板在 `renderCardMap()` 中对应一个预设的卡片 JSON 结构片段：header 配色、column_set 布局、action 按钮组。通过 `Card.Template` 字段传递模板类型，飞书渲染时读取并选择对应的预设结构。

---

## 影响范围

- `core/card.go`：Card 结构体、CardBuilder、CardTemplate 类型
- `core/engine.go`：事件循环工具调用聚合逻辑、模板选择逻辑
- `platform/feishu/card.go`：模板感知的 `renderCardMap()`、ProgressCardUpdater
- 其他平台：无变化