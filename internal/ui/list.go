package ui

import (
	"fmt"
	"strings"
	"time"
	"wlocks/internal/proc"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) handleStaticKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case Matches(key, m.keys.Down):
		if len(m.locks) > 0 {
			newIndex := min(m.selectedIndex+1, len(m.locks)-1)
			if newIndex != m.selectedIndex {
				m.selectedIndex = newIndex
				m.ensureVisible()
			}
		}
		return m, animTickCmd()

	case Matches(key, m.keys.Up):
		newIndex := max(0, m.selectedIndex-1)
		if newIndex != m.selectedIndex {
			m.selectedIndex = newIndex
			m.ensureVisible()
		}
		return m, animTickCmd()

	case Matches(key, m.keys.Enter):
		if m.selectedIndex < len(m.locks) {
			m.detailLock = m.locks[m.selectedIndex]
			m.mode = modeDetail
			m.detailScroll = 0
			m.detailTab = 0
			m.killConfirm = false
			m.pauseConfirm = false
			m.forceConfirm = false
			m.detailFiles = proc.GetProcessOpenFiles(m.detailLock.Process.PID)
			m.detailEnv = proc.GetProcessEnv(m.detailLock.Process.PID)
			m.fadeAnim.FadeIn(200 * time.Millisecond)
		}
		return m, animTickCmd()

	case Matches(key, m.keys.Search):
		m.mode = modeSearch
		m.searchQuery = ""
		m.updateSearchResults()
		m.fadeAnim.FadeIn(200 * time.Millisecond)
		return m, animTickCmd()

	case Matches(key, m.keys.Refresh):
		m.setStatus("refreshing...")
		return m, tea.Batch(m.scanCmd(), statusClearCmd())

	case Matches(key, m.keys.ThemeCycle):
		nextTheme := NextTheme(m.theme.Name)
		m.SetTheme(nextTheme)
		m.setStatus("theme: " + nextTheme)
		m.persistConfig()
		return m, statusClearCmd()

	case Matches(key, m.keys.CommandPalette):
		m.mode = modePalette
		m.paletteIndex = 0
		m.fadeAnim.FadeIn(200 * time.Millisecond)
		return m, animTickCmd()

	case Matches(key, m.keys.Help):
		m.mode = modeHelp
		m.detailScroll = 0
		m.fadeAnim.FadeIn(200 * time.Millisecond)
		return m, animTickCmd()

	case Matches(key, m.keys.Stats):
		m.mode = modeStats
		m.detailScroll = 0
		m.fadeAnim.FadeIn(200 * time.Millisecond)
		return m, animTickCmd()

	case Matches(key, m.keys.Sort):
		m.cycleSortMode()
		m.persistConfig()
		return m, statusClearCmd()

	case Matches(key, m.keys.SortReverse):
		m.sortReverse = !m.sortReverse
		m.sortLocks()
		if m.sortReverse {
			m.setStatus("sort reversed")
		} else {
			m.setStatus("sort normal")
		}
		m.persistConfig()
		return m, statusClearCmd()
	}

	return m, nil
}

func (m *Model) viewStatic() string {
	var b strings.Builder
	styles := m.currentStyles()

	b.WriteString("\n")

	headerText := m.targetPath
	maxPathLen := m.width - 25
	if maxPathLen < 10 {
		maxPathLen = 10
	}
	if len(headerText) > maxPathLen {
		headerText = "..." + headerText[len(headerText)-maxPathLen:]
	}
	pathDisplay := styles.Primary.Bold(true).Render(headerText)

	var countDisplay string
	if len(m.locks) > 0 {
		countDisplay = styles.Secondary.Render(fmt.Sprintf("%d active locks", len(m.locks)))
	} else {
		countDisplay = styles.Tertiary.Render("0 active locks")
	}

	headerLine := pathDisplay
	spacesNeeded := m.width - lipgloss.Width(headerLine) - lipgloss.Width(countDisplay) - 4
	if spacesNeeded > 0 {
		headerLine += strings.Repeat(" ", spacesNeeded) + countDisplay
	}

	b.WriteString("  " + headerLine + "\n\n")

	if len(m.locks) == 0 {
		b.WriteString("\n")
		emptyTitle := styles.Tertiary.Render("no active locks")
		b.WriteString("  " + emptyTitle + "\n")
		return b.String()
	}

	pHeader := styles.Ghost.Render("process")
	pHeader = lipgloss.NewStyle().Width(20).Render(pHeader)

	pidHeader := styles.Ghost.Render("pid")
	pidHeader = lipgloss.NewStyle().Width(7).Align(lipgloss.Right).Render(pidHeader)

	modeHeader := styles.Ghost.Render("mode")
	modeHeader = lipgloss.NewStyle().Width(9).Render(modeHeader)

	durHeader := styles.Ghost.Render("duration")
	durHeader = lipgloss.NewStyle().Width(9).Align(lipgloss.Right).Render(durHeader)

	colHeader := lipgloss.JoinHorizontal(lipgloss.Left, "  ", pHeader, "  ", pidHeader, "  ", modeHeader, "  ", durHeader)
	b.WriteString("  " + colHeader + "\n")
	b.WriteString("\n")

	overhead := 8
	if m.debug && m.permDenied > 0 {
		overhead += 2
	}
	maxVisible := m.height - overhead
	if maxVisible < 1 {
		maxVisible = 1
	}

	animatedScroll := m.scrollAnim.Update()
	if animatedScroll < 0 {
		animatedScroll = 0
	}
	if animatedScroll >= len(m.locks) {
		animatedScroll = max(0, len(m.locks)-1)
	}
	start := animatedScroll
	end := min(start+maxVisible, len(m.locks))

	for i := start; i < end; i++ {
		lock := m.locks[i]
		row := m.renderLockRow(lock, i == m.selectedIndex)
		b.WriteString("  " + row + "\n")
	}

	if len(m.locks) > maxVisible {
		b.WriteString("\n")
		progressText := fmt.Sprintf("  %d - %d of %d locks (j/k to scroll)", start+1, end, len(m.locks))
		b.WriteString(styles.Tertiary.Render(progressText))
	}

	if m.debug && m.permDenied > 0 {
		b.WriteString("\n")
		debugMsg := styles.Tertiary.Render(fmt.Sprintf("  %d processes hidden - insufficient permissions", m.permDenied))
		b.WriteString(debugMsg)
	}

	return b.String()
}

func (m *Model) renderLockRow(lock *proc.LockInfo, selected bool) string {
	styles := m.currentStyles()
	var bar string
	if selected {
		bar = styles.Accent.Render("┃ ")
	} else {
		bar = styles.Ghost.Render("  ")
	}

	processStyle := styles.Secondary
	if selected {
		processStyle = styles.Primary.Bold(true)
	}
	processName := lock.Process.Name
	if processName == "" && len(lock.Process.Cmdline) > 0 {
		processName = lock.Process.Cmdline[0]
	}
	processCol := processStyle.Render(truncate(processName, 20))
	processCol = lipgloss.NewStyle().Width(20).Render(processCol)

	pidText := fmt.Sprintf("%d", lock.Process.PID)
	pidStyle := styles.Tertiary
	if selected {
		pidStyle = styles.Secondary
	}
	pidCol := pidStyle.Render(pidText)
	pidCol = lipgloss.NewStyle().Width(7).Align(lipgloss.Right).Render(pidCol)

	modeText := lock.FD.Mode.String()
	var modeStyle lipgloss.Style
	if lock.FD.Mode == proc.FDModeWrite || lock.FD.Mode == proc.FDModeReadWrite {
		modeStyle = styles.Danger
	} else {
		modeStyle = styles.Positive
	}
	if selected {
		modeStyle = modeStyle.Bold(true)
	}
	modeCol := modeStyle.Render(modeText)
	modeCol = lipgloss.NewStyle().Width(9).Render(modeCol)

	durationText := formatDuration(lock.Duration)
	durationStyle := styles.Tertiary
	if selected {
		durationStyle = styles.Secondary
	}
	durationCol := durationStyle.Render(durationText)
	durationCol = lipgloss.NewStyle().Width(9).Align(lipgloss.Right).Render(durationCol)

	return lipgloss.JoinHorizontal(lipgloss.Left, bar, processCol, "  ", pidCol, "  ", modeCol, "  ", durationCol)
}

func (m *Model) ensureVisible() {
	overhead := 8
	if m.debug && m.permDenied > 0 {
		overhead += 2
	}
	maxVisible := m.height - overhead
	if maxVisible < 1 {
		maxVisible = 1
	}

	targetOffset := m.scrollOffset
	if m.selectedIndex < targetOffset {
		targetOffset = m.selectedIndex
	} else if m.selectedIndex >= targetOffset+maxVisible {
		targetOffset = m.selectedIndex - maxVisible + 1
	}

	if targetOffset != m.scrollOffset {
		m.scrollOffset = targetOffset
		m.scrollAnim.SetTarget(targetOffset, 150*time.Millisecond)
	}
}

func formatDuration(d time.Duration) string {
	if d >= 24*time.Hour {
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	if h > 0 {
		return fmt.Sprintf("%dh%dm%ds", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
