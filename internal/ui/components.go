package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type TextInputModel struct {
	input     textinput.Model
	label     string
	help      string
	focused   bool
	validator func(string) error
	err       error
	prevValue string
}

func NewTextInput(label, placeholder string) TextInputModel {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.Prompt = "> "
	ti.Focus()

	return TextInputModel{
		input:     ti,
		label:     label,
		help:      "",
		focused:   true,
		prevValue: "",
	}
}

func (m TextInputModel) Update(msg tea.Msg) (TextInputModel, tea.Cmd) {
	var cmds []tea.Cmd

	currentValue := m.input.Value()

	ti, cmd := m.input.Update(msg)
	m.input = ti
	cmds = append(cmds, cmd)

	newValue := m.input.Value()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		keyStr := msg.String()
		if keyStr == "esc" && m.focused {
			m.focused = false
			m.input.Blur()
			m.err = nil
		}
		if keyStr == "enter" && m.focused {
			if m.validator != nil {
				m.err = m.validator(m.input.Value())
			}
		}
	default:
		if m.focused && m.validator != nil && newValue != currentValue {
			m.err = m.validator(newValue)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m TextInputModel) View() string {
	var view string
	if m.err != nil {
		view = ErrorStyle.Render(m.input.View()) + "\n" + ErrorStyle.Render(m.err.Error())
	} else {
		view = m.input.View()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		LabelStyle.Render(m.label),
		view,
	)
}

func (m *TextInputModel) SetValue(value string) {
	m.input.SetValue(value)
}

func (m *TextInputModel) Value() string {
	return m.input.Value()
}

func (m *TextInputModel) SetValidator(fn func(string) error) {
	m.validator = fn
}

func (m *TextInputModel) Focus() {
	m.focused = true
	m.input.Focus()
}

func (m *TextInputModel) Blur() {
	m.focused = false
	m.input.Blur()
}

func (m *TextInputModel) IsValid() bool {
	if m.validator == nil {
		return true
	}
	return m.validator(m.input.Value()) == nil
}

type PasswordInputModel struct {
	input     textinput.Model
	label     string
	focused   bool
	validator func(string) error
	err       error
}

func NewPasswordInput(label string) PasswordInputModel {
	ti := textinput.New()
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Prompt = "• "
	ti.Focus()

	return PasswordInputModel{
		input:   ti,
		label:   label,
		focused: true,
	}
}

func (m PasswordInputModel) Update(msg tea.Msg) (PasswordInputModel, tea.Cmd) {
	currentValue := m.input.Value()

	ti, cmd := m.input.Update(msg)
	m.input = ti

	newValue := m.input.Value()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.String() == "esc" && m.focused {
			m.focused = false
			m.input.Blur()
			m.err = nil
		}
		if msg.String() == "enter" && m.focused {
			if m.validator != nil {
				m.err = m.validator(newValue)
			}
		}
	default:
		if m.focused && m.validator != nil && newValue != currentValue {
			m.err = m.validator(newValue)
		}
	}

	return m, cmd
}

func (m PasswordInputModel) View() string {
	var view string
	if m.err != nil {
		view = ErrorStyle.Render(m.input.View()) + "\n" + ErrorStyle.Render(m.err.Error())
	} else {
		view = m.input.View()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		LabelStyle.Render(m.label),
		view,
	)
}

func (m *PasswordInputModel) SetValue(value string) {
	m.input.SetValue(value)
}

func (m PasswordInputModel) Value() string {
	return m.input.Value()
}

func (m *PasswordInputModel) SetValidator(fn func(string) error) {
	m.validator = fn
}

func (m PasswordInputModel) IsValid() bool {
	if m.validator == nil {
		return true
	}
	return m.validator(m.input.Value()) == nil
}

func (m *PasswordInputModel) Focus() {
	m.focused = true
	m.input.Focus()
}

func (m *PasswordInputModel) Blur() {
	m.focused = false
	m.input.Blur()
}

type ProgressBarModel struct {
	percent   float64
	width     int
	label     string
	showValue bool
}

func NewProgressBar(width int) ProgressBarModel {
	return ProgressBarModel{
		percent:   0,
		width:     width,
		label:     "",
		showValue: true,
	}
}

func (m *ProgressBarModel) SetPercent(p float64) {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	m.percent = p
}

func (m *ProgressBarModel) SetLabel(label string) {
	m.label = label
}

func (m *ProgressBarModel) View() string {
	if m.width <= 0 {
		m.width = 40
	}

	filled := int(float64(m.width) * m.percent / 100)
	empty := m.width - filled

	fillStr := string(rune(32))
	filledBar := strings.Repeat(fillStr, filled)
	emptyBar := strings.Repeat(fillStr, empty)

	bar := ProgressFillStyle.Render(filledBar) + ProgressEmptyStyle.Render(emptyBar)

	result := fmt.Sprintf("[%s]", bar)

	if m.showValue {
		result += fmt.Sprintf(" %.1f%%", m.percent)
	}

	if m.label != "" {
		result = lipgloss.JoinHorizontal(lipgloss.Left,
			LabelStyle.Render(m.label+":"),
			" ",
			result,
		)
	}

	return result
}

type SpinnerModel struct {
	spinner spinner.Model
	label   string
	active  bool
}

func NewSpinner(label string) SpinnerModel {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return SpinnerModel{
		spinner: s,
		label:   label,
		active:  true,
	}
}

func (m SpinnerModel) Update(msg tea.Msg) (SpinnerModel, tea.Cmd) {
	if !m.active {
		return m, nil
	}
	s, cmd := m.spinner.Update(msg)
	m.spinner = s
	return m, cmd
}

func (m SpinnerModel) View() string {
	if !m.active {
		return ""
	}

	return lipgloss.JoinHorizontal(
		lipgloss.Left,
		m.spinner.View(),
		" ",
		InfoStyle.Render(m.label),
	)
}

func (m *SpinnerModel) SetLabel(label string) {
	m.label = label
}

func (m *SpinnerModel) Start() {
	m.active = true
}

func (m *SpinnerModel) Stop() {
	m.active = false
}

func (m SpinnerModel) IsActive() bool {
	return m.active
}

type CheckboxModel struct {
	label     string
	checked   bool
	focused   bool
	shortcut  string
	shortcut2 string
}

func NewCheckbox(label string) CheckboxModel {
	return CheckboxModel{
		label:     label,
		checked:   false,
		focused:   false,
		shortcut:  "space",
		shortcut2: "x",
	}
}

func (m CheckboxModel) Update(msg tea.Msg) (CheckboxModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focused {
			keyStr := msg.String()
			if keyStr == m.shortcut || keyStr == m.shortcut2 {
				m.checked = !m.checked
			}
		}
	}
	return m, nil
}

func (m CheckboxModel) View() string {
	checkMark := "[ ]"
	shortcutHint := ""

	if m.checked {
		checkMark = CheckboxCheckedStyle.Render("[✓]")
	} else {
		checkMark = CheckboxUncheckedStyle.Render("[ ]")
	}

	if m.focused {
		shortcutHint = " (" + m.shortcut + ")"
	}

	checkStyle := NormalItemStyle
	if m.focused {
		checkStyle = SelectedItemStyle
	}

	return checkStyle.Render(checkMark + " " + m.label + shortcutHint)
}

func (m *CheckboxModel) SetChecked(checked bool) {
	m.checked = checked
}

func (m *CheckboxModel) Toggle() {
	m.checked = !m.checked
}

func (m CheckboxModel) IsChecked() bool {
	return m.checked
}

func (m *CheckboxModel) Focus() {
	m.focused = true
}

func (m *CheckboxModel) Blur() {
	m.focused = false
}

type SelectOption struct {
	Label       string
	Value       string
	Description string
}

type SelectModel struct {
	options  []SelectOption
	selected int
	filtered []SelectOption
	focused  bool
	expanded bool
	scroll   int
	width    int
	showDesc bool
}

func NewSelect(options []SelectOption, width int) SelectModel {
	return SelectModel{
		options:  options,
		filtered: options,
		selected: 0,
		scroll:   0,
		width:    width,
		showDesc: false,
	}
}

func (m SelectModel) Update(msg tea.Msg) (SelectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.expanded {
			if key.Matches(msg, DefaultKeyMap.Enter) || msg.String() == " " {
				m.expanded = true
				m.focused = true
			}
			if key.Matches(msg, DefaultKeyMap.Down) {
				if m.selected < len(m.filtered)-1 {
					m.selected++
					if m.selected-m.scroll > 8 {
						m.scroll = m.selected - 8
					}
				}
			}
			if key.Matches(msg, DefaultKeyMap.Up) {
				if m.selected > 0 {
					m.selected--
					if m.selected < m.scroll {
						m.scroll = m.selected
					}
				}
			}
		} else {
			if key.Matches(msg, DefaultKeyMap.Escape) {
				m.expanded = false
				m.focused = false
			}
			if key.Matches(msg, DefaultKeyMap.Enter) {
				m.expanded = false
				m.focused = false
			}
			if key.Matches(msg, DefaultKeyMap.Down) {
				if m.selected < len(m.filtered)-1 {
					m.selected++
					if m.selected-m.scroll > 8 {
						m.scroll = m.selected - 8
					}
				}
			}
			if key.Matches(msg, DefaultKeyMap.Up) {
				if m.selected > 0 {
					m.selected--
					if m.selected < m.scroll {
						m.scroll = m.selected
					}
				}
			}
		}
	}
	return m, nil
}

func (m SelectModel) View() string {
	if !m.expanded {
		selectedLabel := ""
		if len(m.filtered) > 0 && m.selected >= 0 && m.selected < len(m.filtered) {
			selectedLabel = m.filtered[m.selected].Label
		}
		trigger := SelectTriggerStyle.Render(selectedLabel)
		if selectedLabel == "" {
			trigger = SelectTriggerStyle.Render("Select...")
		}
		return lipgloss.JoinVertical(
			lipgloss.Left,
			LabelStyle.Render("Select:"),
			lipgloss.JoinHorizontal(lipgloss.Left, trigger, " ▼"),
		)
	}

	var options []string
	maxVisible := 10
	if len(m.filtered) < maxVisible {
		maxVisible = len(m.filtered)
	}

	for i := m.scroll; i < m.scroll+maxVisible && i < len(m.filtered); i++ {
		opt := m.filtered[i]
		prefix := "  "
		if i == m.selected {
			prefix = SelectCurrentStyle.Render("▶")
		}

		optLine := prefix + " " + opt.Label
		if m.showDesc && opt.Description != "" {
			optLine += " - " + opt.Description
		}

		if i == m.selected {
			options = append(options, SelectOptionStyleSelected.Render(optLine))
		} else {
			options = append(options, SelectOptionStyle.Render(optLine))
		}
	}

	boxHeight := maxVisible
	if boxHeight == 0 {
		boxHeight = 1
	}

	optionsView := lipgloss.JoinVertical(lipgloss.Left, options...)
	optionsView = SelectBoxStyle.Height(boxHeight).Render(optionsView)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		LabelStyle.Render("Select:"),
		optionsView,
	)
}

func (m *SelectModel) SetOptions(options []SelectOption) {
	m.options = options
	m.filtered = options
	if m.selected >= len(m.filtered) {
		m.selected = len(m.filtered) - 1
	}
}

func (m SelectModel) SelectedOption() SelectOption {
	if len(m.filtered) > 0 && m.selected >= 0 && m.selected < len(m.filtered) {
		return m.filtered[m.selected]
	}
	return SelectOption{}
}

func (m SelectModel) Value() string {
	opt := m.SelectedOption()
	return opt.Value
}

func (m *SelectModel) SetValue(value string) {
	for i, opt := range m.filtered {
		if opt.Value == value {
			m.selected = i
			return
		}
	}
}

func (m *SelectModel) IsExpanded() bool {
	return m.expanded
}

type StatusBadgeModel struct {
	status  string
	message string
}

func NewStatusBadge(status, message string) StatusBadgeModel {
	return StatusBadgeModel{
		status:  status,
		message: message,
	}
}

func (m StatusBadgeModel) View() string {
	var badge string
	var msgStyle lipgloss.Style

	switch m.status {
	case "success":
		badge = SuccessBadgeStyle.Render("✓ Success")
		msgStyle = SuccessStyle
	case "error":
		badge = ErrorBadgeStyle.Render("✗ Error")
		msgStyle = ErrorStyle
	case "warning":
		badge = WarningBadgeStyle.Render("⚠ Warning")
		msgStyle = WarningStyle
	case "info":
		badge = InfoBadgeStyle.Render("ℹ Info")
		msgStyle = InfoStyle
	case "pending":
		badge = InfoBadgeStyle.Render("◌ Pending")
		msgStyle = InfoStyle
	default:
		badge = InfoBadgeStyle.Render("○ Unknown")
		msgStyle = InfoStyle
	}

	if m.message != "" {
		return lipgloss.JoinHorizontal(
			lipgloss.Left,
			badge,
			" ",
			msgStyle.Render(m.message),
		)
	}

	return badge
}

func (m *StatusBadgeModel) SetStatus(status, message string) {
	m.status = status
	m.message = message
}

func (m *StatusBadgeModel) SetMessage(message string) {
	m.message = message
}

type ConfirmDialogModel struct {
	title          string
	message        string
	confirmed      bool
	focused        bool
	selectedButton int
}

func NewConfirmDialog(title, message string) ConfirmDialogModel {
	return ConfirmDialogModel{
		title:          title,
		message:        message,
		confirmed:      false,
		focused:        false,
		selectedButton: 0,
	}
}

func (m ConfirmDialogModel) Update(msg tea.Msg) (ConfirmDialogModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.focused {
			if key.Matches(msg, DefaultKeyMap.Left) {
				if m.selectedButton > 0 {
					m.selectedButton--
				}
			}
			if key.Matches(msg, DefaultKeyMap.Right) {
				if m.selectedButton < 1 {
					m.selectedButton++
				}
			}
			if key.Matches(msg, DefaultKeyMap.Enter) {
				m.confirmed = m.selectedButton == 0
			}
			if key.Matches(msg, DefaultKeyMap.Escape) {
				m.confirmed = false
			}
			if msg.String() == "y" {
				m.confirmed = true
			}
			if msg.String() == "n" {
				m.confirmed = false
			}
		}
	}
	return m, nil
}

func (m ConfirmDialogModel) View() string {
	dialogWidth := 50
	if len(m.message) > dialogWidth-10 {
		dialogWidth = len(m.message) + 10
	}

	titleStyle := lipgloss.Style{}.
		Foreground(lipgloss.Color("99")).
		Bold(true).
		Width(dialogWidth).
		Align(lipgloss.Center)

	messageStyle := lipgloss.Style{}.
		Foreground(lipgloss.Color("250")).
		Width(dialogWidth).
		Align(lipgloss.Center)

	buttonsStyle := lipgloss.Style{}.
		Width(dialogWidth).
		Align(lipgloss.Center)

	yesBtn := ConfirmYesStyle
	noBtn := ConfirmNoStyle

	if m.focused && m.selectedButton == 0 {
		yesBtn = ConfirmYesStyleSelected
	} else if m.focused && m.selectedButton == 1 {
		noBtn = ConfirmNoStyleSelected
	}

	buttons := lipgloss.JoinHorizontal(
		lipgloss.Center,
		yesBtn.Render("[Y] Yes"),
		noBtn.Render("[N] No"),
	)

	box := lipgloss.JoinVertical(
		lipgloss.Center,
		titleStyle.Render(m.title),
		"",
		messageStyle.Render(m.message),
		"",
		buttonsStyle.Render(buttons),
	)

	return ConfirmDialogStyle.Width(dialogWidth + 4).Render(box)
}

func (m *ConfirmDialogModel) Focus() {
	m.focused = true
}

func (m *ConfirmDialogModel) Blur() {
	m.focused = false
}

func (m ConfirmDialogModel) IsConfirmed() bool {
	return m.confirmed
}

func (m *ConfirmDialogModel) SetTitle(title string) {
	m.title = title
}

func (m *ConfirmDialogModel) SetMessage(message string) {
	m.message = message
}
