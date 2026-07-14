package ui

import (
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
	var b strings.Builder

	header := m.styles.Primary.Render("wlocks help")
	b.WriteString(header)
	b.WriteString("\n\n")

	sections := []struct {
		title string
		items []struct{ key, desc string }
	}{
		{
			title: "navigation",
			items: []struct{ key, desc string }{
				{"j/k or ↑↓", "navigate list"},
				{"enter", "show process details"},
				{"esc", "go back / clear search"},
			},
		},
		{
			title: "actions",
			items: []struct{ key, desc string }{
				{"/", "search processes"},
				{"r", "refresh snapshot"},
				{"K", "kill process (with confirm)"},
			},
		},
		{
			title: "views",
			items: []struct{ key, desc string }{
				{"?", "show this help"},
				{"i", "show statistics"},
				{"ctrl+p", "command palette"},
			},
		},
		{
			title: "customization",
			items: []struct{ key, desc string }{
				{"T", "cycle theme"},
				{"s", "cycle sort mode"},
				{"S", "reverse sort"},
			},
		},
		{
			title: "other",
			items: []struct{ key, desc string }{
				{"q", "quit"},
				{"ctrl+c", "force quit"},
			},
		},
	}

	for _, section := range sections {
		sectionTitle := m.styles.Secondary.Render(section.title)
		b.WriteString(sectionTitle)
		b.WriteString("\n\n")

		for _, item := range section.items {
			keyStyle := m.styles.Accent.Render(item.key)
			keyCol := lipgloss.NewStyle().Width(20).Render(keyStyle)

			descStyle := m.styles.Tertiary.Render(item.desc)

			line := lipgloss.JoinHorizontal(lipgloss.Left, "  ", keyCol, descStyle)
			b.WriteString(line)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	tip := m.styles.Ghost.Render("tip: press ? anytime to see available actions")
	b.WriteString(tip)

	return b.String()
}

func (m *Model) viewStats() string {
	var b strings.Builder

	header := m.styles.Primary.Render("statistics")
	b.WriteString(header)
	b.WriteString("\n\n")

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

	stats := []struct {
		label string
		value string
	}{
		{"target", m.targetPath},
		{"total locks", formatInt(totalLocks)},
		{"unique processes", formatInt(uniqueProcesses)},
		{"read locks", formatInt(readCount)},
		{"write locks", formatInt(writeCount)},
		{"permission denied", formatInt(m.permDenied)},
	}

	labelWidth := 20
	for _, stat := range stats {
		label := m.styles.Tertiary.Render(stat.label)
		label = lipgloss.NewStyle().Width(labelWidth).Render(label)

		value := m.styles.Primary.Render(stat.value)

		line := lipgloss.JoinHorizontal(lipgloss.Left, "  ", label, value)
		b.WriteString(line)
		b.WriteString("\n")
	}

	b.WriteString("\n")

	sortModeStr := []string{"name", "duration", "pid", "mode"}[m.sortBy]
	if m.sortReverse {
		sortModeStr += " (reversed)"
	}

	sortInfo := m.styles.Tertiary.Render("current sort: ") +
		m.styles.Accent.Render(sortModeStr)
	b.WriteString(sortInfo)
	b.WriteString("\n")

	themeInfo := m.styles.Tertiary.Render("active theme: ") +
		m.styles.Accent.Render(m.theme.Name)
	b.WriteString(themeInfo)

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
