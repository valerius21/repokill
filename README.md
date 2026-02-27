# repokill

[![Release](https://img.shields.io/github/v/release/valerius21/repokill?include_prereleases)](https://github.com/valerius21/repokill/releases)
[![Go Report Card](https://goreportcard.com/badge/github.com/valerius21/repokill)](https://goreportcard.com/report/github.com/valerius21/repokill)
[![License](https://img.shields.io/github/license/valerius21/repokill)](LICENSE)
[![Go Version](https://img.shields.io/github/go-mod/go-version/valerius21/repokill)](go.mod)

A TUI tool for bulk-managing GitHub repositories.

## Features

- List and browse repositories for authenticated user or organization
- Filter by visibility (public/private), archived status, or forks
- Search repos by name or description
- Bulk select and delete or archive multiple repositories
- Rate limit handling with automatic retries and exponential backoff
- Progress tracking during deletion and archive operations
- Bulk visibility change (make repos private/public)
- Done/green sorting for processed repositories
- Backup repositories as ZIP or shallow clone

## Requirements

- [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated with `delete_repo` scope

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

## Usage

```bash
repokill                    # List repos for authenticated user
repokill --org myorg        # List repos for an organization
repokill --public           # Show only public repos
repokill --private          # Show only private repos
repokill --archived         # Show only archived repos
repokill --forked           # Show only forks
repokill --version          # Show version
```

### Key Bindings

| Key          | Action                      |
| ------------ | --------------------------- |
| `j`/`↓`      | Move down                   |
| `k`/`↑`      | Move up                     |
| `space`      | Toggle mark                 |
| `a`          | Select/deselect all         |
| `/`          | Search                      |
| `enter`      | Confirm deletion            |
| `A`          | Archive selected repos      |
| `p`          | Make selected repos private |
| `P`          | Make selected repos public  |
| `B`          | Backup selected repos       |
| `u`/`pgup`   | Page up                     |
| `d`/`pgdown` | Page down                   |
| `g`/`home`   | Go to top                   |
| `G`/`end`    | Go to bottom                |
| `q`/`ctrl+c` | Quit                        |

## License

[GPL-2.0](https://www.gnu.org/licenses/old-licenses/gpl-2.0.html)
