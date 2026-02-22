# Restic Backup Client

[![OpenCode Review](https://github.com/gdellis/restic-backup/actions/workflows/pr_review.yml/badge.svg)](https://github.com/gdellis/restic-backup/actions/workflows/pr_review.yml)

A Go-based TUI (Terminal User Interface) restic backup client built with Bubble Tea. Provides an interactive interface for managing backups across multiple repository backends.

## Features

- **Interactive TUI**: Navigate between Dashboard, Repositories, Backup, Restore, Snapshots, Retention, and Settings screens
- **Multiple Backend Support**: Local, SFTP, S3, MinIO, Wasabi, Backblaze B2, Azure, Google Cloud Storage, OpenStack Swift
- **Encrypted Config**: Repository credentials encrypted with age (X25519)
- **1Password Integration**: Fetch passwords from 1Password CLI
- **Retention Policies**: Configure keep-last/daily/weekly/monthly/yearly snapshot policies
- **Notifications**: Webhook notifications on backup success/failure

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/restic-backup.git
cd restic-backup

# Build
go build -o restic-client ./cmd/restic-client

# Run
./restic-client
```

## Requirements

- Go 1.21+
- [restic](https://restic.net/) CLI binary in PATH
- Terminal emulator (TTY required)

## Configuration

On first run, the client creates:
- `~/.config/restic-client/identity.txt` - Age encryption identity
- `~/.config/restic-client/config.json` - Encrypted repository configurations

### Adding a Repository

1. Press `2` to go to Repositories
2. Press `n` to add new repository
3. Enter details:
   - Name: Friendly name
   - Backend: local/sftp/s3/b2/azure/gcs/swift
   - Path: Repository path or URL
   - Password: Encryption password (or use env/1Password)
   - Backup Paths: Comma-separated paths to backup

### Password Sources

The client supports three password sources (in priority order):
1. **Environment Variable**: Set `password_env` to an env var name
2. **1Password**: Set `password_op` to item name (requires 1Password CLI)
3. **Encrypted File**: Stored directly in encrypted config

## Usage

### Navigation

| Key | Action |
|-----|--------|
| `1-7` | Switch screens |
| `↑/↓` | Navigate lists |
| `Enter` | Select/confirm |
| `n` | New repository |
| `b` | Start backup |
| `?` | Help |
| `q` | Quit |

### Screens

- **Dashboard**: Overview of repositories and backup status
- **Repositories**: Manage repository configurations
- **Backup**: Run backup operations
- **Restore**: Restore from snapshots
- **Snapshots**: View and filter snapshots
- **Retention**: Configure snapshot retention policies
- **Settings**: Configure application settings

## Environment Variables

- `RESTIC_PASSWORD`: Default repository password
- `RESTIC_REPOSITORY`: Default repository path
- `RESTIC_CLIENT_PATH`: Path to restic binary (default: "restic")
- `XDG_CONFIG_HOME`: Config directory (default: `~/.config`)

## Project Structure

```
cmd/restic-client/main.go    # Application entry point
internal/
├── config/                  # Configuration with age encryption
│   ├── config.go
│   ├── encrypted.go
│   └── onepassword.go
├── repository/             # Backend definitions
│   └── backends.go
├── restic/                 # CLI wrapper
│   └── executor.go
├── ui/                    # TUI components
│   ├── app.go
│   └── models.go
├── backup/                # Backup operations
└── notifications/         # Notification system
```

## License

MIT
