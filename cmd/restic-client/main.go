package main

import (
	"context"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"restic-client/internal/config"
	"restic-client/internal/restic"
	"restic-client/internal/ui"
)

type model struct {
	config  *config.Config
	execMap map[string]*restic.ResticExecutor
	screen  ui.Screen
	err     error
	width   int
	height  int

	dashboard ui.DashboardModel
	repos     ui.ReposModel
	backup    ui.BackupModel
	restore   ui.RestoreModel
	snapshots ui.SnapshotsModel
	retention ui.RetentionModel
	settings  ui.SettingsModel

	showHelp     bool
	backupCtx    context.Context
	backupCancel context.CancelFunc
}

func newModel() *model {
	return &model{
		screen:    ui.ScreenDashboard,
		execMap:   make(map[string]*restic.ResticExecutor),
		dashboard: ui.NewDashboardModel(),
		repos:     ui.NewReposModel(),
		backup:    ui.NewBackupModel(),
		restore:   ui.NewRestoreModel(),
		snapshots: ui.NewSnapshotsModel(),
		retention: ui.NewRetentionModel(),
		settings:  ui.NewSettingsModel(),
	}
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.loadConfig,
	)
}

func (m *model) loadConfig() tea.Msg {
	identity, err := config.LoadIdentity()
	if err != nil {
		return errorMsg{fmt.Errorf("failed to load identity: %w", err)}
	}

	storage := config.NewEncryptedStorage(identity)
	recipient, err := config.LoadRecipient()
	if err != nil {
		return errorMsg{fmt.Errorf("failed to load recipient: %w", err)}
	}
	storage.SetRecipient(recipient)

	cfg, err := storage.LoadAndDecrypt()
	if err != nil {
		return errorMsg{fmt.Errorf("failed to load config: %w", err)}
	}

	return configLoadedMsg{config: cfg}
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case configLoadedMsg:
		m.config = msg.config
		m.repos.SetRepositories(m.config.Repositories)
		m.backup.SetRepositories(m.config.Repositories)
		m.restore.SetRepositories(m.config.Repositories)
		m.snapshots.SetRepositories(m.config.Repositories)
		m.retention.SetRepositories(m.config.Repositories)
		if m.config.Settings != nil {
			m.settings.SetSettings(m.config.Settings)
		}
		return m, nil

	case errorMsg:
		m.err = msg.err
		return m, nil

	case backupResultMsg:
		if m.backupCancel != nil {
			m.backupCancel()
			m.backupCancel = nil
		}
		if msg.success {
			m.backup.Complete(true)
			m.backup.SetProgress(fmt.Sprintf("Backup complete: %s (Files: %d, Added: %s)",
				msg.snapshotID[:8], msg.stats.FilesProcessed, formatBytes(msg.stats.BytesAdded)))
		} else {
			m.backup.Complete(false)
			m.backup.SetError(msg.error)
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "?":
			m.showHelp = !m.showHelp
		}

		if m.showHelp {
			m.showHelp = false
			return m, nil
		}
		m.handleKey(msg)
	}

	var cmd tea.Cmd
	switch m.screen {
	case ui.ScreenDashboard:
		m.dashboard, cmd = m.dashboard.Update(msg)
	case ui.ScreenRepositories:
		m.repos, cmd = m.repos.Update(msg)
	case ui.ScreenBackup:
		m.backup, cmd = m.backup.Update(msg)
	case ui.ScreenRestore:
		m.restore, cmd = m.restore.Update(msg)
	case ui.ScreenSnapshots:
		m.snapshots, cmd = m.snapshots.Update(msg)
	case ui.ScreenRetention:
		m.retention, cmd = m.retention.Update(msg)
	case ui.ScreenSettings:
		m.settings, cmd = m.settings.Update(msg)
	}

	return m, cmd
}

func (m *model) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch msg.String() {
	case "1":
		m.screen = ui.ScreenDashboard
	case "2":
		m.screen = ui.ScreenRepositories
	case "3":
		m.screen = ui.ScreenBackup
	case "4":
		m.screen = ui.ScreenRestore
	case "5":
		m.screen = ui.ScreenSnapshots
	case "6":
		m.screen = ui.ScreenRetention
	case "7":
		m.screen = ui.ScreenSettings
	case "n":
		if m.screen == ui.ScreenRepositories {
			m.repos.SetEditing(true)
		}
	case "b":
		if m.screen == ui.ScreenBackup && m.backup.GetSelectedRepo() != "" && !m.backup.IsRunning() {
			return m.startBackupAsync()
		}
	case "c":
		if m.screen == ui.ScreenBackup && m.backup.IsRunning() {
			m.backup.Cancel()
		}
	}
	return nil
}

func (m *model) handleHelpKey(msg tea.KeyMsg) tea.Cmd {
	m.showHelp = false
	return nil
}

func (m *model) startBackupAsync() tea.Cmd {
	repoName := m.backup.GetSelectedRepo()
	repo := m.findRepoByName(repoName)
	if repo == nil {
		m.backup.SetError(fmt.Sprintf("repository %q not found", repoName))
		return nil
	}

	exec, err := m.getExecutor(repo)
	if err != nil {
		m.backup.SetError(err.Error())
		return nil
	}

	m.backupCtx, m.backupCancel = context.WithCancel(context.Background())
	paths := m.backup.GetBackupPaths()
	if len(paths) == 0 && len(repo.BackupPaths) > 0 {
		paths = repo.BackupPaths
	}

	opts := []restic.BackupOption{
		restic.BackupWithExclude(m.backup.GetExcludePatterns()),
		restic.BackupWithTags(m.backup.GetTags()),
	}

	return func() tea.Msg {
		result, err := exec.Backup(m.backupCtx, paths, opts...)
		if err != nil {
			return backupResultMsg{success: false, error: err.Error()}
		}
		return backupResultMsg{success: true, snapshotID: result.SnapshotID, stats: result.Stats}
	}
}

type backupResultMsg struct {
	success    bool
	snapshotID string
	stats      restic.BackupStats
	error      string
}

func (m *model) getExecutor(repo *config.Repository) (*restic.ResticExecutor, error) {
	if exec, ok := m.execMap[repo.ID]; ok {
		return exec, nil
	}

	password, err := repo.GetPassword()
	if err != nil {
		return nil, err
	}

	exec, err := restic.NewResticExecutor(
		restic.WithRepository(repo.Path),
		restic.WithPassword(password),
	)
	if err != nil {
		return nil, err
	}

	m.execMap[repo.ID] = exec
	return exec, nil
}

func (m *model) findRepoByName(name string) *config.Repository {
	for i := range m.config.Repositories {
		if m.config.Repositories[i].Name == name {
			return &m.config.Repositories[i]
		}
	}
	return nil
}

func (m *model) View() string {
	if m.err != nil {
		return ui.ErrorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	var content string
	switch m.screen {
	case ui.ScreenDashboard:
		content = m.dashboard.View()
	case ui.ScreenRepositories:
		content = m.repos.View()
	case ui.ScreenBackup:
		content = m.backup.View()
	case ui.ScreenRestore:
		content = m.restore.View()
	case ui.ScreenSnapshots:
		content = m.snapshots.View()
	case ui.ScreenRetention:
		content = m.retention.View()
	case ui.ScreenSettings:
		content = m.settings.View()
	}

	help := ui.HelpStyle.Render("1:Dash 2:Repos 3:Backup 4:Restore 5:Snaps 6:Retent 7:Settings q:Quit ?:Help")

	layout := lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		lipgloss.JoinVertical(
			lipgloss.Center,
			content,
			"",
			help,
		),
	)

	return layout
}

type errorMsg struct {
	err error
}

type configLoadedMsg struct {
	config *config.Config
}

func formatBytes(bytes int64) string {
	const unit = int64(1024)
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	divisor, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		divisor *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(divisor), "KMGTPE"[exp])
}

func main() {
	identPath := config.GetIdentityPath()

	if _, err := os.Stat(identPath); os.IsNotExist(err) {
		if err := config.EnsureConfigDir(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to create config dir: %v\n", err)
			os.Exit(1)
		}

		ident, err := config.LoadIdentity()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to generate identity: %v\n", err)
			os.Exit(1)
		}

		defaultConfig := &config.Config{
			Version:      1,
			Repositories: []config.Repository{},
			Settings: &config.Settings{
				Theme: "dark",
			},
		}

		storage := config.NewEncryptedStorage(ident)
		recipient, err := config.LoadRecipient()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to load recipient: %v\n", err)
			os.Exit(1)
		}
		storage.SetRecipient(recipient)

		if err := storage.EncryptAndSave(defaultConfig); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to save config: %v\n", err)
			os.Exit(1)
		}
	}

	m := newModel()

	p := tea.NewProgram(m)
	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
