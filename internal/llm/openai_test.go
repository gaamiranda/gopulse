package llm

import (
	"strings"
	"testing"
)

func TestParsePRContent(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantTitle   string
		wantDescHas string // substring that should be in description
	}{
		{
			name: "Standard format with Title: prefix",
			content: `Title: Add user authentication

Description:
This PR adds JWT-based authentication.

Key changes:
- Add auth middleware
- Add login endpoint`,
			wantTitle:   "Add user authentication",
			wantDescHas: "JWT-based authentication",
		},
		{
			name: "Title without prefix",
			content: `Add new feature

This adds a cool new feature to the app.`,
			wantTitle:   "Add new feature",
			wantDescHas: "cool new feature",
		},
		{
			name: "Title with quotes",
			content: `Title: "Fix bug in parser"

Description:
Fixed the parsing issue.`,
			wantTitle:   "Fix bug in parser",
			wantDescHas: "parsing issue",
		},
		{
			name: "Lowercase title prefix",
			content: `title: Update dependencies

description:
Updated all npm packages.`,
			wantTitle:   "Update dependencies",
			wantDescHas: "npm packages",
		},
		{
			name: "Title with markdown header",
			content: `# Refactor database layer

This refactors the database.`,
			wantTitle:   "Refactor database layer",
			wantDescHas: "refactors",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePRContent(tt.content)

			if got.Title != tt.wantTitle {
				t.Errorf("parsePRContent() Title = %q, want %q", got.Title, tt.wantTitle)
			}

			if !strings.Contains(got.Description, tt.wantDescHas) {
				t.Errorf("parsePRContent() Description = %q, want to contain %q", got.Description, tt.wantDescHas)
			}
		})
	}
}

func TestBuildCommitPrompt(t *testing.T) {
	diff := "diff --git a/file.go b/file.go\n+new line"
	prompt := buildCommitPrompt(diff)

	if !strings.Contains(prompt, diff) {
		t.Errorf("buildCommitPrompt() should contain the diff")
	}

	if !strings.Contains(prompt, "Generate") {
		t.Errorf("buildCommitPrompt() should contain generation instruction")
	}
}

func TestBuildPRPrompt(t *testing.T) {
	commits := "abc123 First commit\ndef456 Second commit"
	diff := "diff --git a/file.go b/file.go\n+new line"

	prompt := buildPRPrompt(commits, diff)

	if !strings.Contains(prompt, commits) {
		t.Errorf("buildPRPrompt() should contain the commits")
	}

	if !strings.Contains(prompt, diff) {
		t.Errorf("buildPRPrompt() should contain the diff")
	}
}

func TestParseDescription(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  string
	}{
		{
			name:  "Simple description",
			lines: []string{"This is a description."},
			want:  "This is a description.",
		},
		{
			name:  "Description with header",
			lines: []string{"Description:", "This is the content."},
			want:  "This is the content.",
		},
		{
			name:  "Description with empty lines at start",
			lines: []string{"", "", "Content here"},
			want:  "Content here",
		},
		{
			name:  "Multi-line description",
			lines: []string{"Line 1", "Line 2", "Line 3"},
			want:  "Line 1\nLine 2\nLine 3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseDescription(tt.lines)
			if got != tt.want {
				t.Errorf("parseDescription() = %q, want %q", got, tt.want)
			}
		})
	}
}
