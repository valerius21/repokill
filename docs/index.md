# repokill

A terminal user interface tool for bulk-deleting GitHub repositories.

## Overview

repokill provides an interactive terminal interface for managing and deleting multiple GitHub repositories at once. It uses the GitHub CLI (`gh`) for authentication and API access, ensuring secure and reliable operations.

## Features

- **Interactive TUI**: Browse and select repositories using keyboard navigation
- **Filtering**: Filter repositories by visibility, archived status, or fork status
- **Search**: Real-time search by repository name or description
- **Bulk Operations**: Select and delete, archive, or backup multiple repositories at once
- **Visibility Change**: Make repositories private or public in bulk
- **Done Sorting**: Repositories marked as done are sorted to the top (highlighted green)
- **Backup**: Export repositories as ZIP archives or shallow clones
- **Rate Limit Handling**: Automatic retry with exponential backoff for rate limits
- **Progress Tracking**: Visual feedback during deletion, archive, and backup operations

## Installation

### Binary Release

Download the appropriate binary for your platform from the [Releases page](https://github.com/valerius21/repokill/releases).

### Go Install

```bash
go install github.com/valerius21/repokill@latest
```

### Build from Source

```bash
git clone https://github.com/valerius21/repokill.git
cd repokill
make build
```

## Requirements

- [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated
- `delete_repo` scope enabled (run `gh auth refresh -s delete_repo` if needed)

## Usage

### Basic Usage

```bash
repokill                    # List repos for authenticated user
repokill --org myorg        # List repos for an organization
repokill --version          # Show version
```

### Filtering Options

```bash
repokill --public           # Show only public repos
repokill --private          # Show only private repos
repokill --archived         # Show only archived repos
repokill --forked           # Show only forks
```

### Key Bindings

| Key | Action |
|-----|--------|
| `j`/`↓` | Move down |
| `k`/`↑` | Move up |
| `space` | Toggle mark for deletion |
| `a` | Select/deselect all |
| `/` | Search repositories |
| `enter` | Confirm deletion |
| `A` | Archive selected repositories |
| `p` | Make selected repositories private |
| `P` | Make selected repositories public |
| `B` | Backup selected repositories |
| `u`/`pgup` | Page up |
| `d`/`pgdown` | Page down |
| `g`/`home` | Go to top |
| `G`/`end` | Go to bottom |
| `q`/`ctrl+c` | Quit |

## API Reference

### Package github

The `github` package provides a client for interacting with GitHub via the gh CLI.

#### type Client

```go
client := github.NewClient(owner string, executor CommandExecutor) *Client
```

Creates a new GitHub client for the specified owner. If owner is empty, operations target the authenticated user.

#### type Repo

```go
type Repo struct {
    Name           string
    NameWithOwner  string
    Description    string
    PushedAt       time.Time
    Visibility     string
    IsArchived     bool
    IsFork         bool
    StargazerCount int
    ForkCount      int
}
```

### Package filter

The `filter` package provides repository filtering and sorting functionality.

#### FilterOptions

```go
type FilterOptions struct {
    Archived    *bool
    Forked      *bool
    Visibility  string
    SearchQuery string
}
```

#### SortOptions

```go
type SortOptions struct {
    Field SortField  // SortByPushedAt, SortByName, SortByStars
    Order SortOrder  // Ascending, Descending
}
```

## Error Handling

The client handles several error scenarios:

- `ErrGhNotInstalled`: GitHub CLI is not installed
- `ErrNotAuthenticated`: User is not logged in to GitHub CLI
- `ErrMissingScope`: Missing required `delete_repo` scope

During deletion, the following errors are handled:

- **Rate limits (HTTP 429)**: Automatic retry with exponential backoff
- **Permission denied (HTTP 403)**: Reported as failure
- **Not found (HTTP 404)**: Repository already deleted or moved
- **Renamed/transferred**: Repository name changed


## Operations

### Archive

The archive operation marks repositories as archived on GitHub. Archived repositories become read-only and cannot be modified. This is useful for:

- Preserving repository history without deletion
- Marking projects as completed or inactive
- Reducing clutter while maintaining access to code

Press `A` to archive all selected repositories.

### Visibility Change

You can change repository visibility in bulk:

- **`p`** - Make selected repositories **private** (only you can see them)
- **`P`** - Make selected repositories **public** (visible to everyone)

This is useful for managing privacy settings across multiple repositories at once.

### Done/Green Sorting

Repositories that have been processed (deleted, archived, or had visibility changed) are marked as "done" and sorted to the top of the list. They appear highlighted in green to indicate they've been handled.

This helps you:
- Track which repositories you've already acted on
- Focus on remaining repositories that need attention
- Avoid accidentally re-processing the same repositories

### Backup

Press `B` to backup selected repositories. The backup operation supports two modes:

- **ZIP Archive**: Downloads the repository contents as a ZIP file
- **Shallow Clone**: Creates a shallow git clone of the repository

Backups are saved to a `backups/` directory in your current working directory. This provides a safety net before deletion or archiving operations.

## License

[GPL-2.0](https://www.gnu.org/licenses/old-licenses/gpl-2.0.html)
