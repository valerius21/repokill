package tui

import "github.com/charmbracelet/lipgloss"

var (
	// TitleStyle is the style for the application title
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	// NormalRowStyle is the style for a normal row in the list
	NormalRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EEEEEE"))

	// SelectedRowStyle is the style for the currently selected row
	SelectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4"))

	// MarkedRowStyle is the style for a row marked for deletion
	MarkedRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87"))

	// MarkedSelectedRowStyle is the style for a marked row that is also selected
	MarkedSelectedRowStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF5F87")).
				Underline(true)

	// DeletedRowStyle is the style for a row that has been deleted
	DeletedRowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4A4A4A")).
			Strikethrough(true)

	// ColumnHeaderStyle is the style for table column headers
	ColumnHeaderStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#7D56F4")).
				Border(lipgloss.NormalBorder(), false, false, true, false)

	// StatusBarStyle is the style for the status bar at the bottom
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)

	// SearchStyle is the style for the search prompt and input
	SearchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D7FF")).
			Bold(true)

	// ConfirmStyle is the style for confirmation prompts
	ConfirmStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFAF00")).
			Bold(true)

	// ProgressStyle is the style for progress bars
	ProgressStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00D700"))

	// ErrorStyle is the style for error messages
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF0000")).
			Bold(true)

	// SuccessStyle is the style for success messages
	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)

	// StarStyle is the style for repository stars
	StarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFF00"))
)
