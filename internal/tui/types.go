// Package tui provides types for the terminal user interface.
package tui

import (
	"github.com/valerius21/repokill/internal/filter"
)

// AppState represents the current state of the TUI application.
type AppState int

const (
	StateLoading AppState = iota
	StateList
	StateSearch
	StateConfirm
	StateDeleting
	StateError
)

// Config holds the application configuration.
type Config struct {
	Owner         string
	FilterOptions filter.FilterOptions
	SortOptions   filter.SortOptions
}
