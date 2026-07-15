package ui

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"wlocks/internal/config"
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
	detailTab     int
	detailFiles   []string
	detailEnv     []string
	killConfirm   bool
	pauseConfirm  bool
	forceConfirm  bool
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
	cfg             *config.Config
}

type ModelConfig struct {
	TargetPath string
	ThemeName  string
	Debug      bool
	Config     *config.Config
}

func NewModel(mc ModelConfig) *Model {
	theme := GetTheme(mc.ThemeName)

	sortBy := sortByDuration
	switch mc.Config.DefaultSort {
	case "name":
		sortBy = sortByName
	case "pid":
		sortBy = sortByPID
	case "mode":
		sortBy = sortByMode
	}

	return &Model{
		mode:          modeStatic,
		theme:         theme,
		styles:        NewStyles(theme),
		keys:          DefaultKeyMap(),
		targetPath:    mc.TargetPath,
		debug:         mc.Debug,
		sortBy:        sortBy,
		scrollAnim:    NewScrollAnimation(0),
		fadeAnim:      NewFadeAnimation(),
		firstRun:      true,
		searchHistory: make([]string, 0),
		detailTab:     0,
		cfg:           mc.Config,
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

type killProcessMsg struct {
	pid int
	err error
}

type pauseProcessMsg struct {
	pid    int
	paused bool
	err    error
}

type animTickMsg time.Time

type refreshTickMsg time.Time

type statusClearMsg struct{}

func animTickCmd() tea.Cmd {
	return tea.Tick(16*time.Millisecond, func(t time.Time) tea.Msg {
		return animTickMsg(t)
	})
}

func refreshTickCmd() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return refreshTickMsg(t)
	})
}

func statusClearCmd() tea.Cmd {
	return tea.Tick(3*time.Second, func(t time.Time) tea.Msg {
		return statusClearMsg{}
	})
}

func (m *Model) persistConfig() {
	cfg := &config.Config{
		Theme:           m.theme.Name,
		DefaultSort:     []string{"name", "duration", "pid", "mode"}[m.sortBy],
		LiveRefreshRate: 1,
		AnimationSpeed:  "normal",
	}
	if err := config.Save(cfg); err != nil && m.debug {
		m.setStatus(fmt.Sprintf("config save failed: %v", err))
	}
}

func (m *Model) killProcessCmd(pid int) tea.Cmd {
	return func() tea.Msg {
		// Try SIGTERM first (graceful)
		err := killProcess(pid, false)
		return killProcessMsg{pid: pid, err: err}
	}
}

func (m *Model) killForceProcessCmd(pid int) tea.Cmd {
	return func() tea.Msg {
		// Try SIGKILL (force)
		err := killProcess(pid, true)
		return killProcessMsg{pid: pid, err: err}
	}
}

func (m *Model) pauseProcessCmd(pid int, stop bool) tea.Cmd {
	return func() tea.Msg {
		err := pauseProcess(pid, stop)
		return pauseProcessMsg{pid: pid, paused: stop, err: err}
	}
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
		m.pauseConfirm = false
		m.forceConfirm = false
		return m, nil

	case killProcessMsg:
		m.killConfirm = false
		m.forceConfirm = false
		m.mode = modeStatic
		if msg.err != nil {
			m.setStatus(fmt.Sprintf("kill failed: %v", msg.err))
		} else {
			m.setStatus(fmt.Sprintf("killed process %d", msg.pid))
			// Trigger immediate refresh to update the list
			return m, tea.Batch(m.scanCmd(), statusClearCmd())
		}
		return m, statusClearCmd()

	case pauseProcessMsg:
		m.pauseConfirm = false
		if msg.err != nil {
			m.setStatus(fmt.Sprintf("pause failed: %v", msg.err))
		} else {
			if msg.paused {
				m.setStatus(fmt.Sprintf("paused process %d", msg.pid))
			} else {
				m.setStatus(fmt.Sprintf("resumed process %d", msg.pid))
			}
			return m, tea.Batch(m.scanCmd(), statusClearCmd())
		}
		return m, statusClearCmd()

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
	styles := m.currentStyles()
	left := styles.Ghost.Render("wlocks") + styles.Ghost.Render(" • ") + styles.Secondary.Render(m.theme.Name)
	if len(m.locks) > 0 {
		left += styles.Ghost.Render(" • ") + styles.Secondary.Render(fmt.Sprintf("%d locks", len(m.locks)))
	}
	if m.statusMessage != "" {
		left += styles.Ghost.Render(" • ") + styles.Accent.Render(m.statusMessage)
	}

	var right string
	switch m.mode {
	case modeStatic:
		right = styles.Accent.Render("j/k") + styles.Tertiary.Render(" navigate  ") +
			styles.Accent.Render("enter") + styles.Tertiary.Render(" details  ") +
			styles.Accent.Render("/") + styles.Tertiary.Render(" search  ") +
			styles.Accent.Render("?") + styles.Tertiary.Render(" help  ") +
			styles.Accent.Render("q") + styles.Tertiary.Render(" quit")
	case modeDetail:
		right = styles.Accent.Render("esc") + styles.Tertiary.Render(" back  ") +
			styles.Accent.Render("K") + styles.Tertiary.Render(" kill  ") +
			styles.Accent.Render("q") + styles.Tertiary.Render(" quit")
	case modeSearch:
		right = styles.Accent.Render("esc") + styles.Tertiary.Render(" back  ") +
			m.rightNav() +
			styles.Accent.Render("enter") + styles.Tertiary.Render(" details  ") +
			styles.Accent.Render("q") + styles.Tertiary.Render(" quit")
	case modePalette:
		right = styles.Accent.Render("j/k") + styles.Tertiary.Render(" navigate  ") +
			styles.Accent.Render("enter") + styles.Tertiary.Render(" select  ") +
			styles.Accent.Render("esc") + styles.Tertiary.Render(" back  ") +
			styles.Accent.Render("q") + styles.Tertiary.Render(" quit")
	case modeHelp, modeStats:
		right = styles.Accent.Render("esc") + styles.Tertiary.Render(" back  ") +
			styles.Accent.Render("j/k") + styles.Tertiary.Render(" scroll  ") +
			styles.Accent.Render("q") + styles.Tertiary.Render(" quit")
	}

	return m.joinLeftRight(left, right)
}

func (m *Model) currentStyles() *Styles {
	if !m.fadeAnim.IsVisible() {
		return m.styles
	}
	opacity := m.fadeAnim.Opacity()
	fadedTheme := &Theme{
		Name:          m.theme.Name,
		TextPrimary:   interpolateColor(m.theme.TextGhost, m.theme.TextPrimary, opacity),
		TextSecondary: interpolateColor(m.theme.TextGhost, m.theme.TextSecondary, opacity),
		TextTertiary:  interpolateColor(m.theme.TextGhost, m.theme.TextTertiary, opacity),
		TextGhost:     m.theme.TextGhost,
		Accent:        interpolateColor(m.theme.TextGhost, m.theme.Accent, opacity),
		AccentDim:     interpolateColor(m.theme.TextGhost, m.theme.AccentDim, opacity),
		Positive:      interpolateColor(m.theme.TextGhost, m.theme.Positive, opacity),
		Warning:       interpolateColor(m.theme.TextGhost, m.theme.Warning, opacity),
		Danger:        interpolateColor(m.theme.TextGhost, m.theme.Danger, opacity),
	}
	return NewStyles(fadedTheme)
}

func (m *Model) SetTheme(themeName string) {
	m.theme = GetTheme(themeName)
	m.styles = NewStyles(m.theme)
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
