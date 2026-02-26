package github

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

// mockExecutor implements CommandExecutor for testing.
type mockExecutor struct {
	lookPathErr error
	lookPathRes string

	// For Execute calls, we can match on args and return specific results
	executeFunc func(name string, args ...string) ([]byte, error)
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) ([]byte, error) {
	if m.executeFunc != nil {
		return m.executeFunc(name, args...)
	}
	return nil, nil
}

func (m *mockExecutor) LookPath(name string) (string, error) {
	return m.lookPathRes, m.lookPathErr
}

// TestListRepos tests the ListRepos method
func TestListRepos(t *testing.T) {
	t.Run("happy path with 3 repos", func(t *testing.T) {
		mockResponse := `[
			{"name":"repo1","nameWithOwner":"owner/repo1","description":"First repo","pushedAt":"2024-01-01T00:00:00Z","visibility":"public","isArchived":false,"isFork":false,"stargazerCount":10,"forkCount":2},
			{"name":"repo2","nameWithOwner":"owner/repo2","description":"Second repo","pushedAt":"2024-02-01T00:00:00Z","visibility":"private","isArchived":true,"isFork":false,"stargazerCount":5,"forkCount":0},
			{"name":"repo3","nameWithOwner":"owner/repo3","description":"Third repo","pushedAt":"2024-03-01T00:00:00Z","visibility":"public","isArchived":false,"isFork":true,"stargazerCount":0,"forkCount":0}
		]`
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				if name != "gh" {
					t.Errorf("expected gh, got %s", name)
				}
				if len(args) < 3 || args[0] != "repo" || args[1] != "list" {
					t.Errorf("expected repo list args, got %v", args)
				}
				return []byte(mockResponse), nil
			},
		}
		client := NewClient("", mock)
		repos, err := client.ListRepos(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repos) != 3 {
			t.Fatalf("expected 3 repos, got %d", len(repos))
		}
		// Verify sorting (oldest first by PushedAt)
		if repos[0].Name != "repo1" {
			t.Errorf("expected repo1 first, got %s", repos[0].Name)
		}
		if repos[1].Name != "repo2" {
			t.Errorf("expected repo2 second, got %s", repos[1].Name)
		}
		if repos[2].Name != "repo3" {
			t.Errorf("expected repo3 third, got %s", repos[2].Name)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("[]"), nil
			},
		}
		client := NewClient("", mock)
		repos, err := client.ListRepos(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(repos) != 0 {
			t.Errorf("expected 0 repos, got %d", len(repos))
		}
	})

	t.Run("execution error", func(t *testing.T) {
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				return nil, errors.New("command failed")
			},
		}
		client := NewClient("", mock)
		_, err := client.ListRepos(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "gh repo list failed") {
			t.Errorf("expected gh repo list error, got: %v", err)
		}
	})

	t.Run("JSON parse error", func(t *testing.T) {
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				return []byte("invalid json"), nil
			},
		}
		client := NewClient("", mock)
		_, err := client.ListRepos(context.Background())
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !strings.Contains(err.Error(), "failed to parse") {
			t.Errorf("expected parse error, got: %v", err)
		}
	})

	t.Run("with owner parameter", func(t *testing.T) {
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				// Check that owner is appended
				found := false
				for _, arg := range args {
					if arg == "myorg" {
						found = true
						break
					}
				}
				if !found {
					t.Error("expected owner 'myorg' in args")
				}
				return []byte("[]"), nil
			},
		}
		client := NewClient("myorg", mock)
		_, err := client.ListRepos(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

// TestDeleteRepos tests the DeleteRepos method
func TestDeleteRepos(t *testing.T) {
	t.Run("happy path with 2 repos", func(t *testing.T) {
		repos := []Repo{
			{Name: "repo1", NameWithOwner: "owner/repo1"},
			{Name: "repo2", NameWithOwner: "owner/repo2"},
		}
		deleteCount := 0
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				if args[0] == "repo" && args[1] == "delete" {
					deleteCount++
					return []byte(""), nil
				}
				return nil, nil
			},
		}
		client := NewClient("", mock)
		results := client.DeleteRepos(context.Background(), repos, nil)
		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		if deleteCount != 2 {
			t.Errorf("expected 2 delete calls, got %d", deleteCount)
		}
		for i, r := range results {
			if !r.Success {
				t.Errorf("result %d: expected success", i)
			}
			if r.Error != nil {
				t.Errorf("result %d: unexpected error: %v", i, r.Error)
			}
		}
	})

	t.Run("partial failure", func(t *testing.T) {
		repos := []Repo{
			{Name: "repo1", NameWithOwner: "owner/repo1"},
			{Name: "repo2", NameWithOwner: "owner/repo2"},
			{Name: "repo3", NameWithOwner: "owner/repo3"},
		}
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				if len(args) >= 4 && args[0] == "repo" && args[1] == "delete" {
					// repo2 fails
					if args[2] == "owner/repo2" {
						return []byte("HTTP 404: Not Found"), errors.New("exit status 1")
					}
					return []byte(""), nil
				}
				return nil, nil
			},
		}
		client := NewClient("", mock)
		results := client.DeleteRepos(context.Background(), repos, nil)
		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}
		if !results[0].Success {
			t.Error("repo1 should succeed")
		}
		if results[1].Success {
			t.Error("repo2 should fail")
		}
		if results[1].Error == nil {
			t.Error("repo2 should have error")
		}
		if !results[2].Success {
			t.Error("repo3 should succeed")
		}
	})

	t.Run("progress callback", func(t *testing.T) {
		repos := []Repo{
			{Name: "repo1", NameWithOwner: "owner/repo1"},
			{Name: "repo2", NameWithOwner: "owner/repo2"},
		}
		progressCalls := []struct {
			current int
			total   int
		}{}
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				return []byte(""), nil
			},
		}
		client := NewClient("", mock)
		callback := func(repo Repo, result DeleteResult, current, total int) {
			progressCalls = append(progressCalls, struct {
				current int
				total   int
			}{current, total})
		}
		_ = client.DeleteRepos(context.Background(), repos, callback)
		if len(progressCalls) != 2 {
			t.Fatalf("expected 2 progress calls, got %d", len(progressCalls))
		}
		if progressCalls[0].current != 1 || progressCalls[0].total != 2 {
			t.Errorf("first call: expected (1,2), got (%d,%d)", progressCalls[0].current, progressCalls[0].total)
		}
		if progressCalls[1].current != 2 || progressCalls[1].total != 2 {
			t.Errorf("second call: expected (2,2), got (%d,%d)", progressCalls[1].current, progressCalls[1].total)
		}
	})

	t.Run("throttle timing", func(t *testing.T) {
		// This test verifies that there's a delay between deletions
		repos := []Repo{
			{Name: "repo1", NameWithOwner: "owner/repo1"},
			{Name: "repo2", NameWithOwner: "owner/repo2"},
			{Name: "repo3", NameWithOwner: "owner/repo3"},
		}
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				return []byte(""), nil
			},
		}
		client := NewClient("", mock)
		start := time.Now()
		_ = client.DeleteRepos(context.Background(), repos, nil)
		elapsed := time.Since(start)
		// With 3 repos, we should have 2 throttles of 1 second each
		// So minimum 2 seconds (allow some margin for test execution)
		if elapsed < 1900*time.Millisecond {
			t.Errorf("expected at least 2 seconds for 3 repos, got %v", elapsed)
		}
		// But not too much - should be roughly 2 seconds
		if elapsed > 4*time.Second {
			t.Errorf("expected roughly 2 seconds, got %v (too long)", elapsed)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		repos := []Repo{
			{Name: "repo1", NameWithOwner: "owner/repo1"},
			{Name: "repo2", NameWithOwner: "owner/repo2"},
		}
		mock := &mockExecutor{
			executeFunc: func(name string, args ...string) ([]byte, error) {
				return []byte(""), nil
			},
		}
		client := NewClient("", mock)
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately
		results := client.DeleteRepos(ctx, repos, nil)
		// Should have 0 results since context is cancelled before first iteration
		// (ctx.Err() is checked at start of loop)
		if len(results) > 0 {
			t.Errorf("expected 0 results due to cancellation, got %d", len(results))
		}
	})
}

// TestCheckAuth tests the CheckAuth method
func TestCheckAuth(t *testing.T) {
	t.Run("gh installed and authenticated", func(t *testing.T) {
		mock := &mockExecutor{
			lookPathRes: "/usr/bin/gh",
			lookPathErr: nil,
			executeFunc: func(name string, args ...string) ([]byte, error) {
				if len(args) >= 2 && args[0] == "auth" && args[1] == "status" {
					// gh auth status succeeds with output on stderr but exit 0
					return []byte(""), nil
				}
				return nil, nil
			},
		}
		client := NewClient("", mock)
		err := client.CheckAuth(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("gh not in PATH", func(t *testing.T) {
		mock := &mockExecutor{
			lookPathRes: "",
			lookPathErr: errors.New("file not found"),
		}
		client := NewClient("", mock)
		err := client.CheckAuth(context.Background())
		if err != ErrGhNotInstalled {
			t.Errorf("expected ErrGhNotInstalled, got: %v", err)
		}
	})

	t.Run("not authenticated", func(t *testing.T) {
		mock := &mockExecutor{
			lookPathRes: "/usr/bin/gh",
			lookPathErr: nil,
			executeFunc: func(name string, args ...string) ([]byte, error) {
				if len(args) >= 2 && args[0] == "auth" && args[1] == "status" {
					return []byte("not logged in"), errors.New("exit status 1")
				}
				return nil, nil
			},
		}
		client := NewClient("", mock)
		err := client.CheckAuth(context.Background())
		if err != ErrNotAuthenticated {
			t.Errorf("expected ErrNotAuthenticated, got: %v", err)
		}
	})
}

// TestNewClient tests the NewClient constructor
func TestNewClient(t *testing.T) {
	t.Run("with nil executor uses default", func(t *testing.T) {
		client := NewClient("testowner", nil)
		if client.Owner != "testowner" {
			t.Errorf("expected owner 'testowner', got %s", client.Owner)
		}
		if client.executor == nil {
			t.Error("expected non-nil executor")
		}
		// Verify it's the default implementation
		_, ok := client.executor.(*execCommandExecutor)
		if !ok {
			t.Error("expected execCommandExecutor type")
		}
	})

	t.Run("with custom executor", func(t *testing.T) {
		mock := &mockExecutor{}
		client := NewClient("", mock)
		if client.executor != mock {
			t.Error("expected custom executor to be used")
		}
	})
}

// Example test showing the mockExecutor in action
func Example_mockExecutor() {
	mock := &mockExecutor{
		lookPathRes: "/usr/bin/gh",
		executeFunc: func(name string, args ...string) ([]byte, error) {
			return []byte(`[{"name":"test","nameWithOwner":"owner/test","pushedAt":"2024-01-01T00:00:00Z"}]`), nil
		},
	}
	client := NewClient("", mock)
	repos, _ := client.ListRepos(context.Background())
	fmt.Printf("Found %d repos\n", len(repos))
	// Output: Found 1 repos
}
