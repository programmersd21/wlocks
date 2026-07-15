package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) handleHelpKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case Matches(key, m.keys.Esc), Matches(key, m.keys.Help):
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

func (m *Model) viewHelp() string {
	styles := m.currentStyles()
	var b strings.Builder

	b.WriteString("\n")

	pathDisplay := styles.Primary.Bold(true).Render("keyboard shortcuts")

	headerLine := pathDisplay
	b.WriteString("  " + headerLine + "\n\n")

	type helpItem struct{ key, desc string }
	type helpSection struct {
		title string
		items []helpItem
	}

	sections := []helpSection{
		{
			title: "navigation",
			items: []helpItem{
				{"j/k or ↑↓", "navigate list"},
				{"enter", "show details"},
				{"esc", "go back"},
			},
		},
		{
			title: "actions",
			items: []helpItem{
				{"/", "fuzzy search"},
				{"r", "refresh snapshot"},
				{"K", "kill process"},
			},
		},
		{
			title: "views",
			items: []helpItem{
				{"?", "toggle help"},
				{"i", "show statistics"},
				{"ctrl+p", "command palette"},
			},
		},
		{
			title: "customization",
			items: []helpItem{
				{"T", "cycle theme"},
				{"s", "cycle sort mode"},
				{"S", "reverse sort"},
			},
		},
		{
			title: "other",
			items: []helpItem{
				{"q", "quit"},
				{"ctrl+c", "force quit"},
			},
		},
	}

	renderSection := func(sec helpSection) string {
		var sb strings.Builder
		sb.WriteString(styles.Secondary.Render(sec.title) + "\n")
		for _, item := range sec.items {
			keyStyle := styles.Accent.Render(item.key)
			keyCol := lipgloss.NewStyle().Width(12).Render(keyStyle)
			descStyle := styles.Tertiary.Render(item.desc)
			sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, keyCol, descStyle) + "\n")
		}
		return sb.String()
	}

	// 2-Column layout
	col1 := renderSection(sections[0]) + "\n" + renderSection(sections[1])
	col2 := renderSection(sections[2]) + "\n" + renderSection(sections[3]) + "\n" + renderSection(sections[4])

	cols := lipgloss.JoinHorizontal(lipgloss.Top, col1, "      ", col2)

	// Indent the columns
	lines := strings.Split(cols, "\n")
	for _, line := range lines {
		b.WriteString("  " + line + "\n")
	}

	return b.String()
}

func (m *Model) viewStats() string {
	var b strings.Builder
	styles := m.currentStyles()

	b.WriteString("\n")

	statsHeader := styles.Primary.Bold(true).Render("aggregate metrics")
	b.WriteString("  " + statsHeader + "\n\n")

	totalLocks := len(m.locks)
	readCount := 0
	writeCount := 0
	processMap := make(map[int]bool)

	for _, lock := range m.locks {
		processMap[lock.Process.PID] = true
		if lock.FD.Mode == 0 {
			readCount++
		} else {
			writeCount++
		}
	}

	uniqueProcesses := len(processMap)

	labelStyle := styles.Tertiary
	valStyle := styles.Accent

	col1 := labelStyle.Render("active locks") + "\n" + valStyle.Render(fmt.Sprintf("  %d", totalLocks))
	col2 := labelStyle.Render("unique processes") + "\n" + valStyle.Render(fmt.Sprintf("  %d", uniqueProcesses))
	col3 := labelStyle.Render("read / write") + "\n" + valStyle.Render(fmt.Sprintf("  %d / %d", readCount, writeCount))

	metrics := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(24).Render(col1),
		lipgloss.NewStyle().Width(24).Render(col2),
		lipgloss.NewStyle().Width(24).Render(col3),
	)

	b.WriteString("  " + metrics + "\n\n\n")

	b.WriteString("  " + styles.Secondary.Render("system details") + "\n")

	details := []struct {
		label string
		value string
	}{
		{"target resource", m.targetPath},
		{"permission denied", formatInt(m.permDenied)},
		{"active theme", m.theme.Name},
		{"current sort mode", []string{"name", "duration", "pid", "mode"}[m.sortBy]},
	}

	for _, detail := range details {
		lbl := styles.Tertiary.Render(detail.label)
		lbl = lipgloss.NewStyle().Width(20).Render(lbl)
		val := styles.Primary.Render(detail.value)
		fmt.Fprintf(&b, "  %s %s\n", lbl, val)
	}

	return b.String()
}

func formatInt(n int) string {
	if n == 0 {
		return "0"
	}

	s := ""
	for i := 0; n > 0; i++ {
		if i > 0 && i%3 == 0 {
			s = "," + s
		}
		s = string(rune('0'+(n%10))) + s
		n /= 10
	}
	return s
}
