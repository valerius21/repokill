// Package tui provides the Bubble Tea model for the terminal user interface.
package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/valerius21/repokill/internal/filter"
	"github.com/valerius21/repokill/internal/github"
)

type flashTimeoutMsg struct{}

// Model represents the TUI application state.
type Model struct {
	repos         []github.Repo
	filtered      []github.Repo
	cursor        int
	selected      map[int]struct{}
	state         AppState
	viewport      viewport.Model
	width         int
	height        int
	client        *github.Client
	keys          KeyMap
	err           error
	filterOpts    filter.FilterOptions
	sortOpts      filter.SortOptions
	deleteResults []github.DeleteResult
	deletingRepos []github.Repo
	flashMessage  string
	searchInput   textinput.Model
}

// New creates a new TUI model.
func New(client *github.Client, filterOpts filter.FilterOptions, sortOpts filter.SortOptions) Model {
	ti := textinput.New()
	ti.Placeholder = "Search repos..."
	ti.PromptStyle = SearchStyle
	ti.TextStyle = SearchStyle

	return Model{
		client:      client,
		keys:        DefaultKeyMap(),
		state:       StateLoading,
		selected:    make(map[int]struct{}),
		filterOpts:  filterOpts,
		sortOpts:    sortOpts,
		searchInput: ti,
	}
}

// Init initializes the TUI.
func (m Model) Init() tea.Cmd {
	return fetchRepos(m.client)
}

// Update handles messages and updates the model.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 6 // Adjust for header and status bar
		m.updateViewport()

	case reposLoadedMsg:
		m.repos = msg
		m.filtered = filter.FilterAndSort(m.repos, m.filterOpts, m.sortOpts)
		m.state = StateList
		m.updateViewport()

	case reposLoadErrorMsg:
		m.err = msg
		m.state = StateError

	case tea.KeyMsg:
		// Search mode: forward all keys to textinput
		if m.state == StateSearch {
			switch msg.Type {
			case tea.KeyEnter:
				// Accept search and return to list (filter stays active)
				m.state = StateList
				m.searchInput.Blur()
				return m, nil
			case tea.KeyEsc:
				// Cancel search: clear query and restore unfiltered list
				m.state = StateList
				m.filterOpts.SearchQuery = ""
				m.searchInput.SetValue("")
				m.searchInput.Blur()
				m.filtered = filter.FilterAndSort(m.repos, m.filterOpts, m.sortOpts)
				m.cursor = 0
				m.updateViewport()
				return m, nil
			}

			// Real-time filtering: every keystroke triggers filter
			var tiCmd tea.Cmd
			m.searchInput, tiCmd = m.searchInput.Update(msg)
			m.filterOpts.SearchQuery = m.searchInput.Value()
			m.filtered = filter.FilterAndSort(m.repos, m.filterOpts, m.sortOpts)
			m.cursor = 0
			m.updateViewport()
			return m, tiCmd
		}

		// Normal key handling for non-search states
		switch {
		case key.Matches(msg, m.keys.Quit):
			return m, tea.Quit

		case msg.String() == "/":
			if m.state == StateList {
				m.state = StateSearch
				m.searchInput.Focus()
				m.searchInput.SetValue(m.filterOpts.SearchQuery)
				return m, textinput.Blink
			}

		case key.Matches(msg, m.keys.Up):
			if m.cursor > 0 {
				m.cursor--
				m.fixViewportScroll()
			}

		case key.Matches(msg, m.keys.Down):
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.fixViewportScroll()
			}

		case key.Matches(msg, m.keys.PageUp):
			m.cursor -= m.viewport.Height
			if m.cursor < 0 {
				m.cursor = 0
			}
			m.fixViewportScroll()

		case key.Matches(msg, m.keys.PageDown):
			m.cursor += m.viewport.Height
			if m.cursor >= len(m.filtered) {
				m.cursor = len(m.filtered) - 1
			}
			m.fixViewportScroll()

		case key.Matches(msg, m.keys.Home):
			m.cursor = 0
			m.fixViewportScroll()

		case key.Matches(msg, m.keys.End):
			m.cursor = len(m.filtered) - 1
			m.fixViewportScroll()

		case key.Matches(msg, m.keys.ToggleMark):
			if _, ok := m.selected[m.cursor]; ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}
			m.updateViewport()

		case key.Matches(msg, m.keys.SelectAll):
			if len(m.selected) == len(m.filtered) {
				m.selected = make(map[int]struct{})
			} else {
				for i := range m.filtered {
					m.selected[i] = struct{}{}
				}
			}
			m.updateViewport()

		case key.Matches(msg, m.keys.ConfirmDelete):
			if m.state == StateList {
				if len(m.selected) == 0 {
					m.flashMessage = "No repos selected"
					return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
						return flashTimeoutMsg{}
					})
				} else {
					m.state = StateConfirm
				}
			} else if m.state == StateDeleting {
				// If deletion is done, return to list
				if len(m.deleteResults) == len(m.deletingRepos) {
					m.state = StateList
					m.selected = make(map[int]struct{})
					return m, fetchRepos(m.client)
				}
			}

		case msg.String() == "y" || msg.String() == "Y":
			if m.state == StateConfirm {
				m.state = StateDeleting
				m.deleteResults = nil
				m.deletingRepos = make([]github.Repo, 0, len(m.selected))
				for i := range m.selected {
					m.deletingRepos = append(m.deletingRepos, m.filtered[i])
				}
				return m, deleteReposCmd(m.client, m.deletingRepos)
			}

		case msg.String() == "n" || msg.String() == "N" || key.Matches(msg, m.keys.Esc):
			if m.state == StateConfirm {
				m.state = StateList
			}
		}

	case flashTimeoutMsg:
		m.flashMessage = ""
		return m, nil

	case repoDeletedMsg:
		m.deleteResults = append(m.deleteResults, msg.result)
		return m, nil

	case allDeletesDoneMsg:
		m.deleteResults = msg.results
		return m, nil
	}

	return m, cmd
}

func (m *Model) fixViewportScroll() {
	if m.cursor < m.viewport.YOffset {
		m.viewport.YOffset = m.cursor
	} else if m.cursor >= m.viewport.YOffset+m.viewport.Height {
		m.viewport.YOffset = m.cursor - m.viewport.Height + 1
	}
	m.updateViewport()
}

func (m *Model) updateViewport() {
	var b strings.Builder
	for i, repo := range m.filtered {
		b.WriteString(m.renderRow(i, repo))
		if i < len(m.filtered)-1 {
			b.WriteString("\n")
		}
	}
	m.viewport.SetContent(b.String())
}

func (m Model) renderRow(index int, repo github.Repo) string {
	isSelected := m.cursor == index
	_, isMarked := m.selected[index]

	var style lipgloss.Style
	if isMarked && isSelected {
		style = MarkedSelectedRowStyle
	} else if isMarked {
		style = MarkedRowStyle
	} else if isSelected {
		style = SelectedRowStyle
	} else {
		style = NormalRowStyle
	}

	checkbox := "[ ]"
	if isMarked {
		checkbox = "[x]"
	}

	// Relative time
	relTime := formatRelativeTime(repo.PushedAt)

	// Badges
	badges := ""
	if repo.IsArchived {
		badges += " ARCH"
	}
	if repo.IsFork {
		badges += " FORK"
	}
	visibility := strings.ToUpper(repo.Visibility)

	// Row layout: [x] Name (Visibility) | Stars | Forks | Pushed | Flags
	row := fmt.Sprintf("%s %-25s %-4s %-5d %-5d %-15s %s",
		checkbox,
		truncate(repo.Name, 25),
		visibility[:3],
		repo.StargazerCount,
		repo.ForkCount,
		relTime,
		badges,
	)

	return style.Render(row)
}

func (m Model) View() string {
	switch m.state {
	case StateLoading:
		return "\n  Loading repos...\n"
	case StateError:
		return ErrorStyle.Render(fmt.Sprintf("\n  Error: %v\n", m.err))
	case StateList:
		header := TitleStyle.Render(" GitHub Repo Cleaner ") + "\n\n"

		// Column Headers
		colHeaders := ColumnHeaderStyle.Render(fmt.Sprintf("    %-25s %-4s %-5s %-5s %-15s %s",
			"NAME", "VIS", "STAR", "FORK", "PUSHED", "FLAGS")) + "\n"

		// Status Bar
		status := fmt.Sprintf(" %d selected | %d repos | ↑↓ navigate | space mark | / search | enter delete | q quit ",
			len(m.selected), len(m.filtered))
		if m.filterOpts.SearchQuery != "" {
			status = fmt.Sprintf(" 🔍 %q | %d results | %d selected | ↑↓ navigate | space mark | / search | enter delete | q quit ",
				m.filterOpts.SearchQuery, len(m.filtered), len(m.selected))
		}
		if m.flashMessage != "" {
			status = " " + ErrorStyle.Render(m.flashMessage) + " "
		}
		statusBar := StatusBarStyle.Render(status)

		return header + colHeaders + m.viewport.View() + "\n" + statusBar
	case StateConfirm:
		return m.confirmView()
	case StateSearch:
		return m.searchView()
	case StateDeleting:
		return m.deletingView()
	default:
		return ""
	}
}

func (m Model) searchView() string {
	header := TitleStyle.Render(" GitHub Repo Cleaner ") + "\n\n"
	colHeaders := ColumnHeaderStyle.Render(fmt.Sprintf("    %-25s %-4s %-5s %-5s %-15s %s",
		"NAME", "VIS", "STAR", "FORK", "PUSHED", "FLAGS")) + "\n"

	content := m.viewport.View()
	if len(m.filtered) == 0 {
		content = "\n  No repos match your search\n"
	}

	searchPrompt := SearchStyle.Render(" 🔍 Search: ") + m.searchInput.View()
	status := fmt.Sprintf(" %d repos found | enter accept | esc cancel ", len(m.filtered))
	statusBar := StatusBarStyle.Render(status)

	return header + colHeaders + content + "\n" + searchPrompt + "\n" + statusBar
}

func formatRelativeTime(t time.Time) string {
	d := time.Since(t)
	if d.Hours() < 24 {
		return "today"
	}
	days := int(d.Hours() / 24)
	if days < 30 {
		return fmt.Sprintf("%d days ago", days)
	}
	months := days / 30
	if months < 12 {
		return fmt.Sprintf("%d months ago", months)
	}
	years := months / 12
	return fmt.Sprintf("%d years ago", years)
}

func truncate(s string, l int) string {
	if len(s) <= l {
		return s
	}
	return s[:l-3] + "..."
}

func (m Model) confirmView() string {
	var b strings.Builder
	b.WriteString(ConfirmStyle.Render(fmt.Sprintf("Delete %d repositories?", len(m.selected))))
	b.WriteString("\n\n")

	count := 0
	for i := range m.selected {
		if count >= 10 {
			b.WriteString(fmt.Sprintf("...and %d more\n", len(m.selected)-10))
			break
		}
		b.WriteString(fmt.Sprintf("• %s\n", m.filtered[i].NameWithOwner))
		count++
	}

	b.WriteString("\n")
	b.WriteString(ErrorStyle.Render("⚠️ This action is IRREVERSIBLE"))
	b.WriteString("\n\n")
	b.WriteString("[y] Confirm    [n/Esc] Cancel")

	dialog := lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2).
			Render(b.String()),
	)
	return dialog
}

func (m Model) deletingView() string {
	var b strings.Builder
	total := len(m.deletingRepos)
	current := len(m.deleteResults)
	b.WriteString(TitleStyle.Render(fmt.Sprintf("Deleting repositories... (%d/%d)", current, total)))
	b.WriteString("\n\n")

	// Show results so far
	for _, res := range m.deleteResults {
		if res.Success {
			b.WriteString(SuccessStyle.Render(fmt.Sprintf("✓ %s (%s)", res.Repo.NameWithOwner, res.Duration.Round(time.Millisecond))))
		} else {
			b.WriteString(ErrorStyle.Render(fmt.Sprintf("✗ %s %v", res.Repo.NameWithOwner, res.Error)))
		}
		b.WriteString("\n")
	}

	// Show current and pending
	if current < total {
		b.WriteString(ProgressStyle.Render(fmt.Sprintf("○ %s deleting...", m.deletingRepos[current].NameWithOwner)))
		b.WriteString("\n")
		for i := current + 1; i < total; i++ {
			b.WriteString(fmt.Sprintf("○ %s waiting", m.deletingRepos[i].NameWithOwner))
			b.WriteString("\n")
		}
	} else {
		// Done
		successCount := 0
		for _, res := range m.deleteResults {
			if res.Success {
				successCount++
			}
		}
		b.WriteString("\n")
		b.WriteString(SuccessStyle.Render(fmt.Sprintf("%d succeeded, %d failed", successCount, total-successCount)))
		b.WriteString("\n\nPress Enter or q to continue")
	}

	return lipgloss.NewStyle().Padding(1, 2).Render(b.String())
}
