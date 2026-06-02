package feishu

import (
	"testing"

	"github.com/chenhg5/cc-connect/core"
)

func TestRenderTemplateCard_Code(t *testing.T) {
	card := core.NewCard().WithTemplate(core.TemplateCode).Markdown("```go\nfunc main() {}\n```").Note("main.go").Build()
	result := renderTemplateCard(card, "sess1")

	header := result["header"].(map[string]any)
	template := header["template"].(string)
	if template != "blue" {
		t.Errorf("code template header color = %q, want %q", template, "blue")
	}

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
	card := core.NewCard().WithTemplate(core.CardTemplate("unknown")).Markdown("content").Build()
	result := renderTemplateCard(card, "sess1")

	// Unknown template should fall back to generic rendering
	// Generic rendering uses default "blue" if no header color set
	if _, ok := result["header"]; ok {
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

func TestRenderTemplateCard_HeaderOverride(t *testing.T) {
	// When Card has explicit Header, template preset should allow override
	card := core.NewCard().WithTemplate(core.TemplateExplain).Title("Custom Title", "yellow").Markdown("content").Build()
	result := renderTemplateCard(card, "sess1")

	header := result["header"].(map[string]any)
	template := header["template"].(string)
	if template != "yellow" {
		t.Errorf("template with explicit header color override = %q, want %q", template, "yellow")
	}
	titleObj := header["title"].(map[string]any)
	title := titleObj["content"].(string)
	if title != "Custom Title" {
		t.Errorf("template with explicit header title override = %q, want %q", title, "Custom Title")
	}
}