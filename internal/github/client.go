package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"sort"
	"strings"
	"time"
)

// CommandExecutor abstracts command execution for testability.
type CommandExecutor interface {
	Execute(ctx context.Context, name string, args ...string) ([]byte, error)
	LookPath(name string) (string, error)
}

// execCommandExecutor is the default implementation using exec.CommandContext.
type execCommandExecutor struct{}

func (e *execCommandExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.CombinedOutput()
}

func (e *execCommandExecutor) LookPath(name string) (string, error) {
	return exec.LookPath(name)
}

// Sentinel errors for common failure cases.
var (
	ErrGhNotInstalled   = errors.New("gh CLI not found. Install from https://cli.github.com")
	ErrNotAuthenticated = errors.New("Not logged in. Run `gh auth login` first.")
	ErrMissingScope     = errors.New("Missing delete_repo scope. Run `gh auth refresh -s delete_repo`.")
)

// Client handles interactions with GitHub via the gh CLI.
type Client struct {
	Owner    string
	executor CommandExecutor
}

// NewClient creates a new GitHub client for the specified owner.
// If owner is empty, operations target the authenticated user.
// If executor is nil, the default exec-based executor is used.
func NewClient(owner string, executor CommandExecutor) *Client {
	if executor == nil {
		executor = &execCommandExecutor{}
	}
	return &Client{
		Owner:    owner,
		executor: executor,
	}
}

// ListRepos fetches repositories for the configured owner using the gh CLI.
// If Owner is empty, it lists repositories for the authenticated user.
// Results are sorted by PushedAt in ascending order (oldest first).
func (c *Client) ListRepos(ctx context.Context) ([]Repo, error) {
	args := []string{
		"repo", "list",
		"--json", "name,nameWithOwner,description,pushedAt,visibility,isArchived,isFork,stargazerCount,forkCount",
		"--limit", "1000",
	}

	if c.Owner != "" {
		args = append(args, c.Owner)
	}

	output, err := c.executor.Execute(ctx, "gh", args...)
	if err != nil {
		return nil, fmt.Errorf("gh repo list failed: %w", err)
	}

	var repos []Repo
	if err := json.Unmarshal(output, &repos); err != nil {
		return nil, fmt.Errorf("failed to parse gh output: %w", err)
	}

	sort.Slice(repos, func(i, j int) bool {
		return repos[i].PushedAt.Before(repos[j].PushedAt)
	})

	return repos, nil
}

// CheckAuth verifies that the gh CLI is installed and the user is authenticated.
func (c *Client) CheckAuth(ctx context.Context) error {
	if _, err := c.executor.LookPath("gh"); err != nil {
		return ErrGhNotInstalled
	}

	_, err := c.executor.Execute(ctx, "gh", "auth", "status")
	if err != nil {
		return ErrNotAuthenticated
	}

	return nil
}

// DeleteProgressFn is a callback for tracking deletion progress.
type DeleteProgressFn func(repo Repo, result DeleteResult, current int, total int)

// DeleteRepos deletes the specified repositories sequentially with a 1-second throttle
// between requests to comply with GitHub's rate limits. It includes automatic retry
// logic with exponential backoff for rate-limited requests (HTTP 429).
func (c *Client) DeleteRepos(ctx context.Context, repos []Repo, onProgress DeleteProgressFn) []DeleteResult {
	results := make([]DeleteResult, 0, len(repos))
	total := len(repos)

	for i, repo := range repos {
		if ctx.Err() != nil {
			break
		}

		if i > 0 {
			time.Sleep(1 * time.Second)
		}

		start := time.Now()
		var lastErr error
		var success bool
		backoff := 2 * time.Second
		maxRetries := 5

		for retry := 0; retry <= maxRetries; retry++ {
			if ctx.Err() != nil {
				lastErr = ctx.Err()
				break
			}

			output, err := c.executor.Execute(ctx, "gh", "repo", "delete", repo.NameWithOwner, "--yes")

			if err == nil {
				success = true
				lastErr = nil
				break
			}

			errMsg := string(output)
			if strings.Contains(errMsg, "HTTP 429") {
				if retry < maxRetries {
					select {
					case <-ctx.Done():
						lastErr = ctx.Err()
						goto nextRepo
					case <-time.After(backoff):
						backoff *= 2
						if backoff > 32*time.Second {
							backoff = 32 * time.Second
						}
						continue
					}
				}
				lastErr = fmt.Errorf("rate limited after %d retries: %s", maxRetries, errMsg)
				break
			}

			if strings.Contains(errMsg, "could not resolve") {
				lastErr = fmt.Errorf("repo renamed or transferred: %s", errMsg)
			} else if strings.Contains(errMsg, "HTTP 403") {
				lastErr = fmt.Errorf("permission denied: %s", errMsg)
			} else if strings.Contains(errMsg, "HTTP 404") {
				lastErr = fmt.Errorf("repo not found (already deleted): %s", errMsg)
			} else {
				lastErr = fmt.Errorf("gh repo delete failed: %v (output: %s)", err, errMsg)
			}
			break
		}

	nextRepo:
		result := DeleteResult{
			Repo:     repo,
			Success:  success,
			Error:    lastErr,
			Duration: time.Since(start),
		}
		results = append(results, result)

		if onProgress != nil {
			onProgress(repo, result, i+1, total)
		}
	}

	return results
}