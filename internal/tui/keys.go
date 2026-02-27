// Package tui provides keybindings for the terminal user interface.
package tui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the keybindings for the TUI
type KeyMap struct {
	Up                      key.Binding
	Down                    key.Binding
	ToggleMark              key.Binding
	ConfirmDelete           key.Binding
	ConfirmArchive          key.Binding
	Search                  key.Binding
	SelectAll               key.Binding
	Quit                    key.Binding
	PageUp                  key.Binding
	PageDown                key.Binding
	Home                    key.Binding
	End                     key.Binding
	Help                    key.Binding
	Esc                     key.Binding
	ToggleVisibilityPrivate key.Binding
	ToggleVisibilityPublic  key.Binding
	Backup                  key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/↓", "down"),
		),
		ToggleMark: key.NewBinding(
			key.WithKeys(" "),
			key.WithHelp("space", "toggle mark"),
		),
		ConfirmDelete: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "confirm delete"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		SelectAll: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "select all"),
		),
		ConfirmArchive: key.NewBinding(
			key.WithKeys("A"),
			key.WithHelp("A", "archive"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
		PageUp: key.NewBinding(
			key.WithKeys("pgup", "u"),
			key.WithHelp("u/pgup", "page up"),
		),
		PageDown: key.NewBinding(
			key.WithKeys("pgdown", "d"),
			key.WithHelp("d/pgdown", "page down"),
		),
		Home: key.NewBinding(
			key.WithKeys("home", "g"),
			key.WithHelp("g/home", "home"),
		),
		End: key.NewBinding(
			key.WithKeys("end", "G"),
			key.WithHelp("G/end", "end"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Esc: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		ToggleVisibilityPrivate: key.NewBinding(
			key.WithKeys("p"),
			key.WithHelp("p", "make selected private"),
		),
		ToggleVisibilityPublic: key.NewBinding(
			key.WithKeys("P"),
			key.WithHelp("P", "make selected public"),
		),
		Backup: key.NewBinding(
			key.WithKeys("B"),
			key.WithHelp("B", "backup selected"),
		),
	}
}

// ShortHelp returns a slice of keybindings for the short help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Quit}
}

// FullHelp returns a slice of keybindings for the full help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.ToggleMark, k.Search},
		{k.PageUp, k.PageDown, k.Home, k.End},
		{k.SelectAll, k.ConfirmDelete, k.ConfirmArchive, k.Backup},
		{k.ToggleVisibilityPrivate, k.ToggleVisibilityPublic, k.Help, k.Quit},
	}
}
