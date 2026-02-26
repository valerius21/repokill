// Package filter provides repository filtering and sorting functionality.
package filter

import (
	"slices"
	"strings"

	"github.com/valerius21/repokill/internal/github"
)

// Filter returns a new slice of repositories that match the given criteria.
func Filter(repos []github.Repo, opts FilterOptions) []github.Repo {
	var filtered []github.Repo

	for _, repo := range repos {
		if opts.Visibility != "" && !strings.EqualFold(repo.Visibility, opts.Visibility) {
			continue
		}

		if opts.Archived != nil && repo.IsArchived != *opts.Archived {
			continue
		}

		if opts.Forked != nil && repo.IsFork != *opts.Forked {
			continue
		}

		if opts.SearchQuery != "" {
			query := strings.ToLower(opts.SearchQuery)
			name := strings.ToLower(repo.Name)
			desc := strings.ToLower(repo.Description)
			if !strings.Contains(name, query) && !strings.Contains(desc, query) {
				continue
			}
		}

		filtered = append(filtered, repo)
	}

	return filtered
}

// Sort returns a new slice of repositories sorted according to the given options.
func Sort(repos []github.Repo, opts SortOptions) []github.Repo {
	sorted := make([]github.Repo, len(repos))
	copy(sorted, repos)

	slices.SortFunc(sorted, func(a, b github.Repo) int {
		var cmp int
		switch opts.Field {
		case SortByName:
			cmp = strings.Compare(a.Name, b.Name)
		case SortByStars:
			if a.StargazerCount < b.StargazerCount {
				cmp = -1
			} else if a.StargazerCount > b.StargazerCount {
				cmp = 1
			} else {
				cmp = 0
			}
		case SortByPushedAt:
			fallthrough
		default:
			if a.PushedAt.Before(b.PushedAt) {
				cmp = -1
			} else if a.PushedAt.After(b.PushedAt) {
				cmp = 1
			} else {
				cmp = 0
			}
		}

		if opts.Order == Descending {
			return -cmp
		}
		return cmp
	})

	return sorted
}

// FilterAndSort applies both filtering and sorting to the input repositories.
func FilterAndSort(repos []github.Repo, filterOpts FilterOptions, sortOpts SortOptions) []github.Repo {
	filtered := Filter(repos, filterOpts)
	return Sort(filtered, sortOpts)
}
