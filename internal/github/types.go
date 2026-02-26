package github

import "time"

// Repo represents a GitHub repository with fields matching gh CLI output.
type Repo struct {
	Name           string    `json:"name"`
	NameWithOwner  string    `json:"nameWithOwner"`
	Description    string    `json:"description"`
	PushedAt       time.Time `json:"pushedAt"`
	Visibility     string    `json:"visibility"`
	IsArchived     bool      `json:"isArchived"`
	IsFork         bool      `json:"isFork"`
	StargazerCount int       `json:"stargazerCount"`
	ForkCount      int       `json:"forkCount"`
}

// RepoList is a collection of repositories.
type RepoList []Repo

// DeleteResult represents the outcome of a repository deletion attempt.
type DeleteResult struct {
	Repo     Repo
	Success  bool
	Error    error
	Duration time.Duration
}
