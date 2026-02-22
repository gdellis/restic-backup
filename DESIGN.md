# Restic Backup Client - Design Document

## Overview

A Go-based Terminal User Interface (TUI) client for restic backup, built with Bubble Tea. Provides an interactive interface for managing backups, repositories, snapshots, and retention policies.

## Architecture

### Package Structure

```mermaid
graph TB
    subgraph cmd/restic-client
        main[main.go]
    end
    
    subgraph internal
        subgraph ui
            app[app.go]
            models[models.go]
        end
        
        config[config/]
        restic[restic/]
        repository[repository/]
        notifications[notifications/]
    end
    
    main --> app
    app --> models
    app --> config
    app --> restic
    config --> repository
```

### Component Diagram

```mermaid
classDiagram
    class model {
        +config *Config
        +execMap map[string]*ResticExecutor
        +screen Screen
        +dashboard DashboardModel
        +backup BackupModel
        +restore RestoreModel
        +snapshots SnapshotsModel
        +retention RetentionModel
        +settings SettingsModel
        +Init() tea.Cmd
        +Update(tea.Msg) (tea.Model, tea.Cmd)
        +View() string
    }
    
    class BackupModel {
        +repos []string
        +selectedIdx int
        +inProgress bool
        +progress string
        +paths string
        +exclude string
        +tags string
        +Update(tea.Msg) (tea.Model, tea.Cmd)
        +View() string
        +SetRepositories([]Repository)
        +GetBackupPaths() []string
        +GetExcludePatterns() []string
        +GetTags() []string
        +IsRunning() bool
        +Cancel()
    }
    
    class ResticExecutor {
        -resticPath string
        -repo string
        -password string
        -lock sync.Mutex
        +Backup(ctx, paths, opts) (*BackupResult, error)
        +Restore(ctx, snapshotID, opts) error
        +Snapshots(ctx, filters) ([]Snapshot, error)
        +Forget(ctx, snapshotIDs, opts) error
    }
    
    model --> BackupModel
    model --> ResticExecutor
```

## Screen Navigation

```mermaid
stateDiagram-v2
    [*] --> Dashboard
    Dashboard --> Repositories
    Dashboard --> Backup
    Dashboard --> Restore
    Dashboard --> Snapshots
    Dashboard --> Retention
    Dashboard --> Settings
    
    Repositories --> Dashboard
    Backup --> Dashboard
    Restore --> Dashboard
    Snapshots --> Dashboard
    Retention --> Dashboard
    Settings --> Dashboard
```

## Backup Flow

```mermaid
sequenceDiagram
    participant User
    participant TUI as Bubble Tea
    participant Model
    participant BackupModel
    participant ResticExecutor
    participant ResticCLI as restic CLI
    
    User->>TUI: Press 'b' to start backup
    TUI->>Model: tea.KeyMsg
    Model->>Model: handleKey()
    
    alt repository selected and not running
        Model->>Model: startBackupAsync()
        Model->>BackupModel: Set inProgress=true
        Model->>Model: Returns tea.Cmd (goroutine)
        
        par Async execution
            Model->>ResticExecutor: Backup(ctx, paths, opts)
            ResticExecutor->>ResticCLI: restic backup --json [paths]
            ResticCLI-->>ResticExecutor: JSON output
            ResticExecutor-->>Model: BackupResult
        and UI Update
            TUI->>Model: Update()
            Model->>BackupModel: Update()
            BackupModel-->>TUI: View() with progress
        end
        
        alt success
            Model->>BackupModel: Complete(true)
            Model->>BackupModel: SetProgress(stats)
            User->>TUI: See success message
        else failure
            Model->>BackupModel: Complete(false)
            Model->>BackupModel: SetError(msg.error)
            User->>TUI: See error message
        end
    else backup already running
        User->>TUI: Press 'c' to cancel
        Model->>BackupModel: Cancel()
    end
```

## Data Models

### Config Structure

```mermaid
erDiagram
    Config ||--o| Repository : contains
    Config ||--o| Settings : contains
    Config ||--o| Identity : contains
    
    Repository ||--o| RetentionPolicy : has
    
    Settings ||--o| Notifications : has
    Notifications ||--o| SMTPConfig : has
    
    Repository {
        string id
        string name
        string backend
        string path
        string password
        string password_env
        string password_op
        list backup_paths
        list exclude
        list tags
    }
    
    RetentionPolicy {
        int keep_last
        int keep_daily
        int keep_weekly
        int keep_monthly
        int keep_yearly
    }
```

## Configuration Flow

```mermaid
flowchart LR
    A[Start] --> B{Config exists?}
    
    B -->|No| C[Create config dir]
    C --> D[Generate age identity]
    D --> E[Load recipient]
    E --> F[Create default config]
    F --> G[Encrypt & save]
    G --> H[Start TUI]
    
    B -->|Yes| I[Load identity]
    I --> J[Load recipient]
    J --> K[Decrypt config]
    K --> H
    
    H --> L[Display Dashboard]
```

## Key Bindings

| Key | Action |
|-----|--------|
| 1 | Navigate to Dashboard |
| 2 | Navigate to Repositories |
| 3 | Navigate to Backup |
| 4 | Navigate to Restore |
| 5 | Navigate to Snapshots |
| 6 | Navigate to Retention |
| 7 | Navigate to Settings |
| b | Start backup (on Backup screen) |
| c | Cancel backup (when running) |
| n | Add new repository |
| q | Quit |
| ? | Show help |

## Environment Variables

| Variable | Description |
|----------|-------------|
| RESTIC_PASSWORD | Password for repository |
| RESTIC_REPOSITORY | Repository path/URL |
| RESTIC_CLIENT_PATH | Path to restic binary |
| XDG_CONFIG_HOME | Config directory base |

## Dependencies

- **Bubble Tea** - TUI framework
- **Lipgloss** - Terminal styling
- **Age** - Encryption for config storage

## Security

- Config stored in `~/.config/restic-client/`
- Encrypted with age (X25519)
- Repository passwords stored encrypted or via environment variable
- 1Password integration for password retrieval
