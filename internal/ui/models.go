package ui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"restic-client/internal/config"
	"restic-client/internal/restic"
)

type DashboardModel struct {
	repos       []string
	selectedIdx int
	stats       map[string]RepoStats
}

type RepoStats struct {
	LastBackup   time.Time
	SnapshotCount int
	TotalSize   int64
	Status      string
}

func NewDashboardModel() DashboardModel {
	return DashboardModel{
		repos: []string{},
		stats: make(map[string]RepoStats),
	}
}

func (m DashboardModel) Update(msg tea.Msg) (DashboardModel, tea.Cmd) {
	return m, nil
}

func (m DashboardModel) View() string {
	header := TitleStyle.Render("Dashboard")
	
	menuItems := GetMenuItems(ScreenDashboard)
	menu := renderMenu(menuItems, 0)
	
	var repoList string
	for i, r := range m.repos {
		prefix := "  "
		if i == m.selectedIdx {
			prefix = "● "
		}
		repoList += prefix + r + "\n"
	}
	
	content := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		menu,
		"",
		BorderStyle.Render(repoList),
	)
	
	return content
}

func (m *DashboardModel) SetRepositories(repos []string) {
	m.repos = repos
}

func (m *DashboardModel) UpdateStats(repoID string, stats RepoStats) {
	m.stats[repoID] = stats
}

type ReposModel struct {
	repos       []string
	selectedIdx int
	editing     bool
	form        RepoForm
	formData    config.Repository
}

type RepoForm struct {
	Name        string
	Backend     string
	Path        string
	Password    string
	PasswordEnv string
	BackupPaths string
	Exclude     string
	Tags        string
}

func NewReposModel() ReposModel {
	return ReposModel{
		repos:   []string{},
		form:    RepoForm{},
	}
}

func (m ReposModel) Update(msg tea.Msg) (ReposModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			if m.selectedIdx < len(m.repos)-1 {
				m.selectedIdx++
			}
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "enter":
			m.editing = !m.editing
		case "n":
			m.editing = true
			m.form = RepoForm{}
		}
	}
	return m, nil
}

func (m ReposModel) View() string {
	header := TitleStyle.Render("Repositories")
	menuItems := GetMenuItems(ScreenRepositories)
	menu := renderMenu(menuItems, 1)
	
	var content string
	if m.editing {
		content = BorderStyle.Render(m.renderForm())
	} else {
		var repoList string
		for i, r := range m.repos {
			prefix := "  "
			if i == m.selectedIdx {
				prefix = "● "
			}
			repoList += prefix + r + "\n"
		}
		if repoList == "" {
			repoList = "  No repositories configured\n  Press 'n' to add one"
		}
		content = BorderStyle.Render(repoList)
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			menu,
			content,
		),
		"",
	)
}

func (m ReposModel) renderForm() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Add Repository"),
		"",
		"Name:    "+m.form.Name,
		"Backend: "+m.form.Backend,
		"Path:    "+m.form.Path,
		"Password: "+m.form.Password,
		"Password Env: "+m.form.PasswordEnv,
		"Backup Paths: "+m.form.BackupPaths,
		"Exclude: "+m.form.Exclude,
		"Tags: "+m.form.Tags,
	)
}

func (m *ReposModel) SetRepositories(repos []config.Repository) {
	m.repos = make([]string, len(repos))
	for i, r := range repos {
		m.repos[i] = r.Name + " (" + r.Backend + ")"
	}
}

func (m *ReposModel) GetSelectedRepo() string {
	if len(m.repos) == 0 {
		return ""
	}
	return m.repos[m.selectedIdx]
}

func (m ReposModel) GetFormData() config.Repository {
	repo := config.Repository{
		Name:        m.form.Name,
		Backend:     m.form.Backend,
		Path:        m.form.Path,
		Password:    m.form.Password,
		PasswordEnv: m.form.PasswordEnv,
	}
	
	if m.form.BackupPaths != "" {
		repo.BackupPaths = splitCommaList(m.form.BackupPaths)
	}
	
	if m.form.Exclude != "" {
		repo.Exclude = splitCommaList(m.form.Exclude)
	}
	
	if m.form.Tags != "" {
		repo.Tags = splitCommaList(m.form.Tags)
	}
	
	return repo
}

func (m *ReposModel) SetEditing(editing bool) {
	m.editing = editing
}

func (m *ReposModel) ToggleEditing() {
	m.editing = !m.editing
}

type BackupModel struct {
	repos       []string
	selectedIdx int
	inProgress  bool
	progress    string
	paths       string
	exclude     string
	tags        string
}

func NewBackupModel() BackupModel {
	return BackupModel{
		repos:   []string{},
		paths:   "",
		exclude: "",
		tags:    "",
	}
}

func (m BackupModel) Update(msg tea.Msg) (BackupModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			if m.selectedIdx < len(m.repos)-1 {
				m.selectedIdx++
			}
		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
		case "b":
			m.inProgress = true
			m.progress = "Starting backup..."
		}
	}
	
	if m.inProgress {
		m.progress = "Backup in progress..."
	}
	
	return m, nil
}

func (m BackupModel) View() string {
	header := TitleStyle.Render("Backup")
	menuItems := GetMenuItems(ScreenBackup)
	menu := renderMenu(menuItems, 2)
	
	var content string
	if m.inProgress {
		content = BorderStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				SuccessStyle.Render("Backup in progress..."),
				InfoStyle.Render(m.progress),
			),
		)
	} else {
		var repoList string
		for i, r := range m.repos {
			prefix := "  "
			if i == m.selectedIdx {
				prefix = "● "
			}
			repoList += prefix + r + "\n"
		}
		if repoList == "" {
			repoList = "  No repositories configured"
		}
		content = BorderStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Select Repository:"),
				repoList,
				"",
				lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Paths to Backup:"),
				"  "+m.paths,
				"",
				lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Exclude Patterns:"),
				"  "+m.exclude,
				"",
				lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Tags:"),
				"  "+m.tags,
			),
		)
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			menu,
			content,
		),
		"",
	)
}

func (m *BackupModel) SetRepositories(repos []config.Repository) {
	m.repos = make([]string, len(repos))
	for i, r := range repos {
		m.repos[i] = r.Name
	}
}

func (m *BackupModel) GetSelectedRepo() string {
	if len(m.repos) == 0 {
		return ""
	}
	return m.repos[m.selectedIdx]
}

func (m BackupModel) GetBackupPaths() []string {
	if m.paths == "" {
		return nil
	}
	return splitCommaList(m.paths)
}

func (m BackupModel) GetExcludePatterns() []string {
	if m.exclude == "" {
		return nil
	}
	return splitCommaList(m.exclude)
}

func (m BackupModel) GetTags() []string {
	if m.tags == "" {
		return nil
	}
	return splitCommaList(m.tags)
}

func (m *BackupModel) SetProgress(p string) {
	m.progress = p
}

func (m *BackupModel) Complete(success bool) {
	m.inProgress = false
	if success {
		m.progress = "Backup completed successfully!"
	} else {
		m.progress = "Backup failed!"
	}
}

type RestoreModel struct {
	repos         []string
	selectedRepoIdx int
	snapshots     []string
	selectedSnapIdx int
	target        string
	showSnapshots bool
}

func NewRestoreModel() RestoreModel {
	return RestoreModel{
		repos:         []string{},
		snapshots:     []string{},
		target:        "",
	}
}

func (m RestoreModel) Update(msg tea.Msg) (RestoreModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			if m.showSnapshots && m.selectedSnapIdx < len(m.snapshots)-1 {
				m.selectedSnapIdx++
			} else if !m.showSnapshots && m.selectedRepoIdx < len(m.repos)-1 {
				m.selectedRepoIdx++
			}
		case "up", "k":
			if m.showSnapshots && m.selectedSnapIdx > 0 {
				m.selectedSnapIdx--
			} else if !m.showSnapshots && m.selectedRepoIdx > 0 {
				m.selectedRepoIdx--
			}
		case "r":
			m.showSnapshots = true
		}
	}
	return m, nil
}

func (m RestoreModel) View() string {
	header := TitleStyle.Render("Restore")
	menuItems := GetMenuItems(ScreenRestore)
	menu := renderMenu(menuItems, 3)
	
	var content string
	if m.showSnapshots {
		var snapList string
		for i, s := range m.snapshots {
			prefix := "  "
			if i == m.selectedSnapIdx {
				prefix = "● "
			}
			snapList += prefix + s + "\n"
		}
		if snapList == "" {
			snapList = "  No snapshots found"
		}
		content = BorderStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Select Snapshot:"),
				snapList,
			),
		)
	} else {
		var repoList string
		for i, r := range m.repos {
			prefix := "  "
			if i == m.selectedRepoIdx {
				prefix = "● "
			}
			repoList += prefix + r + "\n"
		}
		if repoList == "" {
			repoList = "  No repositories configured"
		}
		content = BorderStyle.Render(
			lipgloss.JoinVertical(
				lipgloss.Left,
				lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Select Repository:"),
				repoList,
				"",
				lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Restore Target:"),
				"  "+m.target,
			),
		)
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			menu,
			content,
		),
		"",
	)
}

func (m *RestoreModel) SetRepositories(repos []config.Repository) {
	m.repos = make([]string, len(repos))
	for i, r := range repos {
		m.repos[i] = r.Name
	}
}

func (m *RestoreModel) SetSnapshots(snapshots []restic.Snapshot) {
	m.snapshots = make([]string, len(snapshots))
	for i, s := range snapshots {
		m.snapshots[i] = formatSnapshotItem(s)
	}
}

func formatSnapshotItem(s restic.Snapshot) string {
	return s.ID[:8] + " - " + s.Time.Format("2006-01-02 15:04") + " - " + s.Hostname
}

func (m RestoreModel) GetTarget() string {
	return m.target
}

func (m RestoreModel) GetSelectedSnapshot() string {
	if len(m.snapshots) == 0 {
		return ""
	}
	return m.snapshots[m.selectedSnapIdx]
}

type SnapshotsModel struct {
	repos           []string
	selectedRepoIdx int
	snapshots       []string
	selectedSnapIdx int
	hostFilter      string
	pathFilter      string
	tagFilter       string
}

func NewSnapshotsModel() SnapshotsModel {
	return SnapshotsModel{
		repos:   []string{},
		snapshots: []string{},
	}
}

func (m SnapshotsModel) Update(msg tea.Msg) (SnapshotsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			if m.selectedSnapIdx < len(m.snapshots)-1 {
				m.selectedSnapIdx++
			} else if m.selectedRepoIdx < len(m.repos)-1 {
				m.selectedRepoIdx++
			}
		case "up", "k":
			if m.selectedSnapIdx > 0 {
				m.selectedSnapIdx--
			} else if m.selectedRepoIdx > 0 {
				m.selectedRepoIdx--
			}
		}
	}
	return m, nil
}

func (m SnapshotsModel) View() string {
	header := TitleStyle.Render("Snapshots")
	menuItems := GetMenuItems(ScreenSnapshots)
	menu := renderMenu(menuItems, 4)
	
	var repoList string
	for i, r := range m.repos {
		prefix := "  "
		if i == m.selectedRepoIdx {
			prefix = "● "
		}
		repoList += prefix + r + "\n"
	}
	
	var snapList string
	for i, s := range m.snapshots {
		prefix := "  "
		if i == m.selectedSnapIdx {
			prefix = "● "
		}
		snapList += prefix + s + "\n"
	}
	if snapList == "" {
		snapList = "  No snapshots"
	}
	
	filters := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Filters:"),
		"Host: "+m.hostFilter,
		"Path: "+m.pathFilter,
		"Tag:  "+m.tagFilter,
	)
	
	content := BorderStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Repository:"),
			repoList,
			"",
			filters,
			"",
			lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Snapshots:"),
			snapList,
		),
	)
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			menu,
			content,
		),
		"",
	)
}

func (m *SnapshotsModel) SetRepositories(repos []config.Repository) {
	m.repos = make([]string, len(repos))
	for i, r := range repos {
		m.repos[i] = r.Name
	}
}

func (m *SnapshotsModel) SetSnapshots(snapshots []restic.Snapshot) {
	m.snapshots = make([]string, len(snapshots))
	for i, s := range snapshots {
		m.snapshots[i] = formatSnapshotItem(s)
	}
}

func (m SnapshotsModel) GetSelectedRepo() string {
	if len(m.repos) == 0 {
		return ""
	}
	return m.repos[m.selectedRepoIdx]
}

func (m SnapshotsModel) GetSelectedSnapshot() string {
	if len(m.snapshots) == 0 {
		return ""
	}
	return m.snapshots[m.selectedSnapIdx]
}

type SettingsModel struct {
	resticPath     string
	defaultBackend string
	themeIdx      int
	notifEnabled  bool
	showAdvanced  bool
	showNotif     bool
}

func NewSettingsModel() SettingsModel {
	return SettingsModel{
		resticPath:     "restic",
		defaultBackend: "local",
		themeIdx:      0,
		notifEnabled:  false,
	}
}

func (m SettingsModel) Update(msg tea.Msg) (SettingsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "down", "j":
			if m.showNotif {
				m.notifEnabled = !m.notifEnabled
			}
		case "t":
			m.showAdvanced = !m.showAdvanced
		case "n":
			m.showNotif = !m.showNotif
		}
	}
	return m, nil
}

func (m SettingsModel) View() string {
	header := TitleStyle.Render("Settings")
	menuItems := GetMenuItems(ScreenSettings)
	menu := renderMenu(menuItems, 5)
	
	var content string
	if m.showNotif {
		content = BorderStyle.Render(m.renderNotificationSettings())
	} else {
		content = BorderStyle.Render(m.renderGeneralSettings())
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		"",
		lipgloss.JoinHorizontal(
			lipgloss.Top,
			menu,
			content,
		),
		"",
	)
}

func (m SettingsModel) renderGeneralSettings() string {
	theme := "Dark"
	if m.themeIdx == 1 {
		theme = "Light"
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("General:"),
		"  Restic Binary: "+m.resticPath,
		"  Default Backend: "+m.defaultBackend,
		"  Theme: "+theme,
		"",
		InfoStyle.Render("Press t for advanced, n for notifications"),
	)
}

func (m SettingsModel) renderNotificationSettings() string {
	notifStatus := WarningStyle.Render("Disabled")
	if m.notifEnabled {
		notifStatus = SuccessStyle.Render("Enabled")
	}
	
	return lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.Style{}.Foreground(lipgloss.Color("99")).Bold(true).Render("Notifications:"),
		"  Status: "+notifStatus,
		"",
		InfoStyle.Render("Press j to toggle, Esc to go back"),
	)
}

func (m *SettingsModel) SetSettings(s *config.Settings) {
	if s != nil {
		m.resticPath = s.ResticPath
		m.defaultBackend = s.DefaultBackend
		if s.Theme == "light" {
			m.themeIdx = 1
		}
		if s.Notifications != nil {
			m.notifEnabled = s.Notifications.Enabled
		}
	}
}

func (m SettingsModel) GetSettings() *config.Settings {
	theme := "dark"
	if m.themeIdx == 1 {
		theme = "light"
	}
	
	return &config.Settings{
		ResticPath:     m.resticPath,
		DefaultBackend: m.defaultBackend,
		Theme:         theme,
		Notifications: &config.Notifications{
			Enabled: m.notifEnabled,
		},
	}
}

func (m *SettingsModel) ToggleNotifications() {
	m.notifEnabled = !m.notifEnabled
}

func (m SettingsModel) GetResticPath() string {
	return m.resticPath
}

func (m SettingsModel) GetTheme() string {
	if m.themeIdx == 1 {
		return "light"
	}
	return "dark"
}

func splitCommaList(s string) []string {
	if s == "" {
		return nil
	}
	var result []string
	for _, part := range splitAndTrimList(s, ",") {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitAndTrimList(s, sep string) []string {
	parts := make([]string, 0)
	for _, p := range splitList(s, sep) {
		parts = append(parts, trimList(p))
	}
	return parts
}

func splitList(s, sep string) []string {
	if s == "" {
		return nil
	}
	var result []string
	start := 0
	for i := 0; i <= len(s)-len(sep); i++ {
		if s[i:i+len(sep)] == sep {
			result = append(result, s[start:i])
			start = i + len(sep)
		}
	}
	result = append(result, s[start:])
	return result
}

func trimList(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}

func renderMenu(items []string, selected int) string {
	var lines []string
	for i, item := range items {
		if i == selected {
			lines = append(lines, SelectedItemStyle.Render(item))
		} else {
			lines = append(lines, NormalItemStyle.Render(item))
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return "Never"
	}
	return t.Format("2006-01-02 15:04")
}
