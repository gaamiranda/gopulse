package github

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/google/go-github/v60/github"
	"golang.org/x/oauth2"
)

// Client wraps the GitHub API client
type Client struct {
	client *github.Client
	ctx    context.Context
}

// RepoInfo holds repository owner and name
type RepoInfo struct {
	Owner string
	Name  string
}

// PRResult holds the result of creating a PR
type PRResult struct {
	Number int
	URL    string
}

// NewClient creates a new GitHub client from environment variable
func NewClient() (*Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &Client{
		client: github.NewClient(tc),
		ctx:    ctx,
	}, nil
}

// CreatePR creates a new pull request
func (c *Client) CreatePR(owner, repo, base, head, title, body string) (*PRResult, error) {
	pr, _, err := c.client.PullRequests.Create(c.ctx, owner, repo, &github.NewPullRequest{
		Title: github.String(title),
		Body:  github.String(body),
		Head:  github.String(head),
		Base:  github.String(base),
	})
	if err != nil {
		return nil, formatGitHubError(err)
	}

	return &PRResult{
		Number: pr.GetNumber(),
		URL:    pr.GetHTMLURL(),
	}, nil
}

// formatGitHubError converts GitHub API errors into user-friendly messages
func formatGitHubError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for GitHub error response
	if ghErr, ok := err.(*github.ErrorResponse); ok {
		switch ghErr.Response.StatusCode {
		case 401:
			return fmt.Errorf(`GitHub authentication failed

Please check your GITHUB_TOKEN:
  1. Verify the token is correct at https://github.com/settings/tokens
  2. Make sure the token hasn't expired
  3. Ensure the token has 'repo' scope`)

		case 403:
			if strings.Contains(errStr, "rate limit") {
				return fmt.Errorf(`GitHub API rate limit exceeded

Please wait a few minutes and try again.
Check your rate limit at: https://api.github.com/rate_limit`)
			}
			return fmt.Errorf(`GitHub access denied

Your token may not have sufficient permissions.
Ensure your GITHUB_TOKEN has 'repo' scope.`)

		case 404:
			return fmt.Errorf(`repository not found or not accessible

Please verify:
  1. The repository exists on GitHub
  2. Your GITHUB_TOKEN has access to this repository
  3. The remote URL is correct`)

		case 422:
			if strings.Contains(errStr, "already exists") {
				return fmt.Errorf("a pull request already exists for this branch")
			}
			if strings.Contains(errStr, "No commits between") {
				return fmt.Errorf("no changes between the base branch and your branch - nothing to merge")
			}
			return fmt.Errorf("GitHub validation error: %s", ghErr.Message)
		}
	}

	return fmt.Errorf("GitHub API error: %w", err)
}

// ParseRemoteURL extracts owner and repo from a git remote URL
// Supports both HTTPS and SSH formats:
// - https://github.com/owner/repo.git
// - git@github.com:owner/repo.git
// - https://github.com/owner/repo
// - git@github.com:owner/repo
func ParseRemoteURL(url string) (*RepoInfo, error) {
	url = strings.TrimSpace(url)

	// SSH format: git@github.com:owner/repo.git
	sshPattern := regexp.MustCompile(`git@github\.com[:/]([^/]+)/([^/]+?)(?:\.git)?$`)
	if matches := sshPattern.FindStringSubmatch(url); len(matches) == 3 {
		return &RepoInfo{
			Owner: matches[1],
			Name:  matches[2],
		}, nil
	}

	// HTTPS format: https://github.com/owner/repo.git
	httpsPattern := regexp.MustCompile(`https?://github\.com/([^/]+)/([^/]+?)(?:\.git)?$`)
	if matches := httpsPattern.FindStringSubmatch(url); len(matches) == 3 {
		return &RepoInfo{
			Owner: matches[1],
			Name:  matches[2],
		}, nil
	}

	return nil, fmt.Errorf("could not parse GitHub remote URL: %s", url)
}

// GetDefaultBranch fetches the default branch for a repository
func (c *Client) GetDefaultBranch(owner, repo string) (string, error) {
	repository, _, err := c.client.Repositories.Get(c.ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get repository info: %w", err)
	}

	return repository.GetDefaultBranch(), nil
}

// BranchExists checks if a branch exists on the remote
func (c *Client) BranchExists(owner, repo, branch string) (bool, error) {
	_, _, err := c.client.Repositories.GetBranch(c.ctx, owner, repo, branch, 0)
	if err != nil {
		// Check if it's a 404 (branch not found)
		if _, ok := err.(*github.ErrorResponse); ok {
			return false, nil
		}
		return false, fmt.Errorf("failed to check branch: %w", err)
	}
	return true, nil
}
