package ui

import (
	"fmt"
	"strings"
	"wlocks/internal/proc"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type mode int

const (
	modeStatic mode = iota
	modeDetail
	modeSearch
	modePalette
)

type Model struct {
	mode          mode
	theme         *Theme
	styles        *Styles
	keys          KeyMap
	width         int
	height        int
	targetPath    string
	locks         []*proc.LockInfo
	selectedIndex int
	scrollOffset  int
	detailLock    *proc.LockInfo
	detailScroll  int
	killConfirm   bool
	searchQuery   string
	searchResults []*proc.LockInfo
	paletteIndex  int
	debug         bool
	permDenied    int
}

func NewModel(targetPath string, themeName string, debug bool) *Model {
	theme := GetTheme(themeName)
	return &Model{
		mode:       modeStatic,
		theme:      theme,
		styles:     NewStyles(theme),
		keys:       DefaultKeyMap(),
		targetPath: targetPath,
		debug:      debug,
	}
}

func (m *Model) Init() tea.Cmd {
	return m.scanCmd()
}

func (m *Model) scanCmd() tea.Cmd {
	path := m.targetPath
	debug := m.debug
	return func() tea.Msg {
		result := proc.ScanForPath(path, debug)
		return scanCompleteMsg{result}
	}
}

type scanCompleteMsg struct {
	result *proc.ScanResult
}

type killTimeoutMsg struct{}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case scanCompleteMsg:
		m.locks = msg.result.Locks
		m.permDenied = msg.result.PermissionDenied
		return m, nil

	case killTimeoutMsg:
		m.killConfirm = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()

	if Matches(key, m.keys.Quit) {
		return m, tea.Quit
	}

	switch m.mode {
	case modeStatic:
		return m.handleStaticKey(key)
	case modeDetail:
		return m.handleDetailKey(key)
	case modeSearch:
		return m.handleSearchKey(key)
	case modePalette:
		return m.handlePaletteKey(key)
	}

	return m, nil
}

func (m *Model) View() tea.View {
	if m.width == 0 || m.height == 0 {
		v := tea.NewView("")
		v.AltScreen = true
		return v
	}

	var content string
	switch m.mode {
	case modeStatic:
		content = m.viewStatic()
	case modeDetail:
		content = m.viewDetail()
	case modeSearch:
		content = m.viewSearch()
	case modePalette:
		content = m.viewPalette()
	}

	footer := m.viewFooter()
	paddedContent := m.padAndAppendFooter(content, footer)

	styledContent := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Render(paddedContent)

	v := tea.NewView(styledContent)
	v.AltScreen = true
	return v
}

func (m *Model) viewFooter() string {
	left := m.styles.Primary.Render("w l o c k s") + m.styles.Ghost.Render(" • ") + m.styles.Secondary.Render(m.theme.Name)
	if len(m.locks) > 0 {
		left += m.styles.Ghost.Render(" • ") + m.styles.Secondary.Render(fmt.Sprintf("%d locks", len(m.locks)))
	}

	var right string
	switch m.mode {
	case modeStatic:
		right = m.styles.Accent.Render("j/k") + m.styles.Tertiary.Render(" navigate  ") +
			m.styles.Accent.Render("enter") + m.styles.Tertiary.Render(" details  ") +
			m.styles.Accent.Render("/") + m.styles.Tertiary.Render(" search  ") +
			m.styles.Accent.Render("?") + m.styles.Tertiary.Render(" menu  ") +
			m.styles.Accent.Render("T") + m.styles.Tertiary.Render(" theme  ") +
			m.styles.Accent.Render("q") + m.styles.Tertiary.Render(" quit")
	case modeDetail:
		right = m.styles.Accent.Render("esc") + m.styles.Tertiary.Render(" back  ") +
			m.styles.Accent.Render("K") + m.styles.Tertiary.Render(" kill  ") +
			m.styles.Accent.Render("q") + m.styles.Tertiary.Render(" quit")
	case modeSearch:
		right = m.styles.Accent.Render("esc") + m.styles.Tertiary.Render(" back  ") +
			m.rightNav() +
			m.styles.Accent.Render("enter") + m.styles.Tertiary.Render(" details  ") +
			m.styles.Accent.Render("q") + m.styles.Tertiary.Render(" quit")
	case modePalette:
		right = m.styles.Accent.Render("j/k") + m.styles.Tertiary.Render(" navigate  ") +
			m.styles.Accent.Render("enter") + m.styles.Tertiary.Render(" select  ") +
			m.styles.Accent.Render("esc") + m.styles.Tertiary.Render(" back  ") +
			m.styles.Accent.Render("q") + m.styles.Tertiary.Render(" quit")
	}

	return m.joinLeftRight(left, right)
}

func (m *Model) rightNav() string {
	if len(m.searchResults) > 0 {
		return m.styles.Accent.Render("j/k") + m.styles.Tertiary.Render(" navigate  ")
	}
	return ""
}

func (m *Model) padAndAppendFooter(content string, footer string) string {
	lines := strings.Split(content, "\n")
	contentHeight := len(lines)

	if contentHeight >= m.height-1 {
		if m.height > 1 {
			return strings.Join(lines[:m.height-2], "\n") + "\n\n" + footer
		}
		return footer
	}

	padding := strings.Repeat("\n", m.height-contentHeight-1)
	return content + padding + footer
}

func (m *Model) joinLeftRight(left string, right string) string {
	leftLen := lipgloss.Width(left)
	rightLen := lipgloss.Width(right)

	if leftLen+rightLen >= m.width {
		return left + " " + right
	}

	spaces := strings.Repeat(" ", m.width-leftLen-rightLen)
	return left + spaces + right
}

func (m *Model) SetTheme(themeName string) {
	m.theme = GetTheme(themeName)
	m.styles = NewStyles(m.theme)
}
