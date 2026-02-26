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

	// Build filter options
	var filterOpts filter.FilterOptions
	if *public {
		filterOpts.Visibility = "public"
	} else if *private {
		filterOpts.Visibility = "private"
	}
	if *archived {
		trueVal := true
		filterOpts.Archived = &trueVal
	}
	if *forked {
		trueVal := true
		filterOpts.Forked = &trueVal
	}

	// Default sort: by pushedAt ascending (oldest first)
	sortOpts := filter.SortOptions{
		Field: filter.SortByPushedAt,
		Order: filter.Ascending,
	}

	// Create GitHub client
	client := github.NewClient(*org, nil)

	// Check authentication
	if err := client.CheckAuth(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Run TUI
	model := tui.New(client, filterOpts, sortOpts)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
		os.Exit(1)
	}
}
