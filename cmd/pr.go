package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/user/vibe/internal/git"
	"github.com/user/vibe/internal/github"
	"github.com/user/vibe/internal/llm"
	"github.com/user/vibe/internal/ui"
)

var prCmd = &cobra.Command{
	Use:   "pr",
	Short: "Create a GitHub PR with AI-generated title and description",
	Long: `Creates a GitHub Pull Request with an AI-generated title and description.

The command will:
1. Detect your current branch and the base branch (main/master)
2. Get the commits ahead of the base branch
3. Generate a diff of all changes
4. Use OpenAI to generate a PR title and description
5. Show you the PR details for review
6. Allow you to accept, edit, or cancel
7. Push your branch if needed
8. Create the PR on GitHub

Requirements:
- Must be in a git repository with a GitHub remote
- Must be on a feature branch (not main/master)
- Must have commits ahead of the base branch
- OPENAI_API_KEY environment variable must be set
- GITHUB_TOKEN environment variable must be set`,
	RunE: runPR,
}

func init() {
	rootCmd.AddCommand(prCmd)
}

func runPR(cmd *cobra.Command, args []string) error {
	// Check for required environment variables
	if err := checkOpenAIKey(); err != nil {
		return err
	}
	if err := checkGitHubToken(); err != nil {
		return err
	}

	// Open the git repository
	repo, err := git.OpenCurrent()
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	// Get current branch
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	// Get default branch (main or master)
	baseBranch, err := repo.GetDefaultBranch()
	if err != nil {
		return fmt.Errorf("failed to detect base branch: %w", err)
	}

	// Check we're not on the base branch
	if currentBranch == baseBranch {
		return fmt.Errorf(`cannot create PR from %s branch

Create a feature branch first:
  git checkout -b feature/my-feature`, baseBranch)
	}

	ui.ShowInfo(fmt.Sprintf("Analyzing branch '%s' against '%s'...", currentBranch, baseBranch))

	// Get commits ahead of base
	commits, err := repo.GetCommitsAhead(baseBranch)
	if err != nil {
		return fmt.Errorf("failed to get commits: %w", err)
	}

	if len(commits) == 0 {
		return fmt.Errorf(`no commits ahead of %s

Make some commits first, then run vibe pr again.`, baseBranch)
	}

	ui.ShowInfo(fmt.Sprintf("Found %d commit(s) ahead of %s", len(commits), baseBranch))

	// Format commits for the prompt
	var commitLines []string
	for _, c := range commits {
		commitLines = append(commitLines, fmt.Sprintf("%s %s", c.Hash, c.Message))
	}
	commitsText := strings.Join(commitLines, "\n")

	// Get the diff from base branch
	diff, err := repo.GetDiffFromBase(baseBranch)
	if err != nil {
		return fmt.Errorf("failed to get diff: %w", err)
	}

	if diff == "" {
		return fmt.Errorf("no changes found compared to %s", baseBranch)
	}

	// Get remote URL and parse owner/repo
	remoteURL, err := repo.GetRemoteURL()
	if err != nil {
		return fmt.Errorf("failed to get remote URL: %w", err)
	}

	repoInfo, err := github.ParseRemoteURL(remoteURL)
	if err != nil {
		return fmt.Errorf("failed to parse GitHub remote: %w", err)
	}

	// Create OpenAI client and generate PR content
	llmClient, err := llm.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create AI client: %w", err)
	}

	prContent, err := llmClient.GeneratePRContent(commitsText, diff)
	if err != nil {
		return fmt.Errorf("failed to generate PR content: %w", err)
	}

	// Show the PR and get user confirmation
	result, err := ui.ConfirmPR(prContent.Title, prContent.Description)
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	switch result.Action {
	case ui.ActionCancel:
		ui.ShowInfo("PR creation cancelled.")
		return nil

	case ui.ActionAccept, ui.ActionEdit:
		// Check if we need to push
		needsPush, err := repo.NeedsPush()
		if err != nil {
			return fmt.Errorf("failed to check push status: %w", err)
		}

		if needsPush {
			ui.ShowInfo("Pushing branch to origin...")
			if err := repo.Push(); err != nil {
				return fmt.Errorf("failed to push branch: %w", err)
			}
		}

		// Create the PR
		ui.ShowInfo("Creating pull request...")

		ghClient, err := github.NewClient()
		if err != nil {
			return fmt.Errorf("failed to create GitHub client: %w", err)
		}

		prResult, err := ghClient.CreatePR(
			repoInfo.Owner,
			repoInfo.Name,
			baseBranch,
			currentBranch,
			result.Title,
			result.Description,
		)
		if err != nil {
			return fmt.Errorf("failed to create PR: %w", err)
		}

		ui.ShowSuccess(fmt.Sprintf("PR created: %s", prResult.URL))
		return nil

	default:
		return fmt.Errorf("unexpected action")
	}
}
