// Package github provides job implementations for the worker pool.
package github

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/valerius21/repokill/internal/worker"
)

// DeleteJob implements worker.Job for deleting a single repository.
type DeleteJob struct {
	Repo     Repo
	Client   *Client
	Result   DeleteResult
	Progress chan<- DeleteResult // Optional channel for progress updates
}

// Ensure DeleteJob implements worker.Job interface.
var _ worker.Job = (*DeleteJob)(nil)

// Execute performs the repository deletion with retry logic.
func (j *DeleteJob) Execute(ctx context.Context) error {
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

		output, err := j.Client.executor.Execute(ctx, "gh", "repo", "delete", j.Repo.NameWithOwner, "--yes")

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
					break
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

	j.Result = DeleteResult{
		Repo:     j.Repo,
		Success:  success,
		Error:    lastErr,
		Duration: time.Since(start),
	}

	// Send progress update if channel is provided
	if j.Progress != nil {
		select {
		case <-ctx.Done():
		case j.Progress <- j.Result:
		}
	}

	// Return nil even on error - errors are tracked in Result, not as job failures
	// This allows other jobs to continue processing
	return nil
}

// Name returns a human-readable identifier for the job.
func (j *DeleteJob) Name() string {
	return fmt.Sprintf("delete-%s", j.Repo.Name)
}

// ArchiveJob implements worker.Job for archiving a single repository.
type ArchiveJob struct {
	Repo     Repo
	Client   *Client
	Result   DeleteResult
	Progress chan<- DeleteResult // Optional channel for progress updates
}

// Ensure ArchiveJob implements worker.Job interface.
var _ worker.Job = (*ArchiveJob)(nil)

// Execute performs the repository archival with retry logic.
func (j *ArchiveJob) Execute(ctx context.Context) error {
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

		output, err := j.Client.executor.Execute(ctx, "gh", "repo", "archive", j.Repo.NameWithOwner, "--yes")

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
					break
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
			lastErr = fmt.Errorf("repo not found: %s", errMsg)
		} else if strings.Contains(errMsg, "already archived") {
			success = true // Already archived is effectively a success
			lastErr = nil
		} else {
			lastErr = fmt.Errorf("gh repo archive failed: %v (output: %s)", err, errMsg)
		}
		break
	}

	j.Result = DeleteResult{
		Repo:     j.Repo,
		Success:  success,
		Error:    lastErr,
		Duration: time.Since(start),
	}

	// Send progress update if channel is provided
	if j.Progress != nil {
		select {
		case <-ctx.Done():
		case j.Progress <- j.Result:
		}
	}

	// Return nil even on error - errors are tracked in Result, not as job failures
	// This allows other jobs to continue processing
	return nil
}

// Name returns a human-readable identifier for the job.
func (j *ArchiveJob) Name() string {
	return fmt.Sprintf("archive-%s", j.Repo.Name)
}

// VisibilityJob implements worker.Job for changing a repository's visibility.
type VisibilityJob struct {
	Repo        Repo
	Client      *Client
	MakePrivate bool
	Result      ChangeVisibilityResult
	Progress    chan<- ChangeVisibilityResult // Optional channel for progress updates
}

// Ensure VisibilityJob implements worker.Job interface.
var _ worker.Job = (*VisibilityJob)(nil)

// Execute performs the repository visibility change with retry logic.
func (j *VisibilityJob) Execute(ctx context.Context) error {
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

		privateFlag := "true"
		if !j.MakePrivate {
			privateFlag = "false"
		}

		output, err := j.Client.executor.Execute(ctx, "gh", "api",
			"repos/"+j.Repo.NameWithOwner,
			"-X", "PATCH",
			"-f", "private="+privateFlag)

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
					break
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
			lastErr = fmt.Errorf("repo not found: %s", errMsg)
		} else {
			lastErr = fmt.Errorf("gh api failed: %v (output: %s)", err, errMsg)
		}
		break
	}

	j.Result = ChangeVisibilityResult{
		Repo:     j.Repo,
		Success:  success,
		Error:    lastErr,
		Duration: time.Since(start),
		Private:  j.MakePrivate,
	}

	// Send progress update if channel is provided
	if j.Progress != nil {
		select {
		case <-ctx.Done():
		case j.Progress <- j.Result:
		}
	}

	// Return nil even on error - errors are tracked in Result, not as job failures
	// This allows other jobs to continue processing
	return nil
}

// Name returns a human-readable identifier for the job.
func (j *VisibilityJob) Name() string {
	visibility := "private"
	if !j.MakePrivate {
		visibility = "public"
	}
	return fmt.Sprintf("visibility-%s-%s", visibility, j.Repo.Name)
}

// BackupJob implements worker.Job for backing up a single repository.
type BackupJob struct {
	Repo      Repo
	Client    *Client
	BackupDir string
	Mode      string // "zip" or "clone"
	Result    BackupResult
	Progress  chan<- BackupResult // Optional channel for progress updates
	Ref       string              // Git ref to download (default: default branch)
}

// Ensure BackupJob implements worker.Job interface.
var _ worker.Job = (*BackupJob)(nil)

// Execute performs the repository backup.
// For ZIP mode: downloads via gh api repos/{owner}/{repo}/zipball/{ref}
// For clone mode: runs git clone --depth=1
func (j *BackupJob) Execute(ctx context.Context) error {
	start := time.Now()

	// Create safe directory name from owner/repo
	safeName := strings.ReplaceAll(j.Repo.NameWithOwner, "/", "-")
	targetPath := filepath.Join(j.BackupDir, safeName)

	// Check if repo already exists - skip if present
	if _, err := os.Stat(targetPath); err == nil {
		j.Result = BackupResult{
			Repo:     j.Repo,
			Success:  true,
			Error:    nil,
			Duration: time.Since(start),
			Skipped:  true,
		}
		if j.Progress != nil {
			select {
			case <-ctx.Done():
			case j.Progress <- j.Result:
			}
		}
		return nil
	}

	// Ensure backup directory exists
	if err := os.MkdirAll(j.BackupDir, 0755); err != nil {
		j.Result = BackupResult{
			Repo:     j.Repo,
			Success:  false,
			Error:    fmt.Errorf("failed to create backup directory: %w", err),
			Duration: time.Since(start),
			Skipped:  false,
		}
		if j.Progress != nil {
			select {
			case <-ctx.Done():
			case j.Progress <- j.Result:
			}
		}
		return nil
	}

	var lastErr error

	// Determine ref to use
	ref := j.Ref
	if ref == "" {
		ref = "HEAD" // Use default branch
	}

	if j.Mode == "zip" {
		lastErr = j.backupZip(ctx, targetPath, ref)
	} else {
		lastErr = j.backupClone(ctx, targetPath)
	}

	j.Result = BackupResult{
		Repo:     j.Repo,
		Success:  lastErr == nil,
		Error:    lastErr,
		Duration: time.Since(start),
		Skipped:  false,
	}

	if j.Progress != nil {
		select {
		case <-ctx.Done():
		case j.Progress <- j.Result:
		}
	}

	return nil
}

// backupZip downloads the repository as a ZIP file using gh CLI.
func (j *BackupJob) backupZip(ctx context.Context, targetPath, ref string) error {
	// Use gh api to download ZIP
	zipPath := targetPath + ".zip"

	output, err := j.Client.executor.Execute(ctx, "gh", "api",
		"repos/"+j.Repo.NameWithOwner+"/zipball/"+ref,
		"--output", zipPath)

	if err != nil {
		return fmt.Errorf("gh api zipball failed: %v (output: %s)", err, string(output))
	}

	return nil
}

// backupClone performs a shallow git clone of the repository.
func (j *BackupJob) backupClone(ctx context.Context, targetPath string) error {
	// Construct clone URL
	cloneURL := fmt.Sprintf("https://github.com/%s.git", j.Repo.NameWithOwner)

	// Run git clone --depth=1
	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", cloneURL, targetPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("git clone failed: %v (output: %s)", err, string(output))
	}

	return nil
}

// Name returns a human-readable identifier for the job.
func (j *BackupJob) Name() string {
	return fmt.Sprintf("backup-%s", j.Repo.Name)
}
