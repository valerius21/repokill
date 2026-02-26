// Package main provides the entry point for repokill, a TUI tool for bulk-deleting
// GitHub repositories.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/valerius21/repokill/internal/filter"
	"github.com/valerius21/repokill/internal/github"
	"github.com/valerius21/repokill/internal/tui"
)

// Version is set at build time via ldflags.
var Version = "dev"

func main() {
	org := flag.String("org", "", "List repos for an organization")
	public := flag.Bool("public", false, "Show only public repos")
	private := flag.Bool("private", false, "Show only private repos")
	archived := flag.Bool("archived", false, "Show only archived repos")
	forked := flag.Bool("forked", false, "Show only forks")
	version := flag.Bool("version", false, "Show version")
	flag.Parse()

	if *version {
		fmt.Printf("repokill %s\n", Version)
		os.Exit(0)
	}

	filterOpts := buildFilterOptions(*public, *private, *archived, *forked)
	sortOpts := filter.SortOptions{
		Field: filter.SortByPushedAt,
		Order: filter.Ascending,
	}

	client := github.NewClient(*org, nil)

	if err := client.CheckAuth(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	model := tui.New(client, filterOpts, sortOpts)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}

// buildFilterOptions constructs filter options from command-line flags.
func buildFilterOptions(public, private, archived, forked bool) filter.FilterOptions {
	var opts filter.FilterOptions

	if public {
		opts.Visibility = "public"
	} else if private {
		opts.Visibility = "private"
	}

	if archived {
		trueVal := true
		opts.Archived = &trueVal
	}

	if forked {
		trueVal := true
		opts.Forked = &trueVal
	}

	return opts
}
