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

// templateHeaderColor maps a CardTemplate to its default header color.
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