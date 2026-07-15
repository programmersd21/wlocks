package ui

import (
	"fmt"
	"os"
	"strings"
	"time"
	"wlocks/internal/proc"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *Model) handleDetailKey(key string) (tea.Model, tea.Cmd) {
	// Reset other confirmations if a different control key is pressed
	resetConf := func() {
		m.killConfirm = false
		m.forceConfirm = false
		m.pauseConfirm = false
	}

	switch {
	case Matches(key, m.keys.Esc):
		m.mode = modeStatic
		resetConf()
		return m, nil

	case Matches(key, m.keys.Left):
		m.detailTab = max(0, m.detailTab-1)
		m.detailScroll = 0
		resetConf()
		return m, animTickCmd()

	case Matches(key, m.keys.Right):
		m.detailTab = min(2, m.detailTab+1)
		m.detailScroll = 0
		resetConf()
		return m, animTickCmd()

	case Matches(key, m.keys.Down):
		m.detailScroll++
		return m, animTickCmd()

	case Matches(key, m.keys.Up):
		m.detailScroll = max(0, m.detailScroll-1)
		return m, animTickCmd()

	case Matches(key, m.keys.Kill):
		if m.detailLock == nil {
			return m, nil
		}
		if !m.killConfirm {
			resetConf()
			m.killConfirm = true
			return m, m.killTimeoutCmd()
		}
		// Confirmed
		cmd := m.killProcessCmd(m.detailLock.Process.PID)
		resetConf()
		return m, cmd

	case Matches(key, m.keys.KillForce):
		if m.detailLock == nil {
			return m, nil
		}
		if !m.forceConfirm {
			resetConf()
			m.forceConfirm = true
			return m, m.killTimeoutCmd()
		}
		// Confirmed
		cmd := m.killForceProcessCmd(m.detailLock.Process.PID)
		resetConf()
		return m, cmd

	case Matches(key, m.keys.PauseToggle):
		if m.detailLock == nil {
			return m, nil
		}
		if !m.pauseConfirm {
			resetConf()
			m.pauseConfirm = true
			return m, m.killTimeoutCmd()
		}
		// Confirmed
		state := proc.GetProcessState(m.detailLock.Process.PID)
		isStopped := strings.Contains(strings.ToLower(state), "stopped")
		cmd := m.pauseProcessCmd(m.detailLock.Process.PID, !isStopped)
		resetConf()
		return m, cmd

	default:
		resetConf()
	}

	return m, nil
}

func (m *Model) killTimeoutCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second) // 3 seconds grace timeout
		return killTimeoutMsg{}
	}
}

func (m *Model) viewDetail() string {
	styles := m.currentStyles()
	if m.detailLock == nil {
		return styles.Tertiary.Render("no details available.")
	}

	var b strings.Builder
	lock := m.detailLock
	procInfo := lock.Process

	b.WriteString("\n")

	headerText := fmt.Sprintf("process %d", procInfo.PID)
	pathDisplay := styles.Primary.Bold(true).Render(headerText)
	b.WriteString("  " + pathDisplay + "\n\n")

	tabs := []string{"info", "files", "env"}
	var renderedTabs []string
	for i, tab := range tabs {
		if i == m.detailTab {
			renderedTabs = append(renderedTabs, styles.Accent.Render(fmt.Sprintf("  ▸ %s", tab)))
		} else {
			renderedTabs = append(renderedTabs, styles.Ghost.Render(fmt.Sprintf("    %s", tab)))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Left, renderedTabs...) + "\n\n")

	// 3. Tab contents
	switch m.detailTab {
	case 0:
		m.viewDetailTabInfo(&b, styles, lock)
	case 1:
		m.viewDetailTabFiles(&b, styles)
	case 2:
		m.viewDetailTabEnv(&b, styles)
	}

	// 4. ephemaral action prompt confirm
	if m.killConfirm {
		b.WriteString("\n")
		prompt := styles.Danger.Bold(true).Render("  ▸ press K again to confirm SIGTERM (graceful kill)")
		b.WriteString(prompt)
	} else if m.forceConfirm {
		b.WriteString("\n")
		prompt := styles.Danger.Bold(true).Render("  ▸ press F again to confirm SIGKILL (force kill)")
		b.WriteString(prompt)
	} else if m.pauseConfirm {
		b.WriteString("\n")
		state := proc.GetProcessState(procInfo.PID)
		isStopped := strings.Contains(strings.ToLower(state), "stopped")
		action := "suspend (SIGSTOP)"
		if isStopped {
			action = "resume (SIGCONT)"
		}
		prompt := styles.Warning.Bold(true).Render(fmt.Sprintf("  ▸ press P again to confirm %s", action))
		b.WriteString(prompt)
	}

	return b.String()
}

func (m *Model) viewDetailTabInfo(b *strings.Builder, styles *Styles, lock *proc.LockInfo) {
	procInfo := lock.Process

	name := procInfo.Name
	if name == "" && len(procInfo.Cmdline) > 0 {
		name = procInfo.Cmdline[0]
	}

	ppid := proc.GetParentPID(procInfo.PID)
	parentName := ""
	if ppid > 0 {
		if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", ppid)); err == nil {
			parentName = fmt.Sprintf(" (%s)", strings.TrimSpace(string(data)))
		}
	}

	runState := proc.GetProcessState(procInfo.PID)

	rows := []struct {
		label string
		value string
	}{
		{"name", name},
		{"pid", fmt.Sprintf("%d", procInfo.PID)},
		{"parent pid", fmt.Sprintf("%d%s", ppid, parentName)},
		{"state", runState},
		{"command", procInfo.GetCmdlineString()},
		{"cwd", procInfo.GetCWDDisplay()},
		{"open fds", fmt.Sprintf("%d", procInfo.CountOpenFDs())},
		{"opened target", formatDuration(lock.Duration) + " ago"},
	}

	labelWidth := 15
	for _, row := range rows {
		label := styles.Tertiary.Render(row.label)
		label = lipgloss.NewStyle().Width(labelWidth).Render(label)

		value := styles.Primary.Render(row.value)
		// Highlight state values
		if row.label == "state" {
			if strings.Contains(strings.ToLower(row.value), "stopped") {
				value = styles.Warning.Bold(true).Render(row.value)
			} else {
				value = styles.Positive.Render(row.value)
			}
		}

		maxValWidth := m.width - labelWidth - 6
		if maxValWidth > 10 && lipgloss.Width(value) > maxValWidth {
			value = truncate(row.value, maxValWidth)
			value = styles.Primary.Render(value)
		}

		line := lipgloss.JoinHorizontal(lipgloss.Left, label, value)
		b.WriteString("  " + line + "\n")
	}
}

func (m *Model) viewDetailTabFiles(b *strings.Builder, styles *Styles) {
	if len(m.detailFiles) == 0 {
		b.WriteString("  " + styles.Tertiary.Render("no open files detected (or permission denied).") + "\n")
		return
	}

	maxVisible := m.height - 10
	if maxVisible < 1 {
		maxVisible = 1
	}

	if m.detailScroll >= len(m.detailFiles) {
		m.detailScroll = max(0, len(m.detailFiles)-1)
	}

	start := m.detailScroll
	end := min(start+maxVisible, len(m.detailFiles))

	for i := start; i < end; i++ {
		filePath := m.detailFiles[i]
		// Collapse home directory path
		home, err := os.UserHomeDir()
		if err == nil && strings.HasPrefix(filePath, home) {
			filePath = "~" + strings.TrimPrefix(filePath, home)
		}

		maxPathLen := m.width - 8
		if maxPathLen > 10 && len(filePath) > maxPathLen {
			filePath = "..." + filePath[len(filePath)-maxPathLen:]
		}

		b.WriteString("    " + styles.Secondary.Render(filePath) + "\n")
	}

	if len(m.detailFiles) > maxVisible {
		b.WriteString("\n")
		progressText := fmt.Sprintf("  showing %d - %d of %d open files (j/k to scroll)", start+1, end, len(m.detailFiles))
		b.WriteString(styles.Tertiary.Render(progressText))
	}
}

func (m *Model) viewDetailTabEnv(b *strings.Builder, styles *Styles) {
	if len(m.detailEnv) == 0 {
		b.WriteString("  " + styles.Tertiary.Render("no environment variables readable.") + "\n")
		return
	}

	maxVisible := m.height - 10
	if maxVisible < 1 {
		maxVisible = 1
	}

	if m.detailScroll >= len(m.detailEnv) {
		m.detailScroll = max(0, len(m.detailEnv)-1)
	}

	start := m.detailScroll
	end := min(start+maxVisible, len(m.detailEnv))

	for i := start; i < end; i++ {
		line := m.detailEnv[i]
		parts := strings.SplitN(line, "=", 2)

		var renderedLine string
		if len(parts) == 2 {
			// Highlight key differently from value
			key := styles.Secondary.Bold(true).Render(parts[0])
			val := styles.Primary.Render(parts[1])

			renderedLine = key + styles.Ghost.Render("=") + val
		} else {
			renderedLine = styles.Primary.Render(line)
		}

		// Truncate to terminal width
		maxLen := m.width - 6
		if maxLen > 10 && lipgloss.Width(renderedLine) > maxLen {
			renderedLine = truncate(line, maxLen)
			// re-parse and highlight if possible
			partsTrunc := strings.SplitN(renderedLine, "=", 2)
			if len(partsTrunc) == 2 {
				renderedLine = styles.Secondary.Bold(true).Render(partsTrunc[0]) + styles.Ghost.Render("=") + styles.Primary.Render(partsTrunc[1])
			} else {
				renderedLine = styles.Primary.Render(renderedLine)
			}
		}

		b.WriteString("    " + renderedLine + "\n")
	}

	if len(m.detailEnv) > maxVisible {
		b.WriteString("\n")
		progressText := fmt.Sprintf("  showing %d - %d of %d environment variables (j/k to scroll)", start+1, end, len(m.detailEnv))
		b.WriteString(styles.Tertiary.Render(progressText))
	}
}
