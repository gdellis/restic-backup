# AGENTS.md - Restic Backup Client

## Build Commands

```bash
# Build the application
go build -o restic-client ./cmd/restic-client

# Run the application (requires TTY)
go run ./cmd/restic-client

# Run tests (none yet, but structure is):
go test ./...

# Run a single test
go test -run TestName ./internal/config/...

# Lint with go vet
go vet ./...

# Format code
go fmt ./...

# Tidy dependencies
go mod tidy
```

## Project Overview

This is a Go-based TUI (Terminal User Interface) restic backup client using Bubble Tea. It wraps the restic CLI and provides an interactive interface for managing backups.

### Directory Structure

```
cmd/restic-client/main.go    # Application entry point
internal/
├── config/                  # Configuration management with age encryption
│   ├── config.go
│   └── encrypted.go
├── repository/             # Repository backend definitions
│   └── backends.go
├── restic/                 # Restic CLI wrapper
│   └── executor.go
├── ui/                     # TUI screens and models
│   ├── app.go
│   └── models.go
├── backup/                 # Backup operations (stubs)
└── notifications/          # Notification system (stubs)
```

## Code Style Guidelines

### General Principles
- Keep functions short and focused
- Use meaningful variable names
- Avoid global state where possible
- Return early for error cases
- Use Go's zero value where appropriate

### Imports
- Group imports: standard library, external packages, internal packages
- Use the project's modules: `restic-client/internal/...`
- Run `go fmt` before committing

```go
import (
    "context"
    "fmt"
    "os"

    "github.com/charmbracelet/lipgloss"
    "restic-client/internal/config"
)
```

### Naming Conventions
- **Variables**: `camelCase` (e.g., `repoPath`, `password`)
- **Constants**: `CamelCase` or `snake_case` (e.g., `DefaultBackend`, `max_retries`)
- **Types**: `CamelCase` (e.g., `ResticExecutor`, `Repository`)
- **Interfaces**: `CamelCase` with `er` suffix (e.g., `Reader`, `Executor`)
- **Packages**: `snake_case` (e.g., `config`, `restic`, `ui`)
- **Acronyms**: Keep original case (e.g., `URL`, `ID`, `HTML`)

### Error Handling
- Use sentinel errors for known conditions
- Wrap errors with context using `fmt.Errorf("...: %w", err)`
- Never silently ignore errors with `_`
- Return errors to callers, handle at appropriate level

```go
// Good
if err != nil {
    return nil, fmt.Errorf("failed to load config: %w", err)
}

// Sentinel errors
var (
    ErrNotFound  = errors.New("restic not found")
    ErrExecFailed = errors.New("restic command failed")
)
```

### Types
- Use specific types where possible (not `interface{}`)
- Use `time.Time` for timestamps
- Use `context.Context` for cancellation
- Prefer value receivers unless you need pointer receivers
- Use structs with tags for JSON serialization

### Functions
- Keep functions under 50 lines when possible
- Use functional options for configurable constructors
- Document exported functions with comments

```go
// Functional options pattern
type ResticOption func(*ResticExecutor)

func WithRepository(repo string) ResticOption {
    return func(e *ResticExecutor) {
        e.repo = repo
    }
}
```

### TUI/UI Conventions
- Use exported style variables from `ui` package (e.g., `TitleStyle`, `BorderStyle`)
- Use `lipgloss` for styling
- Implement `tea.Model` interface for TUI components
- Handle keyboard input in `Update` method
- Return view as string in `View` method

### Testing
- Place tests in `*_test.go` files in same package
- Use table-driven tests for multiple cases
- Test happy path and error cases
- Use `t.Helper()` for test utilities

### Configuration
- Config stored in `~/.config/restic-client/`
- Encrypted with age (X25519)
- Repository passwords stored encrypted or via environment variable reference

## Dependencies

- **Bubble Tea** (`github.com/charmbracelet/bubbletea`): TUI framework
- **Lipgloss** (`github.com/charmbracelet/lipgloss`): Terminal styling
- **Age** (`filippo.io/age`): Encryption for config storage

## Key Interfaces

```go
// tea.Model interface for all screens
type Model interface {
    Init() tea.Cmd
    Update(tea.Msg) (tea.Model, tea.Cmd)
    View() string
}
```

## Environment Variables

- `RESTIC_PASSWORD`: Password for repository
- `RESTIC_REPOSITORY`: Repository path/URL
- `RESTIC_CLIENT_PATH`: Path to restic binary (defaults to "restic")
- `XDG_CONFIG_HOME`: Config directory base (default: `~/.config`)

## Common Tasks

### Adding a New Screen
1. Add screen constant to `ui.Screen` in `internal/ui/app.go`
2. Create model struct in `internal/ui/models.go`
3. Add case in main `Update` and `View` methods
4. Add navigation key binding

### Adding a New Restic Command
1. Add method to `ResticExecutor` in `internal/restic/executor.go`
2. Use `ExecuteJSON` for JSON output or `Execute` for plain text
3. Add option functions if needed
4. Return parsed result or error
