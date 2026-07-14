package ui

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var paletteCommands = []string{
	"search",
	"refresh",
	"cycle theme",
	"cycle sort",
	"reverse sort",
	"show help",
	"show statistics",
}

func (m *Model) handlePaletteKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case Matches(key, m.keys.Esc):
		m.mode = modeStatic
		return m, nil

	case Matches(key, m.keys.Down):
		m.paletteIndex = min(m.paletteIndex+1, len(paletteCommands)-1)
		return m, nil

	case Matches(key, m.keys.Up):
		m.paletteIndex = max(0, m.paletteIndex-1)
		return m, nil

	case Matches(key, m.keys.Enter):
		cmd := m.executePaletteCommand(m.paletteIndex)
		m.mode = modeStatic
		return m, cmd

	default:
		m.mode = modeStatic
		return m, nil
	}
}

func (m *Model) executePaletteCommand(index int) tea.Cmd {
	if index >= len(paletteCommands) {
		return nil
	}

	switch paletteCommands[index] {
	case "search":
		m.mode = modeSearch
		m.searchQuery = ""
		m.updateSearchResults()
		return animTickCmd()

	case "refresh":
		m.setStatus("refreshing...")
		return tea.Batch(m.scanCmd(), statusClearCmd())

	case "cycle theme":
		nextTheme := NextTheme(m.theme.Name)
		m.SetTheme(nextTheme)
		m.setStatus("theme: " + nextTheme)
		return statusClearCmd()

	case "cycle sort":
		m.cycleSortMode()
		return statusClearCmd()

	case "reverse sort":
		m.sortReverse = !m.sortReverse
		m.sortLocks()
		if m.sortReverse {
			m.setStatus("sort reversed")
		} else {
			m.setStatus("sort normal")
		}
		return statusClearCmd()

	case "show help":
		m.mode = modeHelp
		m.detailScroll = 0
		return animTickCmd()

	case "show statistics":
		m.mode = modeStats
		m.detailScroll = 0
		return animTickCmd()
	}

	return nil
}

func (m *Model) viewPalette() string {
	var items []string
	for i, cmd := range paletteCommands {
		var line string
		if i == m.paletteIndex {
			line = m.styles.Accent.Render("▌ ") + m.styles.Accent.Render(cmd)
		} else {
			line = "  " + m.styles.Secondary.Render(cmd)
		}
		items = append(items, line)
	}

	content := strings.Join(items, "\n")

	maxWidth := 30
	palette := lipgloss.NewStyle().
		Width(maxWidth).
		Render(content)

	vPad := max(0, (m.height-len(items)-2)/2)
	hPad := max(0, (m.width-maxWidth)/2)

	padded := lipgloss.NewStyle().
		PaddingTop(vPad).
		PaddingLeft(hPad).
		Render(palette)

	return padded
}
