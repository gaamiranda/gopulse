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
		return nil, fmt.Errorf("failed to create pull request: %w", err)
	}

	return &PRResult{
		Number: pr.GetNumber(),
		URL:    pr.GetHTMLURL(),
	}, nil
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
