package filter

import (
	"slices"
	"testing"
	"time"

	"github.com/valerius21/repokill/internal/github"
)

// makeRepo creates a test repository with the given attributes.
func makeRepo(name, visibility string, archived, fork bool, stars int, pushed time.Time) github.Repo {
	return github.Repo{
		Name:           name,
		Visibility:     visibility,
		IsArchived:     archived,
		IsFork:         fork,
		StargazerCount: stars,
		PushedAt:       pushed,
	}
}

func TestFilter(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	testRepos := []github.Repo{
		makeRepo("public-repo", "public", false, false, 10, baseTime),
		makeRepo("private-repo", "private", false, false, 5, baseTime.Add(1*time.Hour)),
		makeRepo("archived-repo", "public", true, false, 20, baseTime.Add(2*time.Hour)),
		makeRepo("forked-repo", "public", false, true, 2, baseTime.Add(3*time.Hour)),
		makeRepo("TEST-CASE-SENSITIVE", "public", false, false, 15, baseTime.Add(4*time.Hour)),
		makeRepo("golang-project", "public", false, false, 50, baseTime.Add(5*time.Hour)),
	}

	t.Run("no filters returns all repos", func(t *testing.T) {
		opts := FilterOptions{}
		result := Filter(testRepos, opts)

		if len(result) != len(testRepos) {
			t.Errorf("expected %d repos, got %d", len(testRepos), len(result))
		}
	})

	t.Run("filter by visibility public", func(t *testing.T) {
		opts := FilterOptions{Visibility: "public"}
		result := Filter(testRepos, opts)

		for _, r := range result {
			if r.Visibility != "public" {
				t.Errorf("expected public visibility, got %s", r.Visibility)
			}
		}
		if len(result) != 5 {
			t.Errorf("expected 5 public repos, got %d", len(result))
		}
	})

	t.Run("filter by visibility private", func(t *testing.T) {
		opts := FilterOptions{Visibility: "private"}
		result := Filter(testRepos, opts)

		for _, r := range result {
			if r.Visibility != "private" {
				t.Errorf("expected private visibility, got %s", r.Visibility)
			}
		}
		if len(result) != 1 {
			t.Errorf("expected 1 private repo, got %d", len(result))
		}
	})

	t.Run("filter by archived true", func(t *testing.T) {
		archived := true
		opts := FilterOptions{Archived: &archived}
		result := Filter(testRepos, opts)

		for _, r := range result {
			if !r.IsArchived {
				t.Error("expected archived repos only")
			}
		}
		if len(result) != 1 {
			t.Errorf("expected 1 archived repo, got %d", len(result))
		}
	})

	t.Run("filter by archived false", func(t *testing.T) {
		archived := false
		opts := FilterOptions{Archived: &archived}
		result := Filter(testRepos, opts)

		for _, r := range result {
			if r.IsArchived {
				t.Error("expected non-archived repos only")
			}
		}
		if len(result) != 5 {
			t.Errorf("expected 5 non-archived repos, got %d", len(result))
		}
	})

	t.Run("filter by forked true", func(t *testing.T) {
		forked := true
		opts := FilterOptions{Forked: &forked}
		result := Filter(testRepos, opts)

		for _, r := range result {
			if !r.IsFork {
				t.Error("expected forked repos only")
			}
		}
		if len(result) != 1 {
			t.Errorf("expected 1 forked repo, got %d", len(result))
		}
	})

	t.Run("filter by search query case insensitive", func(t *testing.T) {
		opts := FilterOptions{SearchQuery: "test"}
		result := Filter(testRepos, opts)

		if len(result) != 1 {
			t.Errorf("expected 1 repo matching 'test', got %d", len(result))
		}
		if len(result) > 0 && result[0].Name != "TEST-CASE-SENSITIVE" {
			t.Errorf("expected TEST-CASE-SENSITIVE, got %s", result[0].Name)
		}
	})

	t.Run("filter by search query partial match", func(t *testing.T) {
		opts := FilterOptions{SearchQuery: "golang"}
		result := Filter(testRepos, opts)

		if len(result) != 1 {
			t.Errorf("expected 1 repo matching 'golang', got %d", len(result))
		}
	})

	t.Run("empty search query matches all", func(t *testing.T) {
		opts := FilterOptions{SearchQuery: ""}
		result := Filter(testRepos, opts)

		if len(result) != len(testRepos) {
			t.Errorf("expected %d repos with empty query, got %d", len(testRepos), len(result))
		}
	})

	t.Run("combined filters", func(t *testing.T) {
		archived := false
		opts := FilterOptions{
			Visibility: "public",
			Archived:   &archived,
		}
		result := Filter(testRepos, opts)

		for _, r := range result {
			if r.Visibility != "public" {
				t.Errorf("expected public, got %s", r.Visibility)
			}
			if r.IsArchived {
				t.Error("expected non-archived")
			}
		}
		if len(result) != 4 {
			t.Errorf("expected 4 repos with combined filters, got %d", len(result))
		}
	})

	t.Run("does not mutate input slice", func(t *testing.T) {
		original := make([]github.Repo, len(testRepos))
		copy(original, testRepos)

		opts := FilterOptions{Visibility: "public"}
		_ = Filter(testRepos, opts)

		if !slices.Equal(original, testRepos) {
			t.Error("input slice was mutated")
		}
	})
}

func TestSort(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	testRepos := []github.Repo{
		makeRepo("charlie", "public", false, false, 30, baseTime),
		makeRepo("alpha", "public", false, false, 10, baseTime.Add(2*time.Hour)),
		makeRepo("beta", "public", false, false, 20, baseTime.Add(1*time.Hour)),
	}

	t.Run("sort by pushed at ascending", func(t *testing.T) {
		opts := SortOptions{Field: SortByPushedAt, Order: Ascending}
		result := Sort(testRepos, opts)

		if result[0].Name != "charlie" {
			t.Errorf("expected charlie first (oldest), got %s", result[0].Name)
		}
		if result[2].Name != "alpha" {
			t.Errorf("expected alpha last (newest), got %s", result[2].Name)
		}
	})

	t.Run("sort by pushed at descending", func(t *testing.T) {
		opts := SortOptions{Field: SortByPushedAt, Order: Descending}
		result := Sort(testRepos, opts)

		if result[0].Name != "alpha" {
			t.Errorf("expected alpha first (newest), got %s", result[0].Name)
		}
		if result[2].Name != "charlie" {
			t.Errorf("expected charlie last (oldest), got %s", result[2].Name)
		}
	})

	t.Run("sort by name ascending", func(t *testing.T) {
		opts := SortOptions{Field: SortByName, Order: Ascending}
		result := Sort(testRepos, opts)

		if result[0].Name != "alpha" {
			t.Errorf("expected alpha first, got %s", result[0].Name)
		}
		if result[1].Name != "beta" {
			t.Errorf("expected beta second, got %s", result[1].Name)
		}
		if result[2].Name != "charlie" {
			t.Errorf("expected charlie third, got %s", result[2].Name)
		}
	})

	t.Run("sort by stars descending", func(t *testing.T) {
		opts := SortOptions{Field: SortByStars, Order: Descending}
		result := Sort(testRepos, opts)

		if result[0].StargazerCount != 30 {
			t.Errorf("expected 30 stars first, got %d", result[0].StargazerCount)
		}
		if result[1].StargazerCount != 20 {
			t.Errorf("expected 20 stars second, got %d", result[1].StargazerCount)
		}
		if result[2].StargazerCount != 10 {
			t.Errorf("expected 10 stars third, got %d", result[2].StargazerCount)
		}
	})

	t.Run("does not mutate input slice", func(t *testing.T) {
		original := make([]github.Repo, len(testRepos))
		copy(original, testRepos)

		opts := SortOptions{Field: SortByName, Order: Ascending}
		_ = Sort(testRepos, opts)

		if !slices.Equal(original, testRepos) {
			t.Error("input slice was mutated")
		}
	})
}

func TestFilterAndSort(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	testRepos := []github.Repo{
		makeRepo("public-old", "public", false, false, 5, baseTime),
		makeRepo("private-repo", "private", false, false, 10, baseTime.Add(1*time.Hour)),
		makeRepo("public-new", "public", false, false, 20, baseTime.Add(2*time.Hour)),
	}

	t.Run("applies both filter and sort", func(t *testing.T) {
		filterOpts := FilterOptions{Visibility: "public"}
		sortOpts := SortOptions{Field: SortByPushedAt, Order: Descending}

		result := FilterAndSort(testRepos, filterOpts, sortOpts, nil)
		if len(result) != 2 {
			t.Errorf("expected 2 public repos, got %d", len(result))
		}

		// Should be sorted newest first
		if result[0].Name != "public-new" {
			t.Errorf("expected public-new first (newest), got %s", result[0].Name)
		}
		if result[1].Name != "public-old" {
			t.Errorf("expected public-old second (oldest), got %s", result[1].Name)
		}
	})

	t.Run("does not mutate input slice", func(t *testing.T) {
		original := make([]github.Repo, len(testRepos))
		copy(original, testRepos)

		filterOpts := FilterOptions{Visibility: "public"}
		sortOpts := SortOptions{Field: SortByName, Order: Ascending}
		_ = FilterAndSort(testRepos, filterOpts, sortOpts, nil)
		if !slices.Equal(original, testRepos) {
			t.Error("input slice was mutated")
		}
	})
}

func TestFilterAndSortProcessedRepos(t *testing.T) {
	baseTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
	testRepos := []github.Repo{
		{Name: "repo-a", NameWithOwner: "owner/repo-a", Visibility: "public", StargazerCount: 10, PushedAt: baseTime},
		{Name: "repo-b", NameWithOwner: "owner/repo-b", Visibility: "public", StargazerCount: 20, PushedAt: baseTime.Add(1 * time.Hour)},
		{Name: "repo-c", NameWithOwner: "owner/repo-c", Visibility: "public", StargazerCount: 30, PushedAt: baseTime.Add(2 * time.Hour)},
	}

	t.Run("processed repos sorted to bottom", func(t *testing.T) {
		// Mark repo-b as processed (deleted)
		processedRepos := map[string]string{
			"owner/repo-b": "deleted",
		}

		filterOpts := FilterOptions{}
		sortOpts := SortOptions{Field: SortByStars, Order: Descending}

		result := FilterAndSort(testRepos, filterOpts, sortOpts, processedRepos)
		if len(result) != 3 {
			t.Errorf("expected 3 repos, got %d", len(result))
		}

		// Unprocessed repos should come first, sorted by stars descending
		// repo-c (30 stars) should be first, repo-a (10 stars) second
		// repo-b (processed) should be last
		if result[0].Name != "repo-c" {
			t.Errorf("expected repo-c first (30 stars), got %s", result[0].Name)
		}
		if result[1].Name != "repo-a" {
			t.Errorf("expected repo-a second (10 stars), got %s", result[1].Name)
		}
		if result[2].Name != "repo-b" {
			t.Errorf("expected repo-b last (processed), got %s", result[2].Name)
		}
	})

	t.Run("multiple processed repos sorted among themselves", func(t *testing.T) {
		// Mark repo-a and repo-c as processed
		processedRepos := map[string]string{
			"owner/repo-a": "archived",
			"owner/repo-c": "deleted",
		}

		filterOpts := FilterOptions{}
		sortOpts := SortOptions{Field: SortByStars, Order: Descending}

		result := FilterAndSort(testRepos, filterOpts, sortOpts, processedRepos)
		if len(result) != 3 {
			t.Errorf("expected 3 repos, got %d", len(result))
		}

		// Only repo-b is unprocessed, should be first
		// Then processed repos sorted by stars descending: repo-c (30), repo-a (10)
		if result[0].Name != "repo-b" {
			t.Errorf("expected repo-b first (unprocessed), got %s", result[0].Name)
		}
		if result[1].Name != "repo-c" {
			t.Errorf("expected repo-c second (processed, 30 stars), got %s", result[1].Name)
		}
		if result[2].Name != "repo-a" {
			t.Errorf("expected repo-a third (processed, 10 stars), got %s", result[2].Name)
		}
	})

	t.Run("nil processedRepos map", func(t *testing.T) {
		filterOpts := FilterOptions{}
		sortOpts := SortOptions{Field: SortByStars, Order: Descending}

		result := FilterAndSort(testRepos, filterOpts, sortOpts, nil)
		if len(result) != 3 {
			t.Errorf("expected 3 repos, got %d", len(result))
		}

		// All repos unprocessed, should be sorted by stars descending
		if result[0].Name != "repo-c" {
			t.Errorf("expected repo-c first (30 stars), got %s", result[0].Name)
		}
		if result[1].Name != "repo-b" {
			t.Errorf("expected repo-b second (20 stars), got %s", result[1].Name)
		}
		if result[2].Name != "repo-a" {
			t.Errorf("expected repo-a third (10 stars), got %s", result[2].Name)
		}
	})
}
