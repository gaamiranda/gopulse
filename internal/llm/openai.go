package llm

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	openai "github.com/sashabaranov/go-openai"
)

const (
	// DefaultModel is the default OpenAI model to use
	DefaultModel = openai.GPT4o

	// maxDiffLength is the maximum length of diff to send to the API
	maxDiffLength = 10000
)

// Client wraps the OpenAI client
type Client struct {
	client *openai.Client
	model  string
}

// PRContent holds the generated PR title and description
type PRContent struct {
	Title       string
	Description string
}

// NewClient creates a new OpenAI client from environment variable
func NewClient() (*Client, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable is not set")
	}

	return &Client{
		client: openai.NewClient(apiKey),
		model:  DefaultModel,
	}, nil
}

// GenerateCommitMessage generates a commit message from a diff
func (c *Client) GenerateCommitMessage(diff string) (string, error) {
	// Truncate diff if too long
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n\n[diff truncated due to length]"
	}

	prompt := buildCommitPrompt(diff)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: commitSystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.3,
			MaxTokens:   200,
		},
	)

	if err != nil {
		return "", formatAPIError(err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	message := strings.TrimSpace(resp.Choices[0].Message.Content)

	// Remove any quotes if the model wrapped the message
	message = strings.Trim(message, "\"'`")

	return message, nil
}

// GeneratePRContent generates a PR title and description
func (c *Client) GeneratePRContent(commits string, diff string) (*PRContent, error) {
	// Truncate diff if too long
	if len(diff) > maxDiffLength {
		diff = diff[:maxDiffLength] + "\n\n[diff truncated due to length]"
	}

	prompt := buildPRPrompt(commits, diff)

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	resp, err := c.client.CreateChatCompletion(
		ctx,
		openai.ChatCompletionRequest{
			Model: c.model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: prSystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
			Temperature: 0.3,
			MaxTokens:   500,
		},
	)

	if err != nil {
		return nil, formatAPIError(err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no response from OpenAI")
	}

	content := resp.Choices[0].Message.Content
	return parsePRContent(content), nil
}

// buildCommitPrompt creates the user prompt for commit message generation
func buildCommitPrompt(diff string) string {
	return fmt.Sprintf(`Generate a commit message for the following changes:

%s`, diff)
}

// buildPRPrompt creates the user prompt for PR content generation
func buildPRPrompt(commits, diff string) string {
	return fmt.Sprintf(`Generate a PR title and description for the following changes.

Commits:
%s

Diff:
%s`, commits, diff)
}

// parsePRContent parses the PR response into title and description
func parsePRContent(content string) *PRContent {
	lines := strings.Split(strings.TrimSpace(content), "\n")

	pr := &PRContent{}

	// Find title line
	for i, line := range lines {
		line = strings.TrimSpace(line)

		// Look for title markers
		if strings.HasPrefix(strings.ToLower(line), "title:") {
			pr.Title = strings.TrimSpace(strings.TrimPrefix(line, "Title:"))
			pr.Title = strings.TrimSpace(strings.TrimPrefix(pr.Title, "title:"))
			pr.Title = strings.Trim(pr.Title, "\"'`")

			// Rest is description
			if i+1 < len(lines) {
				descLines := lines[i+1:]
				pr.Description = parseDescription(descLines)
			}
			break
		}

		// If first non-empty line and no "Title:" prefix, assume it's the title
		if pr.Title == "" && line != "" {
			pr.Title = strings.Trim(line, "\"'`#")
			if i+1 < len(lines) {
				descLines := lines[i+1:]
				pr.Description = parseDescription(descLines)
			}
			break
		}
	}

	// Clean up title - remove any "Title:" prefix that might remain
	pr.Title = strings.TrimPrefix(pr.Title, "Title:")
	pr.Title = strings.TrimPrefix(pr.Title, "title:")
	pr.Title = strings.TrimSpace(pr.Title)

	return pr
}

// parseDescription cleans up the description portion
func parseDescription(lines []string) string {
	var result []string
	foundContent := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Skip "Description:" header
		if strings.HasPrefix(strings.ToLower(trimmed), "description:") {
			trimmed = strings.TrimPrefix(trimmed, "Description:")
			trimmed = strings.TrimPrefix(trimmed, "description:")
			trimmed = strings.TrimSpace(trimmed)
			if trimmed != "" {
				result = append(result, trimmed)
				foundContent = true
			}
			continue
		}

		// Skip empty lines at the beginning
		if !foundContent && trimmed == "" {
			continue
		}

		foundContent = true
		result = append(result, line)
	}

	return strings.TrimSpace(strings.Join(result, "\n"))
}

// System prompts for the LLM
const commitSystemPrompt = `You are a helpful assistant that generates concise git commit messages.

Rules:
1. Write in imperative mood (e.g., "Add feature" not "Added feature")
2. Keep the message under 72 characters
3. Focus on WHAT changed and WHY, not HOW
4. Be specific but concise
5. Do not include any prefixes like "feat:", "fix:", etc.
6. Return ONLY the commit message, nothing else
7. Do not wrap the message in quotes

Examples of good commit messages:
- Add user authentication with JWT tokens
- Fix memory leak in connection pool
- Update dependencies to latest versions
- Refactor database queries for better performance`

const prSystemPrompt = `You are a helpful assistant that generates GitHub Pull Request titles and descriptions.

Rules:
1. Title should be concise (under 72 characters) and in imperative mood
2. Description should include:
   - A brief summary (1-2 sentences)
   - Key changes as bullet points
   - Any breaking changes or important notes (if applicable)
3. Be specific and helpful for reviewers
4. Format your response as:
   Title: <title here>
   
   Description:
   <description here>

Example response:
Title: Add user authentication system

Description:
This PR introduces JWT-based authentication for the API.

Key changes:
- Add auth middleware for protected routes
- Implement login and logout endpoints
- Add user session management
- Update API documentation

Note: Requires REDIS_URL environment variable for session storage.`

// formatAPIError converts OpenAI API errors into user-friendly messages
func formatAPIError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for network errors
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return fmt.Errorf("request timed out - please check your internet connection and try again")
		}
		return fmt.Errorf("network error - please check your internet connection: %w", err)
	}

	// Check for common API errors by message content
	switch {
	case strings.Contains(errStr, "401") || strings.Contains(errStr, "invalid_api_key"):
		return fmt.Errorf(`invalid OpenAI API key

Please check your OPENAI_API_KEY:
  1. Verify the key is correct at https://platform.openai.com/api-keys
  2. Make sure the key hasn't been revoked
  3. Check that your .env file has the correct format: OPENAI_API_KEY=sk-...`)

	case strings.Contains(errStr, "429"):
		return fmt.Errorf(`OpenAI API rate limit exceeded

You've made too many requests. Please:
  1. Wait a few minutes and try again
  2. Check your usage at https://platform.openai.com/usage
  3. Consider upgrading your OpenAI plan if this persists`)

	case strings.Contains(errStr, "500") || strings.Contains(errStr, "502") || strings.Contains(errStr, "503"):
		return fmt.Errorf("OpenAI service is temporarily unavailable - please try again in a few minutes")

	case strings.Contains(errStr, "insufficient_quota"):
		return fmt.Errorf(`OpenAI API quota exceeded

Your API key has run out of credits. Please:
  1. Check your billing at https://platform.openai.com/account/billing
  2. Add credits or upgrade your plan`)

	case strings.Contains(errStr, "context_length_exceeded"):
		return fmt.Errorf("the diff is too large for the AI model - try staging fewer files")

	default:
		return fmt.Errorf("OpenAI API error: %w", err)
	}
}

// requestTimeout is the timeout for API requests
const requestTimeout = 30 * time.Second
