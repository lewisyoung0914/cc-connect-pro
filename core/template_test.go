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
			want:    0.6,
		},
		{
			name:    "mostly_text_with_small_code",
			content: "Long explanation text that is definitely more than the code.\n```sh\nls\n```\nMore text after.",
			want:    0.1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := codeRatio(tt.content)
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