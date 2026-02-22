package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/lipgloss"
)

type Screen int

const (
	ScreenDashboard Screen = iota
	ScreenRepositories
	ScreenBackup
	ScreenRestore
	ScreenSnapshots
	ScreenRetention
	ScreenSettings
)

var (
	HelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Padding(0, 1)

	TitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("99")).
			Bold(true).
			Padding(0, 1)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("212")).
				Background(lipgloss.Color("236")).
				Padding(0, 1)

	NormalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			Padding(0, 1)

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("42")).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("226")).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("75"))

	BorderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238"))
)

type Navigation struct {
	Current  Screen
	History  []Screen
}

func (n *Navigation) Push(screen Screen) {
	n.History = append(n.History, n.Current)
	n.Current = screen
}

func (n *Navigation) Pop() bool {
	if len(n.History) == 0 {
		return false
	}
	n.Current = n.History[len(n.History)-1]
	n.History = n.History[:len(n.History)-1]
	return true
}

func (n *Navigation) Reset() {
	n.Current = ScreenDashboard
	n.History = []Screen{}
}

type KeyMap struct {
	Up        key.Binding
	Down      key.Binding
	Left      key.Binding
	Right     key.Binding
	Enter     key.Binding
	Escape    key.Binding
	Quit      key.Binding
	Help      key.Binding
	Tab       key.Binding
	BackTab   key.Binding
	Delete    key.Binding
	Refresh   key.Binding
}

var DefaultKeyMap = KeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←/h", "back"),
	),
	Right: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→/l", "forward"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next"),
	),
	BackTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev"),
	),
	Delete: key.NewBinding(
		key.WithKeys("d", "delete"),
		key.WithHelp("d", "delete"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}

func ScreenTitle(screen Screen) string {
	switch screen {
	case ScreenDashboard:
		return "Dashboard"
	case ScreenRepositories:
		return "Repositories"
	case ScreenBackup:
		return "Backup"
	case ScreenRestore:
		return "Restore"
	case ScreenSnapshots:
		return "Snapshots"
	case ScreenSettings:
		return "Settings"
	default:
		return "Unknown"
	}
}

func ScreenDescription(screen Screen) string {
	switch screen {
	case ScreenDashboard:
		return "Overview of all repositories and backup status"
	case ScreenRepositories:
		return "Manage your restic repositories"
	case ScreenBackup:
		return "Run backup operations"
	case ScreenRestore:
		return "Restore from snapshots"
	case ScreenSnapshots:
		return "View and manage snapshots"
	case ScreenSettings:
		return "Configure application settings"
	default:
		return ""
	}
}

func GetMenuItems(active Screen) []string {
	return []string{
		fmt.Sprintf("%s Dashboard", getIcon(ScreenDashboard, active)),
		fmt.Sprintf("%s Repositories", getIcon(ScreenRepositories, active)),
		fmt.Sprintf("%s Backup", getIcon(ScreenBackup, active)),
		fmt.Sprintf("%s Restore", getIcon(ScreenRestore, active)),
		fmt.Sprintf("%s Snapshots", getIcon(ScreenSnapshots, active)),
		fmt.Sprintf("%s Retention", getIcon(ScreenRetention, active)),
		fmt.Sprintf("%s Settings", getIcon(ScreenSettings, active)),
	}
}

func getIcon(screen, active Screen) string {
	if screen == active {
		return "●"
	}
	return "○"
}
