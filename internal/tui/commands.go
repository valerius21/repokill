package tui

import (
	"context"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/valerius21/repokill/internal/github"
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

func fetchRepos(client *github.Client) tea.Cmd {
	return func() tea.Msg {
		repos, err := client.ListRepos(context.Background())
		if err != nil {
			return reposLoadErrorMsg(err)
		}
		return reposLoadedMsg(repos)
	}
}

func deleteReposCmd(client *github.Client, repos []github.Repo) tea.Cmd {
	return func() tea.Msg {
		results := client.DeleteRepos(context.Background(), repos, nil)
		return allDeletesDoneMsg{results: results}
	}
}
