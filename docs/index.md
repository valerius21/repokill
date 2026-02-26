# repokill

A terminal user interface tool for bulk-deleting GitHub repositories.

## Overview

repokill provides an interactive terminal interface for managing and deleting multiple GitHub repositories at once. It uses the GitHub CLI (`gh`) for authentication and API access, ensuring secure and reliable operations.

## Features

- **Interactive TUI**: Browse and select repositories using keyboard navigation
- **Filtering**: Filter repositories by visibility, archived status, or fork status
- **Search**: Real-time search by repository name or description
- **Bulk Operations**: Select and delete multiple repositories at once
- **Rate Limit Handling**: Automatic retry with exponential backoff for rate limits
- **Progress Tracking**: Visual feedback during deletion operations

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

## License

[GPL-2.0](https://www.gnu.org/licenses/old-licenses/gpl-2.0.html)
