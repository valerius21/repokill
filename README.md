# repokill

A TUI tool for bulk-deleting GitHub repositories.

## Requirements

- [GitHub CLI (`gh`)](https://cli.github.com/) installed and authenticated with `delete_repo` scope

## Installation

```bash
go install github.com/valerius21/repokill@latest
```

Or build from source:

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
```

## License

[GPL-2.0](https://www.gnu.org/licenses/old-licenses/gpl-2.0.html)
