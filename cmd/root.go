package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vibe",
	Short: "AI-powered Git CLI for commits and PRs",
	Long: `Vibe is a CLI tool that uses AI to generate commit messages and PR descriptions.

It streamlines your git workflow by analyzing your changes and suggesting
appropriate commit messages or PR descriptions using OpenAI.

Commands:
  vibe commit  - Generate an AI commit message for staged changes
  vibe pr      - Create a GitHub PR with AI-generated title and description

Environment Variables:
  OPENAI_API_KEY  - Your OpenAI API key (required)
  GITHUB_TOKEN    - Your GitHub personal access token (required for PR command)`,
}

// Execute runs the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Disable the default completion command
	rootCmd.CompletionOptions.DisableDefaultCmd = true
}

// checkOpenAIKey validates that OPENAI_API_KEY is set
func checkOpenAIKey() error {
	if os.Getenv("OPENAI_API_KEY") == "" {
		return fmt.Errorf(`OPENAI_API_KEY environment variable is not set.

To fix this:
  export OPENAI_API_KEY="your-api-key"

Get your API key at: https://platform.openai.com/api-keys`)
	}
	return nil
}

// checkGitHubToken validates that GITHUB_TOKEN is set
func checkGitHubToken() error {
	if os.Getenv("GITHUB_TOKEN") == "" {
		return fmt.Errorf(`GITHUB_TOKEN environment variable is not set.

To fix this:
  export GITHUB_TOKEN="your-token"

Create a token at: https://github.com/settings/tokens
Required scope: repo`)
	}
	return nil
}
