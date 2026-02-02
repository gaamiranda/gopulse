package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/user/vibe/internal/git"
	"github.com/user/vibe/internal/llm"
	"github.com/user/vibe/internal/ui"
)

var commitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate an AI commit message for staged changes",
	Long: `Analyzes your staged changes and generates a meaningful commit message using AI.

The command will:
1. Check for staged changes in your git repository
2. Generate a diff of the staged changes
3. Use OpenAI to generate a commit message
4. Show you the message for review
5. Allow you to accept, edit, or cancel
6. Create the commit if accepted

Requirements:
- Must be in a git repository
- Must have staged changes (git add)
- OPENAI_API_KEY environment variable must be set`,
	RunE: runCommit,
}

func init() {
	rootCmd.AddCommand(commitCmd)
}

func runCommit(cmd *cobra.Command, args []string) error {
	// Check for OpenAI API key
	if err := checkOpenAIKey(); err != nil {
		return err
	}

	// Open the git repository
	repo, err := git.OpenCurrent()
	if err != nil {
		return fmt.Errorf("not a git repository: %w", err)
	}

	// Check for staged changes
	hasStaged, err := repo.HasStagedChanges()
	if err != nil {
		return fmt.Errorf("failed to check staged changes: %w", err)
	}

	if !hasStaged {
		return fmt.Errorf(`no staged changes found

To stage changes, use:
  git add <file>       # Stage specific file
  git add .            # Stage all changes
  git add -p           # Stage interactively`)
	}

	// Get the diff
	ui.ShowInfo("Analyzing staged changes...")

	diff, err := repo.GetStagedDiff()
	if err != nil {
		return fmt.Errorf("failed to get staged diff: %w", err)
	}

	if diff == "" {
		return fmt.Errorf("no diff content found for staged changes")
	}

	// Create OpenAI client and generate commit message
	llmClient, err := llm.NewClient()
	if err != nil {
		return fmt.Errorf("failed to create AI client: %w", err)
	}

	message, err := llmClient.GenerateCommitMessage(diff)
	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Show the message and get user confirmation
	result, err := ui.ConfirmCommit(message)
	if err != nil {
		return fmt.Errorf("prompt failed: %w", err)
	}

	switch result.Action {
	case ui.ActionCancel:
		ui.ShowInfo("Commit cancelled.")
		return nil

	case ui.ActionAccept, ui.ActionEdit:
		// Create the commit
		hash, err := repo.Commit(result.Message)
		if err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}

		ui.ShowSuccess(fmt.Sprintf("Committed: %s", hash))
		fmt.Fprintf(os.Stdout, "\n  %s\n", result.Message)
		return nil

	default:
		return fmt.Errorf("unexpected action")
	}
}
