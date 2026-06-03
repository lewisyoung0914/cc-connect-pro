# 飞书消息模板与工具调用聚合 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为飞书平台增加 CardTemplate 消息模板系统，根据内容类型自动选择模板，增强工具调用聚合卡片和各类回复的视觉效果。

**Architecture:** 在 core/ 层定义 CardTemplate 类型（平台无关），在 feishu/ 层实现模板预设渲染（平台特定）。引擎在 EventResult 时根据内容分析选择模板，通过 CardSender 发送模板化的卡片。工具调用聚合已有 compactProgressWriter + buildRichCard 机制，本次增强其视觉布局。

**Tech Stack:** Go, 飞书 Interactive Card v1/v2 JSON, core.Card builder

---

## File Structure

| File | Responsibility |
|---|---|
| `core/card.go` | CardTemplate 类型定义、Card.Template 字段、CardBuilder.WithTemplate() |
| `core/template.go` | SelectTemplate() 内容分析函数、codeRatio()、isErrorResponse() |
| `core/card_test.go` | CardTemplate 和 WithTemplate() 的单元测试 |
| `core/template_test.go` | SelectTemplate() 的单元测试 |
| `platform/feishu/card_template.go` | 飞书模板预设渲染函数（renderCodeTemplate 等） |
| `platform/feishu/card_template_test.go` | 飞书模板渲染的单元测试 |
| `platform/feishu/card.go` | 修改 renderCardMap() 调用模板渲染 |
| `platform/feishu/feishu.go` | 增强 buildProgressCardJSONFromPayload() 的步骤列表视觉 |
| `core/i18n.go` | 添加模板标题的 i18n MsgKey |
| `core/engine.go` | EventResult 中集成模板选择、sendPermissionPrompt 使用模板 |

---

### Task 1: CardTemplate 类型与 Card 扩展

**Files:**
- Modify: `core/card.go`
- Modify: `core/card_test.go`

- [ ] **Step 1: 在 card.go 中添加 CardTemplate 类型定义**

在 `Card` 结构体之前添加：

```go
// CardTemplate identifies the visual template a card should use when rendered
// by platform-specific adapters. The template controls header color, title,
// layout structure, and interactive elements.
type CardTemplate string

const (
	TemplateCode       CardTemplate = "code"       // 代码回复
	TemplateExplain    CardTemplate = "explain"    // 解释/分析
	TemplateProgress   CardTemplate = "progress"   // 工具调用进度
	TemplatePermission CardTemplate = "permission" // 权限请求
	TemplateError      CardTemplate = "error"      // 错误/警告
)
```

- [ ] **Step 2: 在 Card 结构体中添加 Template 字段**

修改 Card 结构体：

```go
type Card struct {
	Header   *CardHeader
	Elements []CardElement
	Template CardTemplate // optional template type for platform-specific rendering
}
```

- [ ] **Step 3: 在 CardBuilder 中添加 WithTemplate 方法**

在 `Note()` 方法之后添加：

```go
// WithTemplate sets the card template type, which influences how platform
// renderers apply visual presets (header color, title, layout).
func (b *CardBuilder) WithTemplate(t CardTemplate) *CardBuilder {
	b.card.Template = t
	return b
}
```

- [ ] **Step 4: 写单元测试**

在 `core/card_test.go` 中添加（如果文件不存在则创建）：

```go
package core

import "testing"

func TestCardTemplateConstants(t *testing.T) {
	tests := []struct {
		name     string
		template CardTemplate
		want     string
	}{
		{"code", TemplateCode, "code"},
		{"explain", TemplateExplain, "explain"},
		{"progress", TemplateProgress, "progress"},
		{"permission", TemplatePermission, "permission"},
		{"error", TemplateError, "error"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.template) != tt.want {
				t.Errorf("CardTemplate %s = %q, want %q", tt.name, tt.template, tt.want)
			}
		})
	}
}

func TestCardBuilderWithTemplate(t *testing.T) {
	card := NewCard().WithTemplate(TemplateCode).Markdown("hello").Build()
	if card.Template != TemplateCode {
		t.Errorf("card.Template = %q, want %q", card.Template, TemplateCode)
	}
	if len(card.Elements) != 1 {
		t.Errorf("len(card.Elements) = %d, want 1", len(card.Elements))
	}
}

func TestCardBuilderWithTemplateAndTitle(t *testing.T) {
	card := NewCard().Title("代码结果", "blue").WithTemplate(TemplateCode).Markdown("hello").Build()
	if card.Template != TemplateCode {
		t.Errorf("card.Template = %q, want %q", card.Template, TemplateCode)
	}
	if card.Header.Title != "代码结果" {
		t.Errorf("card.Header.Title = %q, want %q", card.Header.Title, "代码结果")
	}
}

func TestCardWithoutTemplate(t *testing.T) {
	card := NewCard().Markdown("hello").Build()
	if card.Template != "" {
		t.Errorf("card.Template = %q, want empty", card.Template)
	}
}

func TestCardRenderTextWithTemplate(t *testing.T) {
	// RenderText should work unchanged when Template is set
	card := NewCard().WithTemplate(TemplateExplain).Markdown("some text").Build()
	text := card.RenderText()
	if text != "some text" {
		t.Errorf("RenderText() = %q, want %q", text, "some text")
	}
}
```

- [ ] **Step 5: 运行测试确认通过**

Run: `cd e:/project/cc-connect-pro && go test ./core/ -run TestCardTemplate -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add core/card.go core/card_test.go
git commit -m "feat: add CardTemplate type and WithTemplate() to CardBuilder"
```

---

### Task 2: 模板选择函数 SelectTemplate

**Files:**
- Create: `core/template.go`
- Create: `core/template_test.go`

- [ ] **Step 1: 写 SelectTemplate 的失败测试**

创建 `core/template_test.go`：

```go
package core

import "testing"

func TestSelectTemplate(t *testing.T) {
	tests := []struct {
		name           string
		content        string
		hasToolSteps   bool
		isPermission   bool
		want           CardTemplate
	}{
		{
			name:         "permission_request",
			content:      "需要执行命令",
			hasToolSteps: false,
			isPermission: true,
			want:         TemplatePermission,
		},
		{
			name:         "progress_with_tool_steps",
			content:      "分析完成",
			hasToolSteps: true,
			isPermission: false,
			want:         TemplateProgress,
		},
		{
			name:         "error_response",
			content:      "Error: command failed with exit code 1",
			hasToolSteps: false,
			isPermission: false,
			want:         TemplateError,
		},
		{
			name:         "code_heavy_response",
			content:      "```go\nfunc main() {\n\tfmt.Println(\"hello\")\n}\n```\nThis is the main function.",
			hasToolSteps: false,
			isPermission: false,
			want:         TemplateCode,
		},
		{
			name:         "plain_explanation",
			content:      "This is a simple explanation of the code changes.",
			hasToolSteps: false,
			isPermission: false,
			want:         TemplateExplain,
		},
		{
			name:         "mixed_content_less_code",
			content:      "Here is a snippet:\n```python\nx = 1\n```\nAnd a long explanation that makes code ratio low.",
			hasToolSteps: false,
			isPermission: false,
			want:         TemplateExplain,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SelectTemplate(tt.content, tt.hasToolSteps, tt.isPermission)
			if got != tt.want {
				t.Errorf("SelectTemplate() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCodeRatio(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    float64
	}{
		{
			name:    "no_code",
			content: "Just plain text no code here",
			want:    0.0,
		},
		{
			name:    "single_code_block",
			content: "```go\nfunc main() {}\n```\nBrief note.",
			want:    0.6, // approximately, code > 50%
		},
		{
			name:    "mostly_text_with_small_code",
			content: "Long explanation text that is definitely more than the code.\n```sh\nls\n```\nMore text after.",
			want:    0.1, // approximately, code < 50%
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := codeRatio(tt.content)
			// Use approximate comparison
			if tt.want == 0.0 && got != 0.0 {
				t.Errorf("codeRatio() = %f, want 0.0", got)
			}
			if tt.want > 0.5 && got <= 0.5 {
				t.Errorf("codeRatio() = %f, want > 0.5", got)
			}
			if tt.want < 0.5 && got >= 0.5 {
				t.Errorf("codeRatio() = %f, want < 0.5", got)
			}
		})
	}
}

func TestIsErrorResponse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{"error_keyword", "Error: something went wrong", true},
		{"failed_keyword", "The command failed to execute", true},
		{"exception_keyword", "Exception occurred in processing", true},
		{"no_error", "Everything worked fine", false},
		{"error_in_code_block", "```sh\nError: exit 1\n```\nThis error happened.", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isErrorResponse(tt.content)
			if got != tt.want {
				t.Errorf("isErrorResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd e:/project/cc-connect-pro && go test ./core/ -run TestSelectTemplate -v`
Expected: FAIL（函数未定义）

- [ ] **Step 3: 实现 SelectTemplate 和辅助函数**

创建 `core/template.go`：

```go
package core

import (
	"regexp"
	"strings"
)

// SelectTemplate analyzes response content and context to determine
// the appropriate card template for rendering.
// Priority: permission > progress > error > code > explain (default).
func SelectTemplate(content string, hasToolSteps bool, isPermission bool) CardTemplate {
	if isPermission {
		return TemplatePermission
	}
	if hasToolSteps {
		return TemplateProgress
	}
	if isErrorResponse(content) {
		return TemplateError
	}
	if codeRatio(content) > 0.5 {
		return TemplateCode
	}
	return TemplateExplain
}

// codeBlockRe matches fenced code blocks to extract their content.
var codeBlockRe = regexp.MustCompile("(?s)```.*?\n(.*?)```")

// codeRatio estimates the proportion of code characters in the content.
// It counts characters inside fenced code blocks (``` delimited)
// and compares to total content length.
func codeRatio(content string) float64 {
	total := len(content)
	if total == 0 {
		return 0.0
	}
	matches := codeBlockRe.FindAllStringSubmatch(content, -1)
	codeLen := 0
	for _, m := range matches {
		codeLen += len(m[1])
	}
	return float64(codeLen) / float64(total)
}

// errorKeywords are lowercase indicators that the content describes an error.
var errorKeywords = []string{"error:", "failed", "exception", "fatal", "panic:", "timeout"}

// isErrorResponse checks whether the content indicates an error outcome.
// It looks for error keywords outside of code blocks to avoid
// false positives from code snippets that contain "error" variable names.
func isErrorResponse(content string) bool {
	// Strip code blocks to avoid matching error keywords inside code
	stripped := codeBlockRe.ReplaceAllString(content, "")
	stripped = strings.ToLower(stripped)
	for _, kw := range errorKeywords {
		if strings.Contains(stripped, kw) {
			return true
		}
	}
	return false
}
```

- [ ] **Step 4: 运行测试确认通过**

Run: `cd e:/project/cc-connect-pro && go test ./core/ -run "TestSelectTemplate|TestCodeRatio|TestIsErrorResponse" -v`
Expected: PASS

- [ ] **Step 5: Commit**

```bash
git add core/template.go core/template_test.go
git commit -m "feat: add SelectTemplate() content analysis function"
```

---

### Task 3: 飞书模板预设渲染

**Files:**
- Create: `platform/feishu/card_template.go`
- Create: `platform/feishu/card_template_test.go`
- Modify: `platform/feishu/card.go` (renderCardMap 调用模板渲染)

- [ ] **Step 1: 写模板渲染的失败测试**

创建 `platform/feishu/card_template_test.go`：

```go
package feishu

import (
	"testing"

	"github.com/chenhg5/cc-connect/core"
)

func TestRenderTemplateCard_Code(t *testing.T) {
	card := core.NewCard().WithTemplate(core.TemplateCode).Markdown("```go\nfunc main() {}\n```").Note("main.go").Build()
	result := renderTemplateCard(card, "sess1")

	// Header should have blue template color
	header := result["header"].(map[string]any)
	template := header["template"].(string)
	if template != "blue" {
		t.Errorf("code template header color = %q, want %q", template, "blue")
	}

	// Should have elements
	elements := result["elements"].([]map[string]any)
	if len(elements) < 1 {
		t.Errorf("code template elements len = %d, want >= 1", len(elements))
	}
}

func TestRenderTemplateCard_Explain(t *testing.T) {
	card := core.NewCard().WithTemplate(core.TemplateExplain).Markdown("解释内容").Build()
	result := renderTemplateCard(card, "sess1")

	header := result["header"].(map[string]any)
	template := header["template"].(string)
	if template != "green" {
		t.Errorf("explain template header color = %q, want %q", template, "green")
	}
}

func TestRenderTemplateCard_Progress(t *testing.T) {
	card := core.NewCard().WithTemplate(core.TemplateProgress).Markdown("结果内容").Build()
	result := renderTemplateCard(card, "sess1")

	header := result["header"].(map[string]any)
	template := header["template"].(string)
	if template != "orange" {
		t.Errorf("progress template header color = %q, want %q", template, "orange")
	}
}

func TestRenderTemplateCard_Permission(t *testing.T) {
	card := core.NewCard().WithTemplate(core.TemplatePermission).Markdown("需要执行 rm -rf").ButtonsEqual(
		core.PrimaryBtn("✓ 允许", "perm:allow"),
		core.DangerBtn("✗ 拒绝", "perm:deny"),
	).Build()
	result := renderTemplateCard(card, "sess1")

	header := result["header"].(map[string]any)
	template := header["template"].(string)
	if template != "red" {
		t.Errorf("permission template header color = %q, want %q", template, "red")
	}
}

func TestRenderTemplateCard_Error(t *testing.T) {
	card := core.NewCard().WithTemplate(core.TemplateError).Markdown("Error: command failed").Build()
	result := renderTemplateCard(card, "sess1")

	header := result["header"].(map[string]any)
	template := header["template"].(string)
	if template != "red" {
		t.Errorf("error template header color = %q, want %q", template, "red")
	}
}

func TestRenderTemplateCard_UnknownFallsBackToGeneric(t *testing.T) {
	card := core.NewCard().WithTemplate(CardTemplate("unknown")).Markdown("content").Build()
	result := renderTemplateCard(card, "sess1")

	// Unknown template should fall back to generic rendering (no template preset)
	// Generic rendering uses default "blue" if no header color set
	if result["header"] != nil {
		header := result["header"].(map[string]any)
		template := header["template"].(string)
		if template != "blue" {
			t.Errorf("unknown template fallback header color = %q, want %q", template, "blue")
		}
	}
}

func TestRenderCardMap_WithTemplate(t *testing.T) {
	card := core.NewCard().WithTemplate(core.TemplateExplain).Markdown("some analysis").Build()
	result := renderCardMap(card, "sess1")

	header := result["header"].(map[string]any)
	template := header["template"].(string)
	if template != "green" {
		t.Errorf("renderCardMap with template header color = %q, want %q", template, "green")
	}
}

func TestRenderCardMap_WithoutTemplateUnchanged(t *testing.T) {
	card := core.NewCard().Title("My Title", "purple").Markdown("content").Build()
	result := renderCardMap(card, "sess1")

	header := result["header"].(map[string]any)
	template := header["template"].(string)
	if template != "purple" {
		t.Errorf("renderCardMap without template header color = %q, want %q", template, "purple")
	}
}
```

- [ ] **Step 2: 运行测试确认失败**

Run: `cd e:/project/cc-connect-pro && go test ./platform/feishu/ -run TestRenderTemplateCard -v`
Expected: FAIL（函数未定义）

- [ ] **Step 3: 创建飞书模板预设渲染文件**

创建 `platform/feishu/card_template.go`：

```go
package feishu

import (
	"github.com/chenhg5/cc-connect/core"
)

// templatePreset defines the Feishu-specific visual defaults for a card template.
type templatePreset struct {
	headerColor string // Feishu header template color
	headerTitle string // Default header title (localized by caller via i18n)
}

// templatePresets maps each CardTemplate to its Feishu visual preset.
var templatePresets = map[core.CardTemplate]templatePreset{
	core.TemplateCode:       {headerColor: "blue", headerTitle: "代码结果"},
	core.TemplateExplain:    {headerColor: "green", headerTitle: "分析结果"},
	core.TemplateProgress:   {headerColor: "orange", headerTitle: "正在执行..."},
	core.TemplatePermission: {headerColor: "red", headerTitle: "需要授权"},
	core.TemplateError:      {headerColor: "red", headerTitle: "⚠ 出错了"},
}

// renderTemplateCard applies template-specific visual presets and then
// renders the card elements using the standard element rendering logic.
func renderTemplateCard(card *core.Card, sessionKey string) map[string]any {
	preset, ok := templatePresets[card.Template]
	if !ok {
		// Unknown template: fall back to generic rendering
		return renderGenericCardMap(card, sessionKey)
	}

	result := map[string]any{
		"config": map[string]any{
			"wide_screen_mode": true,
		},
	}

	// Apply template preset header, but allow explicit Card.Header to override
	headerColor := preset.headerColor
	headerTitle := preset.headerTitle
	if card.Header != nil {
		if card.Header.Color != "" {
			headerColor = card.Header.Color
		}
		if card.Header.Title != "" {
			headerTitle = card.Header.Title
		}
	}
	result["header"] = map[string]any{
		"title":    plainText(headerTitle),
		"template": headerColor,
	}

	// Render elements using standard logic (same as renderCardMap)
	var elements []map[string]any
	for _, elem := range card.Elements {
		elements = append(elements, renderTemplateElement(elem, sessionKey)...)
	}
	if len(elements) == 0 {
		elements = []map[string]any{{"tag": "markdown", "content": " "}}
	}
	result["elements"] = elements

	return result
}

// renderTemplateElement renders a single CardElement into Feishu card JSON
// elements. This is the same logic as in renderCardMap's element loop,
// extracted for reuse by template renderers.
func renderTemplateElement(elem core.CardElement, sessionKey string) []map[string]any {
	switch e := elem.(type) {
	case core.CardMarkdown:
		return []map[string]any{{"tag": "markdown", "content": e.Content}}
	case core.CardDivider:
		return []map[string]any{{"tag": "hr"}}
	case core.CardActions:
		return renderCardActions(e, sessionKey)
	case core.CardNote:
		return []map[string]any{{"tag": "note", "elements": []map[string]any{plainText(e.Text)}}}
	case core.CardListItem:
		return renderCardListItem(e, sessionKey)
	case core.CardSelect:
		return renderCardSelect(e, sessionKey)
	default:
		return nil
	}
}

// renderCardActions renders a CardActions element into Feishu JSON.
func renderCardActions(e core.CardActions, sessionKey string) []map[string]any {
	var actions []map[string]any
	for _, btn := range e.Buttons {
		btnType := btn.Type
		if btnType == "" {
			btnType = "default"
		}
		valMap := map[string]string{"action": btn.Value}
		if sessionKey != "" {
			valMap["session_key"] = sessionKey
		}
		for k, v := range btn.Extra {
			valMap[k] = v
		}
		action := map[string]any{
			"tag":   "button",
			"text":  plainText(btn.Text),
			"type":  btnType,
			"value": valMap,
		}
		if e.Layout == core.CardActionLayoutEqualColumns {
			action["width"] = "fill"
		}
		actions = append(actions, action)
	}
	if len(actions) == 0 {
		return nil
	}
	if e.Layout == core.CardActionLayoutEqualColumns {
		columns := make([]map[string]any, 0, len(actions))
		for _, action := range actions {
			columns = append(columns, map[string]any{
				"tag":              "column",
				"width":            "weighted",
				"weight":           1,
				"vertical_align":   "center",
				"horizontal_align": "center",
				"elements":         []map[string]any{action},
			})
		}
		columnSet := map[string]any{
			"tag":     "column_set",
			"columns": columns,
		}
		if len(actions) == 2 {
			columnSet["flex_mode"] = "bisect"
		}
		return []map[string]any{columnSet}
	}
	return []map[string]any{{"tag": "action", "actions": actions}}
}

// renderCardListItem renders a CardListItem element into Feishu JSON.
func renderCardListItem(e core.CardListItem, sessionKey string) []map[string]any {
	btnType := e.BtnType
	if btnType == "" {
		btnType = "default"
	}
	valMap := map[string]string{"action": e.BtnValue}
	if sessionKey != "" {
		valMap["session_key"] = sessionKey
	}
	for k, v := range e.Extra {
		valMap[k] = v
	}
	return []map[string]any{{
		"tag":       "column_set",
		"flex_mode": "none",
		"columns": []map[string]any{
			{
				"tag":            "column",
				"width":          "weighted",
				"weight":         5,
				"vertical_align": "center",
				"elements":       []map[string]any{{"tag": "markdown", "content": e.Text}},
			},
			{
				"tag":            "column",
				"width":          "auto",
				"vertical_align": "center",
				"elements":       []map[string]any{{"tag": "button", "text": plainText(e.BtnText), "type": btnType, "value": valMap}},
			},
		},
	}}
}

// renderCardSelect renders a CardSelect element into Feishu JSON.
func renderCardSelect(e core.CardSelect, sessionKey string) []map[string]any {
	var options []map[string]any
	for _, opt := range e.Options {
		options = append(options, map[string]any{
			"text":  plainText(opt.Text),
			"value": opt.Value,
		})
	}
	selectElem := map[string]any{
		"tag":         "select_static",
		"placeholder": plainText(e.Placeholder),
		"options":     options,
	}
	if sessionKey != "" {
		selectElem["value"] = map[string]string{"session_key": sessionKey}
	}
	if e.InitValue != "" {
		selectElem["initial_option"] = e.InitValue
	}
	return []map[string]any{{"tag": "action", "actions": []map[string]any{selectElem}}}
}

// renderGenericCardMap renders a card without any template preset,
// equivalent to the original renderCardMap logic (used for fallback).
func renderGenericCardMap(card *core.Card, sessionKey string) map[string]any {
	result := map[string]any{
		"config": map[string]any{
			"wide_screen_mode": true,
		},
	}
	if card == nil {
		return result
	}

	if card.Header != nil && card.Header.Title != "" {
		color := card.Header.Color
		if color == "" {
			color = "blue"
		}
		result["header"] = map[string]any{
			"title":    plainText(card.Header.Title),
			"template": color,
		}
	}

	var elements []map[string]any
	for _, elem := range card.Elements {
		elements = append(elements, renderTemplateElement(elem, sessionKey)...)
	}
	if len(elements) == 0 {
		elements = []map[string]any{{"tag": "markdown", "content": " "}}
	}
	result["elements"] = elements
	return result
}
```

- [ ] **Step 4: 修改 renderCardMap() 调用模板渲染**

修改 `platform/feishu/card.go` 的 `renderCardMap` 函数。在函数开头，如果 `card.Template != ""`，则调用 `renderTemplateCard`：

```go
func renderCardMap(card *core.Card, sessionKey string) map[string]any {
	// If a template is specified, use template-specific rendering
	if card != nil && card.Template != "" {
		if transformed, ok := renderDeleteModeCheckerCard(card, nil); ok {
			return transformed
		}
		return renderTemplateCard(card, sessionKey)
	}

	result := map[string]any{
		"config": map[string]any{
			"wide_screen_mode": true,
		},
	}
	if card == nil {
		return result
	}
	// ... existing code unchanged ...
```

注意：保留 `renderDeleteModeCheckerCard` 的调用优先级，即使有模板也要先检查是否是 delete-mode 卡片。

- [ ] **Step 5: 运行测试确认通过**

Run: `cd e:/project/cc-connect-pro && go test ./platform/feishu/ -run "TestRenderTemplateCard|TestRenderCardMap" -v`
Expected: PASS

- [ ] **Step 6: Commit**

```bash
git add platform/feishu/card_template.go platform/feishu/card_template_test.go platform/feishu/card.go
git commit -m "feat: add Feishu template preset rendering with renderTemplateCard()"
```

---

### Task 4: i18n 模板标题

**Files:**
- Modify: `core/i18n.go`

- [ ] **Step 1: 添加模板标题 MsgKey**

在 `core/i18n.go` 的 MsgKey 常量区域添加：

```go
MsgKeyTemplateCodeTitle       = "template_code_title"
MsgKeyTemplateExplainTitle    = "template_explain_title"
MsgKeyTemplateProgressTitle   = "template_progress_title"
MsgKeyTemplateProgressDone    = "template_progress_done"
MsgKeyTemplatePermissionTitle = "template_permission_title"
MsgKeyTemplateErrorTitle      = "template_error_title"
MsgKeyTemplateWarningTitle    = "template_warning_title"
```

- [ ] **Step 2: 添加各语言的翻译**

在翻译映射中添加所有语言（EN, ZH, ZH-TW, JA, ES）的翻译：

```go
// EN
MsgKeyTemplateCodeTitle:       "Code Result",
MsgKeyTemplateExplainTitle:    "Analysis",
MsgKeyTemplateProgressTitle:   "Running...",
MsgKeyTemplateProgressDone:    "Completed",
MsgKeyTemplatePermissionTitle: "Requires Permission",
MsgKeyTemplateErrorTitle:      "Error",
MsgKeyTemplateWarningTitle:    "Warning",

// ZH
MsgKeyTemplateCodeTitle:       "代码结果",
MsgKeyTemplateExplainTitle:    "分析结果",
MsgKeyTemplateProgressTitle:   "正在执行...",
MsgKeyTemplateProgressDone:    "执行完成",
MsgKeyTemplatePermissionTitle: "需要授权",
MsgKeyTemplateErrorTitle:      "⚠ 出错了",
MsgKeyTemplateWarningTitle:    "⚠ 警告",

// ZH-TW
MsgKeyTemplateCodeTitle:       "程式碼結果",
MsgKeyTemplateExplainTitle:    "分析結果",
MsgKeyTemplateProgressTitle:   "正在執行...",
MsgKeyTemplateProgressDone:    "執行完成",
MsgKeyTemplatePermissionTitle: "需要授權",
MsgKeyTemplateErrorTitle:      "⚠ 出錯了",
MsgKeyTemplateWarningTitle:    "⚠ 警告",

// JA
MsgKeyTemplateCodeTitle:       "コード結果",
MsgKeyTemplateExplainTitle:    "分析結果",
MsgKeyTemplateProgressTitle:   "実行中...",
MsgKeyTemplateProgressDone:    "実行完了",
MsgKeyTemplatePermissionTitle: "許可が必要",
MsgKeyTemplateErrorTitle:      "⚠ エラー",
MsgKeyTemplateWarningTitle:    "⚠ 警告",

// ES
MsgKeyTemplateCodeTitle:       "Resultado de código",
MsgKeyTemplateExplainTitle:    "Análisis",
MsgKeyTemplateProgressTitle:   "Ejecutando...",
MsgKeyTemplateProgressDone:    "Completado",
MsgKeyTemplatePermissionTitle: "Requiere permiso",
MsgKeyTemplateErrorTitle:      "⚠ Error",
MsgKeyTemplateWarningTitle:    "⚠ Advertencia",
```

- [ ] **Step 3: 修改 card_template.go 使用 i18n 标题**

由于 `card_template.go` 在 feishu 包中无法直接访问 engine 的 i18n 实例，模板标题的 i18n 需要在调用 `renderCardMap` 之前由 engine 层处理。因此 `templatePresets` 中的标题作为默认 fallback，而 engine 构建 Card 时通过 `.Title(e.i18n.T(MsgKeyTemplateCodeTitle), "blue")` 设置本地化标题。

这个设计意味着 `WithTemplate()` 不设置标题默认值（保持 core 平台无关），标题由 engine 在构建 Card 时显式设置。`templatePresets` 的 headerTitle 只作为 fallback。

- [ ] **Step 4: 运行测试确认通过**

Run: `cd e:/project/cc-connect-pro && go test ./core/ -run TestI18n -v`
Expected: PASS（确认新的 MsgKey 有翻译）

- [ ] **Step 5: Commit**

```bash
git add core/i18n.go
git commit -m "feat: add i18n MsgKeys for card template titles"
```

---

### Task 5: 引擎 EventResult 集成模板选择

**Files:**
- Modify: `core/engine.go`

这个 Task 修改 engine.go 中 EventResult 的最终消息发送逻辑，在选择发送方式时根据内容选择模板，并通过 CardSender 发送模板化的卡片。

- [ ] **Step 1: 在 EventResult 的最终结果发送处，对支持 CardSender 的平台，使用 SelectTemplate 选择模板并发送模板卡片**

在 `processInteractiveEvents` 中，当 `EventResult` 到来且平台支持 `CardSender` 时，在现有的最终消息发送逻辑中增加模板路径。

找到 engine.go 中 EventResult 处理的最终消息发送位置（大约在 line 4258-4349 区域），在现有逻辑之后增加：

```go
// Template-aware final result: if platform supports CardSender,
// select template based on content and send a template card.
if cardSender, ok := p.(CardSender); ok && fullResponse != "" {
	template := SelectTemplate(fullResponse, len(toolSteps) > 0, false)
	card := NewCard().
		WithTemplate(template).
		Title(e.i18n.T(templateTitleKey(template)), templateHeaderColor(template)).
		Markdown(fullResponse).
		Build()
	// Only send template card if it provides visual improvement over plain text
	if template != TemplateExplain || len(toolSteps) > 0 {
		if err := cardSender.SendCard(e.ctx, replyCtx, card); err != nil {
			slog.Warn("engine: template card send failed, falling back to plain text", "error", err)
			// Fall back to existing plain text send logic
		} else {
			return // Template card sent, skip plain text send
		}
	}
}
```

同时在 core/template.go 中添加辅助函数：

```go
// templateTitleKey maps a CardTemplate to its i18n message key.
func templateTitleKey(t CardTemplate) MsgKey {
	switch t {
	case TemplateCode:
		return MsgKeyTemplateCodeTitle
	case TemplateExplain:
		return MsgKeyTemplateExplainTitle
	case TemplateProgress:
		return MsgKeyTemplateProgressTitle
	case TemplatePermission:
		return MsgKeyTemplatePermissionTitle
	case TemplateError:
		return MsgKeyTemplateErrorTitle
	default:
		return MsgKeyTemplateExplainTitle
	}
}

// templateHeaderColor maps a CardTemplate to its default Feishu header color.
// Note: this is a hint for engine-level Card building; the platform renderer
// applies its own color preset which may override this.
func templateHeaderColor(t CardTemplate) string {
	switch t {
	case TemplateCode:
		return "blue"
	case TemplateExplain:
		return "green"
	case TemplateProgress:
		return "orange"
	case TemplatePermission:
		return "red"
	case TemplateError:
		return "red"
	default:
		return "blue"
	}
}
```

- [ ] **Step 2: 运行编译确认无语法错误**

Run: `cd e:/project/cc-connect-pro && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 3: 运行全量测试**

Run: `cd e:/project/cc-connect-pro && go test ./core/ -v`
Expected: PASS（所有现有测试仍然通过）

- [ ] **Step 4: Commit**

```bash
git add core/engine.go core/template.go
git commit -m "feat: integrate SelectTemplate() into EventResult for template-aware final cards"
```

---

### Task 6: 权限请求使用 TemplatePermission

**Files:**
- Modify: `core/engine.go` (sendPermissionPrompt 函数)

- [ ] **Step 1: 找到 sendPermissionPrompt 函数**

在 engine.go 中找到 `sendPermissionPrompt()` 函数（约 line 9066）。该函数目前构建权限请求卡片时使用 `NewCard().Title(...).Markdown(...).ButtonsEqual(...)`。

- [ ] **Step 2: 修改 sendPermissionPrompt 使用 WithTemplate(TemplatePermission)**

在构建权限卡片的 CardBuilder 链中添加 `.WithTemplate(TemplatePermission)`，并将标题改为使用 i18n：

```go
card := NewCard().
	WithTemplate(TemplatePermission).
	Title(e.i18n.T(MsgKeyTemplatePermissionTitle), "red").
	Markdown(description).
	ButtonsEqual(
		PrimaryBtn(allowLabel, permAllowValue),
		DangerBtn(denyLabel, permDenyValue),
	).
	Build()
```

- [ ] **Step 3: 运行编译和测试**

Run: `cd e:/project/cc-connect-pro && go build ./... && go test ./core/ -v`
Expected: BUILD SUCCESS, PASS

- [ ] **Step 4: Commit**

```bash
git add core/engine.go
git commit -m "feat: use TemplatePermission for permission prompt cards"
```

---

### Task 7: 增强进度卡片视觉布局

**Files:**
- Modify: `platform/feishu/feishu.go` (buildProgressCardJSONFromPayload 函数)

当前 `buildProgressCardJSONFromPayload` 使用简单的文本列表展示进度条目。本次增强将其改为折叠式 column_set，每个工具步骤显示状态图标 + 工具名。

- [ ] **Step 1: 找到 buildProgressCardJSONFromPayload 函数**

在 feishu.go 中找到 `buildProgressCardJSONFromPayload`（约 line 3565-3629）和 `progressStateMeta`（约 line 3270-3289）。

- [ ] **Step 2: 修改 progressStateMeta 支持模板化标题**

将 `progressStateMeta` 的标题改为使用 i18n MsgKey：

```go
func progressStateMeta(state core.ProgressCardState, lang string) (title string, template string) {
	switch state {
	case core.ProgressCardStateCompleted:
		return i18nT(lang, core.MsgKeyTemplateProgressDone), "green"
	case core.ProgressCardStateFailed:
		return i18nT(lang, core.MsgKeyTemplateErrorTitle), "red"
	default:
		return i18nT(lang, core.MsgKeyTemplateProgressTitle), "orange"
	}
}
```

注意：feishu.go 中已有 `i18nT()` 辅助函数用于本地化。

- [ ] **Step 3: 增强 buildProgressCardJSONFromPayload 的步骤列表**

将进度条目的渲染从简单文本改为折叠式 column_set，每个步骤显示状态图标：

```go
// Build tool step rows with status icons
var stepElements []map[string]any
for _, item := range payload.Items {
	if item.Kind == "tool_use" || item.Kind == "tool_result" {
		icon := stepIcon(item)
		stepText := fmt.Sprintf("%s **%s**", icon, item.Tool)
		if item.Kind == "tool_use" && item.Text != "" {
			// Brief input summary (truncated)
			summary := truncateString(item.Text, 80)
			stepText += fmt.Sprintf("\n%s", summary)
		}
		stepElements = append(stepElements, map[string]any{
			"tag":     "markdown",
			"content": stepText,
		})
	}
}

// If there are tool steps, wrap in a collapsible panel
if len(stepElements) > 0 {
	collapsiblePanel := map[string]any{
		"tag":             "collapsible_panel",
		"expanded":        state != core.ProgressCardStateCompleted, // expanded while running, collapsed when done
		"vertical_align":  "center",
		"header":          map[string]any{
			"tag":     "markdown",
			"content": fmt.Sprintf("**%s** (%d steps)", title, len(stepElements)),
		},
		"content":         map[string]any{
			"tag":      "column_set",
			"columns":  stepColumnSet(stepElements),
		},
	}
	elements = append(elements, collapsiblePanel)
}
```

添加辅助函数：

```go
func stepIcon(item core.ProgressCardEntry) string {
	if item.Kind == "tool_result" {
		if item.Success != nil && !*item.Success {
			return "✗"
		}
		return "✓"
	}
	return "⏳"
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func stepColumnSet(steps []map[string]any) []map[string]any {
	columns := make([]map[string]any, 0, len(steps))
	for _, step := range steps {
		columns = append(columns, map[string]any{
			"tag":            "column",
			"width":          "weighted",
			"weight":         1,
			"vertical_align": "center",
			"elements":       []map[string]any{step},
		})
	}
	return columns
}
```

- [ ] **Step 4: 运行编译确认**

Run: `cd e:/project/cc-connect-pro && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 5: Commit**

```bash
git add platform/feishu/feishu.go
git commit -m "feat: enhance progress card with collapsible step list and status icons"
```

---

### Task 8: 错误响应使用 TemplateError

**Files:**
- Modify: `core/engine.go` (EventError 处理)

- [ ] **Step 1: 在 EventError 处理中使用 TemplateError**

在 engine.go 中找到 `EventError` 的处理位置。当平台支持 CardSender 时，构建错误卡片：

```go
if cardSender, ok := p.(CardSender); ok {
	card := NewCard().
		WithTemplate(TemplateError).
		Title(e.i18n.T(MsgKeyTemplateErrorTitle), "red").
		Markdown(errorMessage).
		Build()
	if err := cardSender.SendCard(e.ctx, replyCtx, card); err != nil {
		slog.Warn("engine: error template card send failed", "error", err)
		// Fall back to plain text error send
	} else {
		return
	}
}
```

- [ ] **Step 2: 运行编译和测试**

Run: `cd e:/project/cc-connect-pro && go build ./... && go test ./core/ -v`
Expected: BUILD SUCCESS, PASS

- [ ] **Step 3: Commit**

```bash
git add core/engine.go
git commit -m "feat: use TemplateError for error response cards"
```

---

### Task 9: 集成测试与全量验证

**Files:** 无新文件，验证性测试

- [ ] **Step 1: 运行全量编译**

Run: `cd e:/project/cc-connect-pro && go build ./...`
Expected: BUILD SUCCESS

- [ ] **Step 2: 运行全量测试**

Run: `cd e:/project/cc-connect-pro && go test ./... -v`
Expected: ALL PASS

- [ ] **Step 3: 运行 race detector**

Run: `cd e:/project/cc-connect-pro && go test -race ./core/ ./platform/feishu/`
Expected: PASS（无 race condition）

- [ ] **Step 4: 检查 core/ 无硬编码平台名**

Run: `cd e:/project/cc-connect-pro && grep -n "feishu\|telegram\|discord" core/*.go`
Expected: 无匹配（CardTemplate 和 SelectTemplate 是平台无关的）

- [ ] **Step 5: 最终 Commit**

```bash
git add -A
git commit -m "feat: complete Feishu message template system with 5 template types and progress aggregation enhancement"
```