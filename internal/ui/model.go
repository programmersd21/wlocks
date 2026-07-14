package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"
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
	modeHelp
	modeStats
)

type sortMode int

const (
	sortByName sortMode = iota
	sortByDuration
	sortByPID
	sortByMode
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

	sortBy          sortMode
	sortReverse     bool
	scrollAnim      *ScrollAnimation
	fadeAnim        *FadeAnimation
	themeTransition *Animation
	statusMessage   string
	statusTimeout   time.Time
	firstRun        bool
	searchHistory   []string
}

func NewModel(targetPath string, themeName string, debug bool) *Model {
	theme := GetTheme(themeName)
	return &Model{
		mode:          modeStatic,
		theme:         theme,
		styles:        NewStyles(theme),
		keys:          DefaultKeyMap(),
		targetPath:    targetPath,
		debug:         debug,
		sortBy:        sortByDuration,
		scrollAnim:    NewScrollAnimation(0),
		fadeAnim:      NewFadeAnimation(),
		firstRun:      true,
		searchHistory: make([]string, 0),
	}
}

func (m *Model) Init() tea.Cmd {
	m.fadeAnim.FadeIn(300 * time.Millisecond)
	return tea.Batch(
		m.scanCmd(),
		animTickCmd(),
		refreshTickCmd(),
	)
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

type animTickMsg time.Time

type refreshTickMsg time.Time

type statusClearMsg struct{}

func animTickCmd() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

func refreshTickCmd() tea.Cmd {
	return tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

func statusClearCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case animTickMsg:
		m.fadeAnim.Update()
		m.scrollAnim.Update()
		if m.themeTransition != nil {
			m.themeTransition.Update()
		}
		// Keep the animation ticker running while any animation is active
		if m.fadeAnim.IsVisible() || m.scrollAnim.IsAnimating() ||
			(m.themeTransition != nil && m.themeTransition.IsRunning()) {
			return m, animTickCmd()
		}
		return m, nil

	case refreshTickMsg:
		// Always rescan and always reschedule the next tick
		return m, tea.Batch(m.scanCmd(), refreshTickCmd())

	case scanCompleteMsg:
		oldLocks := m.locks
		m.locks = msg.result.Locks
		m.permDenied = msg.result.PermissionDenied
		m.sortLocks()

		if !m.firstRun {
			added := len(m.locks) - len(oldLocks)
			if added != 0 {
				if added > 0 {
					m.setStatus(fmt.Sprintf("+%d new", added))
				} else {
					m.setStatus(fmt.Sprintf("%d closed", -added))
				}
			}
		}
		m.firstRun = false

		return m, nil

	case statusClearMsg:
		if !time.Now().Before(m.statusTimeout) {
			m.statusMessage = ""
		}
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
	case modeHelp:
		return m.handleHelpKey(key)
	case modeStats:
		return m.handleStatsKey(key)
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
	case modeHelp:
		content = m.viewHelp()
	case modeStats:
		content = m.viewStats()
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
	if m.statusMessage != "" {
		left += m.styles.Ghost.Render(" • ") + m.styles.Accent.Render(m.statusMessage)
	}

	var right string
	switch m.mode {
	case modeStatic:
		right = m.styles.Accent.Render("j/k") + m.styles.Tertiary.Render(" navigate  ") +
			m.styles.Accent.Render("enter") + m.styles.Tertiary.Render(" details  ") +
			m.styles.Accent.Render("/") + m.styles.Tertiary.Render(" search  ") +
			m.styles.Accent.Render("?") + m.styles.Tertiary.Render(" help  ") +
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
	case modeHelp, modeStats:
		right = m.styles.Accent.Render("esc") + m.styles.Tertiary.Render(" back  ") +
			m.styles.Accent.Render("j/k") + m.styles.Tertiary.Render(" scroll  ") +
			m.styles.Accent.Render("q") + m.styles.Tertiary.Render(" quit")
	}

	return m.joinLeftRight(left, right)
}

func (m *Model) handleStatsKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case Matches(key, m.keys.Esc), Matches(key, m.keys.Stats):
		m.mode = modeStatic
		return m, nil

	case Matches(key, m.keys.Down):
		m.detailScroll++
		return m, nil

	case Matches(key, m.keys.Up):
		m.detailScroll = max(0, m.detailScroll-1)
		return m, nil
	}

	return m, nil
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

func (m *Model) sortLocks() {
	switch m.sortBy {
	case sortByName:
		sort.Slice(m.locks, func(i, j int) bool {
			nameI := m.locks[i].Process.Name
			nameJ := m.locks[j].Process.Name
			if m.sortReverse {
				return nameI > nameJ
			}
			return nameI < nameJ
		})
	case sortByDuration:
		sort.Slice(m.locks, func(i, j int) bool {
			if m.sortReverse {
				return m.locks[i].Duration < m.locks[j].Duration
			}
			return m.locks[i].Duration > m.locks[j].Duration
		})
	case sortByPID:
		sort.Slice(m.locks, func(i, j int) bool {
			if m.sortReverse {
				return m.locks[i].Process.PID > m.locks[j].Process.PID
			}
			return m.locks[i].Process.PID < m.locks[j].Process.PID
		})
	case sortByMode:
		sort.Slice(m.locks, func(i, j int) bool {
			if m.sortReverse {
				return m.locks[i].FD.Mode < m.locks[j].FD.Mode
			}
			return m.locks[i].FD.Mode > m.locks[j].FD.Mode
		})
	}
}

func (m *Model) setStatus(msg string) {
	m.statusMessage = msg
	m.statusTimeout = time.Now().Add(3 * time.Second)
}

func (m *Model) cycleSortMode() {
	m.sortBy = (m.sortBy + 1) % 4
	m.sortLocks()
	sortNames := []string{"name", "duration", "pid", "mode"}
	m.setStatus("sort: " + sortNames[m.sortBy])
}
