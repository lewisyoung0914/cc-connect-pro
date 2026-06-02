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