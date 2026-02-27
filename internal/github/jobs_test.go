package github

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/valerius21/repokill/internal/worker"
)

func TestDeleteJob_Execute(t *testing.T) {
	tests := []struct {
		name       string
		repo       Repo
		mockOutput []byte
		mockError  error
		wantErr    bool
		wantResult bool
	}{
		{
			name:       "successful deletion",
			repo:       Repo{Name: "test-repo", NameWithOwner: "owner/test-repo"},
			mockOutput: []byte(""),
			mockError:  nil,
			wantErr:    false,
			wantResult: true,
		},
		{
			name:       "deletion failure - not found",
			repo:       Repo{Name: "missing-repo", NameWithOwner: "owner/missing-repo"},
			mockOutput: []byte("HTTP 404: Not Found"),
			mockError:  errors.New("exit status 1"),
			wantErr:    false, // Job doesn't return error, stores in Result
			wantResult: false,
		},
		{
			name:       "already deleted",
			repo:       Repo{Name: "deleted-repo", NameWithOwner: "owner/deleted-repo"},
			mockOutput: []byte("HTTP 404: Not Found"),
			mockError:  errors.New("exit status 1"),
			wantErr:    false,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockExecutor{
				executeFunc: func(name string, args ...string) ([]byte, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			client := NewClient("owner", executor)

			job := &DeleteJob{
				Repo:   tt.repo,
				Client: client,
			}

			err := job.Execute(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteJob.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if job.Result.Success != tt.wantResult {
				t.Errorf("DeleteJob.Result.Success = %v, want %v", job.Result.Success, tt.wantResult)
			}

			if job.Result.Repo.Name != tt.repo.Name {
				t.Errorf("DeleteJob.Result.Repo.Name = %v, want %v", job.Result.Repo.Name, tt.repo.Name)
			}
		})
	}
}

func TestDeleteJob_ProgressChannel(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(name string, args ...string) ([]byte, error) {
			return []byte(""), nil
		},
	}
	client := NewClient("owner", executor)

	progress := make(chan DeleteResult, 1)
	job := &DeleteJob{
		Repo:     Repo{Name: "test-repo", NameWithOwner: "owner/test-repo"},
		Client:   client,
		Progress: progress,
	}

	err := job.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case result := <-progress:
		if result.Repo.Name != "test-repo" {
			t.Errorf("expected repo name 'test-repo', got %v", result.Repo.Name)
		}
		if !result.Success {
			t.Error("expected success to be true")
		}
	case <-time.After(1 * time.Second):
		t.Error("expected progress update but got none")
	}
}

func TestDeleteJob_Name(t *testing.T) {
	job := &DeleteJob{
		Repo: Repo{Name: "my-repo", NameWithOwner: "owner/my-repo"},
	}

	if name := job.Name(); name != "delete-my-repo" {
		t.Errorf("DeleteJob.Name() = %v, want %v", name, "delete-my-repo")
	}
}

func TestArchiveJob_Execute(t *testing.T) {
	tests := []struct {
		name       string
		repo       Repo
		mockOutput []byte
		mockError  error
		wantErr    bool
		wantResult bool
	}{
		{
			name:       "successful archival",
			repo:       Repo{Name: "test-repo", NameWithOwner: "owner/test-repo"},
			mockOutput: []byte(""),
			mockError:  nil,
			wantErr:    false,
			wantResult: true,
		},
		{
			name:       "already archived is success",
			repo:       Repo{Name: "archived-repo", NameWithOwner: "owner/archived-repo"},
			mockOutput: []byte("already archived"),
			mockError:  errors.New("exit status 1"),
			wantErr:    false,
			wantResult: true, // Already archived counts as success
		},
		{
			name:       "archival failure - not found",
			repo:       Repo{Name: "missing-repo", NameWithOwner: "owner/missing-repo"},
			mockOutput: []byte("HTTP 404: Not Found"),
			mockError:  errors.New("exit status 1"),
			wantErr:    false,
			wantResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			executor := &mockExecutor{
				executeFunc: func(name string, args ...string) ([]byte, error) {
					return tt.mockOutput, tt.mockError
				},
			}
			client := NewClient("owner", executor)

			job := &ArchiveJob{
				Repo:   tt.repo,
				Client: client,
			}

			err := job.Execute(context.Background())

			if (err != nil) != tt.wantErr {
				t.Errorf("ArchiveJob.Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if job.Result.Success != tt.wantResult {
				t.Errorf("ArchiveJob.Result.Success = %v, want %v", job.Result.Success, tt.wantResult)
			}
		})
	}
}

func TestArchiveJob_ProgressChannel(t *testing.T) {
	executor := &mockExecutor{
		executeFunc: func(name string, args ...string) ([]byte, error) {
			return []byte(""), nil
		},
	}
	client := NewClient("owner", executor)

	progress := make(chan DeleteResult, 1)
	job := &ArchiveJob{
		Repo:     Repo{Name: "test-repo", NameWithOwner: "owner/test-repo"},
		Client:   client,
		Progress: progress,
	}

	err := job.Execute(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case result := <-progress:
		if result.Repo.Name != "test-repo" {
			t.Errorf("expected repo name 'test-repo', got %v", result.Repo.Name)
		}
		if !result.Success {
			t.Error("expected success to be true")
		}
	case <-time.After(1 * time.Second):
		t.Error("expected progress update but got none")
	}
}

func TestArchiveJob_Name(t *testing.T) {
	job := &ArchiveJob{
		Repo: Repo{Name: "my-repo", NameWithOwner: "owner/my-repo"},
	}

	if name := job.Name(); name != "archive-my-repo" {
		t.Errorf("ArchiveJob.Name() = %v, want %v", name, "archive-my-repo")
	}
}

func TestDeleteJob_ImplementsJobInterface(t *testing.T) {
	// Compile-time check that DeleteJob implements worker.Job
	var _ worker.Job = (*DeleteJob)(nil)
}

func TestArchiveJob_ImplementsJobInterface(t *testing.T) {
	// Compile-time check that ArchiveJob implements worker.Job
	var _ worker.Job = (*ArchiveJob)(nil)
}
