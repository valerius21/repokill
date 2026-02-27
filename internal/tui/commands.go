// Package tui provides Bubble Tea commands for the terminal user interface.
package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/valerius21/repokill/internal/github"
	"github.com/valerius21/repokill/internal/worker"
)

type reposLoadedMsg []github.Repo
type reposLoadErrorMsg error

type repoDeletedMsg struct {
	repo    github.Repo
	result  github.DeleteResult
	current int
	total   int
}

type allDeletesDoneMsg struct {
	results []github.DeleteResult
}

type allArchivesDoneMsg struct {
	results []github.DeleteResult
}

func fetchRepos(client *github.Client) tea.Cmd {
	return func() tea.Msg {
		repos, err := client.ListRepos(context.Background())
		if err != nil {
			return reposLoadErrorMsg(err)
		}
		return reposLoadedMsg(repos)
	}
}

// deleteReposCmd processes repository deletions concurrently using the worker pool.
// It sends progress updates as each deletion completes.
func deleteReposCmd(client *github.Client, repos []github.Repo) tea.Cmd {
	return func() tea.Msg {
		if len(repos) == 0 {
			return allDeletesDoneMsg{results: nil}
		}

		ctx := context.Background()
		pool := worker.NewWorkerPool(ctx, worker.WithConcurrency(worker.DefaultConcurrency))

		// Channel for progress updates
		progress := make(chan github.DeleteResult, len(repos))

		// Create and submit jobs
		jobs := make([]worker.Job, len(repos))
		for i, repo := range repos {
			jobs[i] = &github.DeleteJob{
				Repo:     repo,
				Client:   client,
				Progress: progress,
			}
		}

		// Start a goroutine to collect results
		results := make([]github.DeleteResult, 0, len(repos))
		done := make(chan struct{})
		go func() {
			for res := range progress {
				results = append(results, res)
			}
			close(done)
		}()

		// Process all jobs
		pool.ProcessBatch(ctx, jobs)

		// Close progress channel and wait for collector
		close(progress)
		<-done

		return allDeletesDoneMsg{results: results}
	}
}

// archiveReposCmd processes repository archivals concurrently using the worker pool.
// It sends progress updates as each archival completes.
func archiveReposCmd(client *github.Client, repos []github.Repo) tea.Cmd {
	return func() tea.Msg {
		if len(repos) == 0 {
			return allArchivesDoneMsg{results: nil}
		}

		ctx := context.Background()
		pool := worker.NewWorkerPool(ctx, worker.WithConcurrency(worker.DefaultConcurrency))

		// Channel for progress updates
		progress := make(chan github.DeleteResult, len(repos))

		// Create and submit jobs
		jobs := make([]worker.Job, len(repos))
		for i, repo := range repos {
			jobs[i] = &github.ArchiveJob{
				Repo:     repo,
				Client:   client,
				Progress: progress,
			}
		}

		// Start a goroutine to collect results
		results := make([]github.DeleteResult, 0, len(repos))
		done := make(chan struct{})
		go func() {
			for res := range progress {
				results = append(results, res)
			}
			close(done)
		}()

		// Process all jobs
		pool.ProcessBatch(ctx, jobs)

		// Close progress channel and wait for collector
		close(progress)
		<-done

		return allArchivesDoneMsg{results: results}
	}
}

type visibilityChangedMsg struct {
	repo   github.Repo
	result github.ChangeVisibilityResult
}

type allVisibilityDoneMsg struct {
	results []github.ChangeVisibilityResult
	private bool // true if made private, false if made public
}

// changeVisibilityCmd processes repository visibility changes concurrently using the worker pool.
// It sends progress updates as each visibility change completes.
func changeVisibilityCmd(client *github.Client, repos []github.Repo, makePrivate bool) tea.Cmd {
	return func() tea.Msg {
		if len(repos) == 0 {
			return allVisibilityDoneMsg{results: nil, private: makePrivate}
		}

		ctx := context.Background()
		pool := worker.NewWorkerPool(ctx, worker.WithConcurrency(worker.DefaultConcurrency))

		// Channel for progress updates
		progress := make(chan github.ChangeVisibilityResult, len(repos))

		// Create and submit jobs
		jobs := make([]worker.Job, len(repos))
		for i, repo := range repos {
			jobs[i] = &github.VisibilityJob{
				Repo:        repo,
				Client:      client,
				MakePrivate: makePrivate,
				Progress:    progress,
			}
		}

		// Start a goroutine to collect results
		results := make([]github.ChangeVisibilityResult, 0, len(repos))
		done := make(chan struct{})
		go func() {
			for res := range progress {
				results = append(results, res)
			}
			close(done)
		}()

		// Process all jobs
		pool.ProcessBatch(ctx, jobs)

		// Close progress channel and wait for collector
		close(progress)
		<-done

		return allVisibilityDoneMsg{results: results, private: makePrivate}
	}
}

type backupProgressMsg struct {
	repo    github.Repo
	result  github.BackupResult
	current int
	total   int
}

type allBackupsDoneMsg struct {
	results []github.BackupResult
}

// backupReposCmd processes repository backups concurrently using the worker pool.
// It sends progress updates as each backup completes.
func backupReposCmd(client *github.Client, repos []github.Repo, backupDir string, mode string) tea.Cmd {
	return func() tea.Msg {
		if len(repos) == 0 {
			return allBackupsDoneMsg{results: nil}
		}

		ctx := context.Background()
		pool := worker.NewWorkerPool(ctx, worker.WithConcurrency(worker.DefaultConcurrency))

		// Channel for progress updates
		progress := make(chan github.BackupResult, len(repos))

		// Create and submit jobs
		jobs := make([]worker.Job, len(repos))
		for i, repo := range repos {
			jobs[i] = &github.BackupJob{
				Repo:      repo,
				Client:    client,
				BackupDir: backupDir,
				Mode:      mode,
				Progress:  progress,
			}
		}

		// Start a goroutine to collect results
		results := make([]github.BackupResult, 0, len(repos))
		done := make(chan struct{})
		go func() {
			for res := range progress {
				results = append(results, res)
			}
			close(done)
		}()

		// Process all jobs
		pool.ProcessBatch(ctx, jobs)

		// Close progress channel and wait for collector
		close(progress)
		<-done

		return allBackupsDoneMsg{results: results}
	}
}
