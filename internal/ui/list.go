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
			m.selectedIndex = min(m.selectedIndex+1, len(m.locks)-1)
			m.ensureVisible()
		}
		return m, nil

	case Matches(key, m.keys.Up):
		m.selectedIndex = max(0, m.selectedIndex-1)
		m.ensureVisible()
		return m, nil

	case Matches(key, m.keys.Enter):
		if m.selectedIndex < len(m.locks) {
			m.detailLock = m.locks[m.selectedIndex]
			m.mode = modeDetail
			m.detailScroll = 0
			m.killConfirm = false
		}
		return m, nil

	case Matches(key, m.keys.Search):
		m.mode = modeSearch
		m.searchQuery = ""
		m.updateSearchResults()
		return m, nil

	case Matches(key, m.keys.Refresh):
		return m, m.scanCmd()

	case Matches(key, m.keys.ThemeCycle):
		nextTheme := NextTheme(m.theme.Name)
		m.SetTheme(nextTheme)
		return m, nil

	case Matches(key, m.keys.CommandPalette):
		m.mode = modePalette
		m.paletteIndex = 0
		return m, nil
	}

	return m, nil
}

func (m *Model) viewStatic() string {
	var b strings.Builder

	header := m.styles.Primary.Render(m.targetPath)
	b.WriteString(header)
	b.WriteString("\n\n")

	if len(m.locks) == 0 {
		empty := m.styles.Tertiary.Render("nothing is holding this file.")
		b.WriteString(empty)
		b.WriteString("\n")
		return b.String()
	}

	subheader := m.styles.Tertiary.Render("open by")
	b.WriteString(subheader)
	b.WriteString("\n\n")

	overhead := 6
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
		bar = "  "
	}

	processStyle := m.styles.Secondary
	if selected {
		processStyle = m.styles.Primary
	}
	processName := lock.Process.Name
	if processName == "" && len(lock.Process.Cmdline) > 0 {
		processName = lock.Process.Cmdline[0]
	}
	processCol := processStyle.Render(truncate(processName, 20))
	processCol = lipgloss.NewStyle().Width(20).Render(processCol)

	modeText := lock.FD.Mode.String()
	var modeStyle lipgloss.Style
	if lock.FD.Mode == proc.FDModeWrite || lock.FD.Mode == proc.FDModeReadWrite {
		modeStyle = m.styles.Danger
	} else {
		modeStyle = m.styles.Positive
	}
	modeCol := modeStyle.Render(modeText)
	modeCol = lipgloss.NewStyle().Width(12).Render(modeCol)

	durationText := formatDuration(lock.Duration)
	durationCol := m.styles.Tertiary.Render(durationText)
	durationCol = lipgloss.NewStyle().Width(12).Align(lipgloss.Right).Render(durationCol)

	row := lipgloss.JoinHorizontal(lipgloss.Left, bar, processCol, modeCol, durationCol)
	return row
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
