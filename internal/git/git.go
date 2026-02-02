package git

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/transport/http"
)

// Repository wraps go-git repository with helper methods
type Repository struct {
	repo *git.Repository
	path string
}

// Open opens a git repository at the given path
func Open(path string) (*Repository, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open repository: %w", err)
	}
	return &Repository{repo: repo, path: path}, nil
}

// OpenCurrent opens the git repository in the current directory
func OpenCurrent() (*Repository, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}
	return Open(cwd)
}

// HasStagedChanges checks if there are any staged changes
func (r *Repository) HasStagedChanges() (bool, error) {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return false, fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return false, fmt.Errorf("failed to get status: %w", err)
	}

	for _, s := range status {
		// Check if file is staged (in index)
		if s.Staging != git.Unmodified && s.Staging != git.Untracked {
			return true, nil
		}
	}
	return false, nil
}

// GetStagedDiff returns the diff of all staged changes
func (r *Repository) GetStagedDiff() (string, error) {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	status, err := worktree.Status()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}

	// Get HEAD commit tree (if exists)
	var headTree *object.Tree
	headRef, err := r.repo.Head()
	if err == nil {
		headCommit, err := r.repo.CommitObject(headRef.Hash())
		if err == nil {
			headTree, _ = headCommit.Tree()
		}
	}

	// Get the index (staged changes)
	idx, err := r.repo.Storer.Index()
	if err != nil {
		return "", fmt.Errorf("failed to get index: %w", err)
	}

	var diffBuilder strings.Builder

	for filePath, fileStatus := range status {
		// Only process staged files
		if fileStatus.Staging == git.Unmodified || fileStatus.Staging == git.Untracked {
			continue
		}

		diffBuilder.WriteString(fmt.Sprintf("diff --git a/%s b/%s\n", filePath, filePath))

		switch fileStatus.Staging {
		case git.Added:
			diffBuilder.WriteString("new file\n")
			// Read content from index
			for _, entry := range idx.Entries {
				if entry.Name == filePath {
					blob, err := r.repo.BlobObject(entry.Hash)
					if err == nil {
						reader, err := blob.Reader()
						if err == nil {
							content := make([]byte, blob.Size)
							_, _ = reader.Read(content)
							reader.Close()
							diffBuilder.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))
							for _, line := range strings.Split(string(content), "\n") {
								diffBuilder.WriteString(fmt.Sprintf("+%s\n", line))
							}
						}
					}
					break
				}
			}

		case git.Modified:
			// Get old content from HEAD
			if headTree != nil {
				file, err := headTree.File(filePath)
				if err == nil {
					oldContent, _ := file.Contents()
					diffBuilder.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
					diffBuilder.WriteString(fmt.Sprintf("+++ b/%s\n", filePath))

					// Get new content from index
					for _, entry := range idx.Entries {
						if entry.Name == filePath {
							blob, err := r.repo.BlobObject(entry.Hash)
							if err == nil {
								reader, err := blob.Reader()
								if err == nil {
									content := make([]byte, blob.Size)
									_, _ = reader.Read(content)
									reader.Close()
									newContent := string(content)

									// Simple line-by-line diff
									oldLines := strings.Split(oldContent, "\n")
									newLines := strings.Split(newContent, "\n")
									diffBuilder.WriteString(formatSimpleDiff(oldLines, newLines))
								}
							}
							break
						}
					}
				}
			}

		case git.Deleted:
			diffBuilder.WriteString("deleted file\n")
			if headTree != nil {
				file, err := headTree.File(filePath)
				if err == nil {
					content, _ := file.Contents()
					diffBuilder.WriteString(fmt.Sprintf("--- a/%s\n", filePath))
					for _, line := range strings.Split(content, "\n") {
						diffBuilder.WriteString(fmt.Sprintf("-%s\n", line))
					}
				}
			}
		}
		diffBuilder.WriteString("\n")
	}

	return diffBuilder.String(), nil
}

// formatSimpleDiff creates a simple unified diff format
func formatSimpleDiff(oldLines, newLines []string) string {
	var result strings.Builder

	// Simple approach: show removed lines then added lines
	oldSet := make(map[string]bool)
	for _, line := range oldLines {
		oldSet[line] = true
	}

	newSet := make(map[string]bool)
	for _, line := range newLines {
		newSet[line] = true
	}

	// Lines only in old (removed)
	for _, line := range oldLines {
		if !newSet[line] {
			result.WriteString(fmt.Sprintf("-%s\n", line))
		}
	}

	// Lines only in new (added)
	for _, line := range newLines {
		if !oldSet[line] {
			result.WriteString(fmt.Sprintf("+%s\n", line))
		}
	}

	return result.String()
}

// Commit creates a new commit with the given message
func (r *Repository) Commit(message string) (string, error) {
	worktree, err := r.repo.Worktree()
	if err != nil {
		return "", fmt.Errorf("failed to get worktree: %w", err)
	}

	// Get git config for author info
	cfg, err := r.repo.Config()
	if err != nil {
		return "", fmt.Errorf("failed to get config: %w", err)
	}

	authorName := cfg.User.Name
	authorEmail := cfg.User.Email

	// If no user configured in repo, try environment variables or use defaults
	if authorName == "" {
		authorName = os.Getenv("GIT_AUTHOR_NAME")
		if authorName == "" {
			authorName = "Vibe User"
		}
	}
	if authorEmail == "" {
		authorEmail = os.Getenv("GIT_AUTHOR_EMAIL")
		if authorEmail == "" {
			authorEmail = "vibe@local"
		}
	}

	hash, err := worktree.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  authorName,
			Email: authorEmail,
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", fmt.Errorf("failed to commit: %w", err)
	}

	return hash.String()[:7], nil
}

// GetCurrentBranch returns the name of the current branch
func (r *Repository) GetCurrentBranch() (string, error) {
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	if !head.Name().IsBranch() {
		return "", fmt.Errorf("HEAD is not on a branch (detached HEAD)")
	}

	return head.Name().Short(), nil
}

// GetDefaultBranch returns "main" or "master" depending on what exists
func (r *Repository) GetDefaultBranch() (string, error) {
	// Check for main first
	_, err := r.repo.Reference(plumbing.NewBranchReferenceName("main"), true)
	if err == nil {
		return "main", nil
	}

	// Fall back to master
	_, err = r.repo.Reference(plumbing.NewBranchReferenceName("master"), true)
	if err == nil {
		return "master", nil
	}

	// Check remote references
	remotes, err := r.repo.Remotes()
	if err == nil && len(remotes) > 0 {
		refs, err := r.repo.References()
		if err == nil {
			var defaultBranch string
			_ = refs.ForEach(func(ref *plumbing.Reference) error {
				name := ref.Name().String()
				if strings.Contains(name, "origin/main") {
					defaultBranch = "main"
					return fmt.Errorf("found")
				}
				if strings.Contains(name, "origin/master") {
					defaultBranch = "master"
				}
				return nil
			})
			if defaultBranch != "" {
				return defaultBranch, nil
			}
		}
	}

	return "", fmt.Errorf("could not determine default branch (no main or master found)")
}

// CommitInfo holds basic commit information
type CommitInfo struct {
	Hash    string
	Message string
}

// GetCommitsAhead returns commits on current branch that are ahead of base
func (r *Repository) GetCommitsAhead(base string) ([]CommitInfo, error) {
	// Get current branch HEAD
	head, err := r.repo.Head()
	if err != nil {
		return nil, fmt.Errorf("failed to get HEAD: %w", err)
	}

	// Get base branch reference
	baseRef, err := r.repo.Reference(plumbing.NewBranchReferenceName(base), true)
	if err != nil {
		// Try remote reference
		baseRef, err = r.repo.Reference(plumbing.NewRemoteReferenceName("origin", base), true)
		if err != nil {
			return nil, fmt.Errorf("failed to find base branch %s: %w", base, err)
		}
	}

	// Get commits from HEAD
	commitIter, err := r.repo.Log(&git.LogOptions{
		From: head.Hash(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get log: %w", err)
	}

	var commits []CommitInfo
	baseHash := baseRef.Hash()

	err = commitIter.ForEach(func(c *object.Commit) error {
		// Stop when we reach the base branch
		if c.Hash == baseHash {
			return fmt.Errorf("reached base")
		}

		commits = append(commits, CommitInfo{
			Hash:    c.Hash.String()[:7],
			Message: strings.Split(c.Message, "\n")[0], // First line only
		})
		return nil
	})

	// Ignore the "reached base" error
	if err != nil && err.Error() != "reached base" {
		// Only return error if it's not our stop condition
		if len(commits) == 0 {
			return nil, err
		}
	}

	return commits, nil
}

// GetRemoteURL returns the URL of the origin remote
func (r *Repository) GetRemoteURL() (string, error) {
	remote, err := r.repo.Remote("origin")
	if err != nil {
		return "", fmt.Errorf("failed to get origin remote: %w", err)
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", fmt.Errorf("no URLs configured for origin remote")
	}

	return urls[0], nil
}

// Push pushes the current branch to origin
func (r *Repository) Push() error {
	// Get GitHub token for authentication
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		return fmt.Errorf("GITHUB_TOKEN environment variable is not set")
	}

	// Get current branch name
	head, err := r.repo.Head()
	if err != nil {
		return fmt.Errorf("failed to get HEAD: %w", err)
	}

	branchName := head.Name().Short()
	refSpec := config.RefSpec(fmt.Sprintf("refs/heads/%s:refs/heads/%s", branchName, branchName))

	err = r.repo.Push(&git.PushOptions{
		RemoteName: "origin",
		Auth: &http.BasicAuth{
			Username: "x-access-token", // GitHub uses this for token auth
			Password: token,
		},
		RefSpecs: []config.RefSpec{refSpec},
	})

	if err == git.NoErrAlreadyUpToDate {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to push: %w", err)
	}

	return nil
}

// GetDiffFromBase returns the combined diff from base branch to current HEAD
func (r *Repository) GetDiffFromBase(base string) (string, error) {
	// Get current branch HEAD
	head, err := r.repo.Head()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}

	headCommit, err := r.repo.CommitObject(head.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD commit: %w", err)
	}

	// Get base branch reference
	baseRef, err := r.repo.Reference(plumbing.NewBranchReferenceName(base), true)
	if err != nil {
		// Try remote reference
		baseRef, err = r.repo.Reference(plumbing.NewRemoteReferenceName("origin", base), true)
		if err != nil {
			return "", fmt.Errorf("failed to find base branch %s: %w", base, err)
		}
	}

	baseCommit, err := r.repo.CommitObject(baseRef.Hash())
	if err != nil {
		return "", fmt.Errorf("failed to get base commit: %w", err)
	}

	// Get trees
	headTree, err := headCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD tree: %w", err)
	}

	baseTree, err := baseCommit.Tree()
	if err != nil {
		return "", fmt.Errorf("failed to get base tree: %w", err)
	}

	// Calculate diff
	changes, err := baseTree.Diff(headTree)
	if err != nil {
		return "", fmt.Errorf("failed to calculate diff: %w", err)
	}

	var diffBuilder strings.Builder
	for _, change := range changes {
		patch, err := change.Patch()
		if err != nil {
			continue
		}
		diffBuilder.WriteString(patch.String())
	}

	return diffBuilder.String(), nil
}

// NeedsPush checks if current branch has commits not yet pushed to origin
func (r *Repository) NeedsPush() (bool, error) {
	head, err := r.repo.Head()
	if err != nil {
		return false, fmt.Errorf("failed to get HEAD: %w", err)
	}

	branchName := head.Name().Short()

	// Get remote tracking branch
	remoteRef, err := r.repo.Reference(
		plumbing.NewRemoteReferenceName("origin", branchName),
		true,
	)
	if err != nil {
		// No remote tracking branch, needs push
		return true, nil
	}

	// If hashes differ, needs push
	return head.Hash() != remoteRef.Hash(), nil
}
