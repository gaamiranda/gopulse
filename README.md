# Vibe CLI

AI-powered Git CLI for generating commit messages and PR descriptions using OpenAI.

## Overview

Vibe streamlines your git workflow by analyzing your changes and suggesting appropriate commit messages or PR descriptions using AI.

## Features

- **AI Commit Messages**: Generate meaningful commit messages from your staged changes
- **AI PR Descriptions**: Create GitHub PRs with AI-generated titles and descriptions
- **Interactive Review**: Always review and edit before committing or creating PRs
- **Pure Go**: No external git binary required (uses go-git)

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/user/vibe.git
cd vibe

# Build
go build -o vibe .

# Move to PATH (optional)
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

### Create PR with AI Description

```bash
# On a feature branch with commits
vibe pr
```

## Commands

| Command | Description |
|---------|-------------|
| `vibe commit` | Generate AI commit message for staged changes |
| `vibe pr` | Create GitHub PR with AI-generated title and description |
| `vibe --help` | Show help information |

## Development Status

This project is under active development.

- [x] Phase 1: Project Setup
- [x] Phase 2: Core Infrastructure
- [x] Phase 3: Commands
- [ ] Phase 4: Polish

## Tech Stack

- **Language**: Go 1.21+
- **CLI Framework**: [Cobra](https://github.com/spf13/cobra)
- **Git Operations**: [go-git](https://github.com/go-git/go-git)
- **GitHub API**: [go-github](https://github.com/google/go-github)
- **LLM**: [OpenAI API](https://github.com/sashabaranov/go-openai)
- **Interactive UI**: [huh](https://github.com/charmbracelet/huh)

## License

[MIT](LICENSE)
