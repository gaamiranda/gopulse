# Vibe CLI

AI-powered Git CLI for generating commit messages and PR descriptions using OpenAI.

## Overview

Vibe streamlines your git workflow by analyzing your changes and suggesting appropriate commit messages or PR descriptions using AI.

## Features

- **AI Commit Messages**: Generate meaningful commit messages from your staged changes
- **AI PR Descriptions**: Create GitHub PRs with AI-generated titles and descriptions
- **Interactive Review**: Always review and edit before committing or creating PRs
- **Pure Go**: No external git binary required (uses go-git)
- **Auto .env Loading**: Automatically loads environment variables from `.env` file

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/user/vibe.git
cd vibe

# Build and install (using Makefile)
make install

# Or build and create symlink (recommended for development)
make link
```

### Manual Build

```bash
# Build the binary
go build -o vibe .

# Move to PATH
sudo mv vibe /usr/local/bin/
```

## Configuration

Vibe requires the following environment variables:

```bash
# Required for all commands
export OPENAI_API_KEY="your-openai-api-key"

# Required for PR command
export GITHUB_TOKEN="your-github-token"
```

### Using a .env File

Vibe automatically loads a `.env` file from the current directory if present:

```bash
# Create .env file in your project root
echo 'OPENAI_API_KEY=your-openai-api-key' > .env
echo 'GITHUB_TOKEN=your-github-token' >> .env
```

> **Note**: Never commit your `.env` file to git. It's already in the default `.gitignore`.

### Getting API Keys

- **OpenAI API Key**: Get yours at [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
- **GitHub Token**: Create at [github.com/settings/tokens](https://github.com/settings/tokens) (requires `repo` scope)

## Usage

### Commit with AI Message

```bash
# Stage your changes first
git add .

# Generate and apply AI commit message
vibe commit
```

**Example workflow:**
```
$ git add src/feature.go
$ vibe commit

Analyzing staged changes...

Generated commit message:
--------------------------------------------------
Add user authentication middleware with JWT validation
--------------------------------------------------

? What would you like to do? [Accept / Edit / Cancel]
> Accept

Committed: a1b2c3d

  Add user authentication middleware with JWT validation
```

### Create PR with AI Description

```bash
# On a feature branch with commits
vibe pr
```

**Example workflow:**
```
$ vibe pr

Analyzing branch 'feature/auth' against 'main'...
Found 3 commit(s) ahead of main

Generated PR:
--------------------------------------------------
Title: Add user authentication system

Description:
This PR introduces JWT-based authentication for the API.

Key changes:
- Add auth middleware for protected routes
- Implement login and logout endpoints
- Add user session management
--------------------------------------------------

? What would you like to do? [Accept / Edit / Cancel]
> Accept

Pushing branch to origin...
Creating pull request...

PR created: https://github.com/user/repo/pull/42
```

## Commands

| Command | Description |
|---------|-------------|
| `vibe commit` | Generate AI commit message for staged changes |
| `vibe pr` | Create GitHub PR with AI-generated title and description |
| `vibe version` | Show version information |
| `vibe --help` | Show help information |

## Error Handling

Vibe provides helpful error messages for common issues:

- **No staged changes**: Reminds you to use `git add` first
- **Not a git repository**: Tells you to navigate to a git repo
- **Invalid API key**: Shows how to fix your credentials
- **Rate limits**: Explains what to do when limits are hit
- **Network errors**: Suggests checking your connection

## Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Git Operations**: [go-git](https://github.com/go-git/go-git)
- **GitHub API**: [go-github](https://github.com/google/go-github)
- **LLM**: [OpenAI API](https://github.com/sashabaranov/go-openai)
- **Interactive UI**: [huh](https://github.com/charmbracelet/huh)
- **Env Loading**: [godotenv](https://github.com/joho/godotenv)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

[MIT](LICENSE)
