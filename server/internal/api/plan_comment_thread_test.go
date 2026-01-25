package api

import (
	"testing"

	"github.com/satetsu888/agentrace/server/internal/domain"
)

func TestStripInlineMarkdown(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "plain text",
			input:    "hello world",
			expected: "hello world",
		},
		{
			name:     "inline code",
			input:    "this is `code` here",
			expected: "this is code here",
		},
		{
			name:     "multiple inline codes",
			input:    "use `foo` and `bar` functions",
			expected: "use foo and bar functions",
		},
		{
			name:     "bold text",
			input:    "this is **bold** text",
			expected: "this is bold text",
		},
		{
			name:     "italic text",
			input:    "this is *italic* text",
			expected: "this is italic text",
		},
		{
			name:     "underscore bold",
			input:    "this is __bold__ text",
			expected: "this is bold text",
		},
		{
			name:     "strikethrough",
			input:    "this is ~~deleted~~ text",
			expected: "this is deleted text",
		},
		{
			name:     "mixed formatting",
			input:    "use `code` with **bold** and *italic*",
			expected: "use code with bold and italic",
		},
		{
			name:     "Japanese with inline code",
			input:    "これは`test`というコードです",
			expected: "これはtestというコードです",
		},
		{
			name:     "simple link",
			input:    "see [here](https://example.com) for details",
			expected: "see here for details",
		},
		{
			name:     "link with title",
			input:    "see [here](https://example.com \"title\") for details",
			expected: "see here for details",
		},
		{
			name:     "multiple links",
			input:    "visit [site1](url1) and [site2](url2)",
			expected: "visit site1 and site2",
		},
		{
			name:     "link with inline code",
			input:    "use [`code`](url) function",
			expected: "use `code` function", // Note: inline code inside link text is preserved
		},
		{
			name:     "Japanese with link",
			input:    "詳しくは[こちら](https://example.com)を参照",
			expected: "詳しくはこちらを参照",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stripped, _ := stripInlineMarkdown(tt.input)
			if stripped != tt.expected {
				t.Errorf("stripInlineMarkdown(%q) = %q, want %q", tt.input, stripped, tt.expected)
			}
		})
	}
}

func TestFindThreadPosition_InlineCode(t *testing.T) {
	tests := []struct {
		name          string
		body          string
		targetText    string
		contextBefore string
		contextAfter  string
		expectFound   bool
	}{
		{
			name:          "exact match without inline code",
			body:          "これはtestというコードです",
			targetText:    "はtestと",
			contextBefore: "これ",
			contextAfter:  "いうコード",
			expectFound:   true,
		},
		{
			name:          "match across inline code",
			body:          "これは`test`というコードです",
			targetText:    "はtestと",
			contextBefore: "これ",
			contextAfter:  "いうコード",
			expectFound:   true,
		},
		{
			name:          "match with only target text across inline code",
			body:          "これは`test`というコードです",
			targetText:    "はtestと",
			contextBefore: "",
			contextAfter:  "",
			expectFound:   true,
		},
		{
			name:          "match with bold text",
			body:          "これは**test**というコードです",
			targetText:    "はtestと",
			contextBefore: "これ",
			contextAfter:  "いうコード",
			expectFound:   true,
		},
		{
			name:          "not found",
			body:          "これは`test`というコードです",
			targetText:    "notexist",
			contextBefore: "",
			contextAfter:  "",
			expectFound:   false,
		},
		{
			name:          "match across link",
			body:          "詳しくは[こちら](https://example.com)を参照",
			targetText:    "はこちらを",
			contextBefore: "詳しく",
			contextAfter:  "参照",
			expectFound:   true,
		},
		{
			name:          "match with link text only",
			body:          "see [documentation](https://docs.example.com) for more",
			targetText:    "documentation",
			contextBefore: "see ",
			contextAfter:  " for",
			expectFound:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			thread := &domain.PlanCommentThread{
				TargetText:    tt.targetText,
				ContextBefore: tt.contextBefore,
				ContextAfter:  tt.contextAfter,
			}
			pos := FindThreadPosition(tt.body, thread)
			if pos.Found != tt.expectFound {
				t.Errorf("FindThreadPosition() found = %v, want %v", pos.Found, tt.expectFound)
			}
		})
	}
}

func TestFindThreadPosition_PositionAccuracy(t *testing.T) {
	// Test that the returned position correctly points to the original markdown
	body := "これは`test`というコードです"
	thread := &domain.PlanCommentThread{
		TargetText:    "はtestと",
		ContextBefore: "これ",
		ContextAfter:  "いうコード",
	}

	pos := FindThreadPosition(body, thread)
	if !pos.Found {
		t.Fatal("expected position to be found")
	}

	// The position should point to the text including the backticks
	// "これは`test`と" starts at byte offset 6 (after "これ" = 6 bytes)
	// and ends at the end of "と"
	extracted := body[pos.StartOffset:pos.EndOffset]

	// The extracted text should include the inline code markers
	expected := "は`test`と"
	if extracted != expected {
		t.Errorf("extracted text = %q, want %q (offsets: %d-%d)", extracted, expected, pos.StartOffset, pos.EndOffset)
	}
}
