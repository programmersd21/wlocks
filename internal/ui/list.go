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
		return m, nil

	case Matches(key, m.keys.Up):
		newIndex := max(0, m.selectedIndex-1)
		if newIndex != m.selectedIndex {
			m.selectedIndex = newIndex
			m.ensureVisible()
		}
		return m, nil

	case Matches(key, m.keys.Enter):
		if m.selectedIndex < len(m.locks) {
			m.detailLock = m.locks[m.selectedIndex]
			m.mode = modeDetail
			m.detailScroll = 0
			m.killConfirm = false
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
		return m, statusClearCmd()

	case Matches(key, m.keys.SortReverse):
		m.sortReverse = !m.sortReverse
		m.sortLocks()
		if m.sortReverse {
			m.setStatus("sort reversed")
		} else {
			m.setStatus("sort normal")
		}
		return m, statusClearCmd()
	}

	return m, nil
}

func (m *Model) viewStatic() string {
	var b strings.Builder

	headerText := m.targetPath
	if len(headerText) > m.width-10 {
		headerText = "..." + headerText[len(headerText)-(m.width-13):]
	}
	header := m.styles.Primary.Render(headerText)
	b.WriteString(header)
	b.WriteString("\n")

	underline := m.styles.Ghost.Render(strings.Repeat("─", min(len(headerText), m.width-4)))
	b.WriteString(underline)
	b.WriteString("\n\n")

	if len(m.locks) == 0 {
		empty := m.styles.Tertiary.Render("nothing is holding this file.")
		b.WriteString(empty)
		b.WriteString("\n")
		return b.String()
	}

	sortModeStr := []string{"name", "duration", "pid", "mode"}[m.sortBy]
	arrow := "↓"
	if m.sortReverse {
		arrow = "↑"
	}
	sortIndicator := m.styles.Tertiary.Render(fmt.Sprintf("sorted by %s %s", sortModeStr, arrow))
	b.WriteString(sortIndicator)
	b.WriteString("\n\n")

	overhead := 8
	if m.debug && m.permDenied > 0 {
		overhead += 2
	}
	maxVisible := m.height - overhead
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := m.scrollOffset
	end := min(start+maxVisible, len(m.locks))

	for i := start; i < end; i++ {
		lock := m.locks[i]
		row := m.renderLockRow(lock, i == m.selectedIndex)
		b.WriteString(row)
		b.WriteString("\n")
	}

	if len(m.locks) > maxVisible {
		scrollIndicator := fmt.Sprintf("  %d/%d", m.selectedIndex+1, len(m.locks))
		b.WriteString("\n")
		b.WriteString(m.styles.Ghost.Render(scrollIndicator))
	}

	if m.debug && m.permDenied > 0 {
		b.WriteString("\n")
		debugMsg := m.styles.Tertiary.Render(fmt.Sprintf("%d processes hidden - insufficient permissions", m.permDenied))
		b.WriteString(debugMsg)
	}

	return b.String()
}

func (m *Model) renderLockRow(lock *proc.LockInfo, selected bool) string {
	var bar string
	if selected {
		bar = m.styles.Accent.Render("▌ ")
	} else {
		bar = m.styles.Ghost.Render("  ")
	}

	processStyle := m.styles.Secondary
	if selected {
		processStyle = m.styles.Primary
	}
	processName := lock.Process.Name
	if processName == "" && len(lock.Process.Cmdline) > 0 {
		processName = lock.Process.Cmdline[0]
	}
	processCol := processStyle.Render(truncate(processName, 24))
	processCol = lipgloss.NewStyle().Width(24).Render(processCol)

	pidText := fmt.Sprintf("%d", lock.Process.PID)
	pidStyle := m.styles.Tertiary
	if selected {
		pidStyle = m.styles.Secondary
	}
	pidCol := pidStyle.Render(pidText)
	pidCol = lipgloss.NewStyle().Width(8).Align(lipgloss.Right).Render(pidCol)

	modeText := lock.FD.Mode.String()
	var modeStyle lipgloss.Style
	if lock.FD.Mode == proc.FDModeWrite || lock.FD.Mode == proc.FDModeReadWrite {
		modeStyle = m.styles.Danger
	} else {
		modeStyle = m.styles.Positive
	}
	modeCol := modeStyle.Render(modeText)
	modeCol = lipgloss.NewStyle().Width(10).Render(modeCol)

	durationText := formatDuration(lock.Duration)
	durationStyle := m.styles.Tertiary
	if selected {
		durationStyle = m.styles.Secondary
	}
	durationCol := durationStyle.Render(durationText)
	durationCol = lipgloss.NewStyle().Width(10).Align(lipgloss.Right).Render(durationCol)

	return lipgloss.JoinHorizontal(lipgloss.Left, bar, processCol, pidCol, modeCol, durationCol)
}

func (m *Model) ensureVisible() {
	overhead := 6
	if m.debug && m.permDenied > 0 {
		overhead += 2
	}
	maxVisible := m.height - overhead
	if maxVisible < 1 {
		maxVisible = 1
	}

	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	}
	if m.selectedIndex >= m.scrollOffset+maxVisible {
		m.scrollOffset = m.selectedIndex - maxVisible + 1
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
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
