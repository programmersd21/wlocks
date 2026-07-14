package ui

import (
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) handleDetailKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case Matches(key, m.keys.Esc):
		m.mode = modeStatic
		m.killConfirm = false
		return m, nil

	case Matches(key, m.keys.Down):
		m.detailScroll++
		return m, nil

	case Matches(key, m.keys.Up):
		m.detailScroll = max(0, m.detailScroll-1)
		return m, nil

	case Matches(key, m.keys.Kill):
		if !m.killConfirm {
			m.killConfirm = true
			return m, m.killTimeoutCmd()
		}
		m.mode = modeStatic
		m.killConfirm = false
		return m, nil

	default:
		if m.killConfirm {
			m.killConfirm = false
			return m, nil
		}
	}

	return m, nil
}

func (m *Model) killTimeoutCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(2 * time.Second)
		return killTimeoutMsg{}
	}
}

func (m *Model) viewDetail() string {
	if m.detailLock == nil {
		return m.styles.Tertiary.Render("no details available.")
	}

	var b strings.Builder

	lock := m.detailLock
	proc := lock.Process

	name := proc.Name
	if name == "" && len(proc.Cmdline) > 0 {
		name = proc.Cmdline[0]
	}
	header := m.styles.Primary.Render(name)
	b.WriteString(header)
	b.WriteString("\n\n")

	rows := []struct {
		label string
		value string
	}{
		{"pid", fmt.Sprintf("%d", proc.PID)},
		{"command", proc.GetCmdlineString()},
		{"cwd", proc.GetCWDDisplay()},
		{"open fds", fmt.Sprintf("%d", proc.CountOpenFDs())},
		{"opened", formatDuration(lock.Duration) + " ago"},
	}

	labelWidth := 12
	for _, row := range rows {
		label := m.styles.Tertiary.Render(row.label)
		label = lipgloss.NewStyle().Width(labelWidth).Render(label)

		value := m.styles.Primary.Render(row.value)

		line := lipgloss.JoinHorizontal(lipgloss.Left, label, value)
		b.WriteString(line)
		b.WriteString("\n")
	}

	if m.killConfirm {
		b.WriteString("\n")
		prompt := m.styles.Danger.Render("kill - press again to confirm")
		b.WriteString(prompt)
	}

	return b.String()
}
