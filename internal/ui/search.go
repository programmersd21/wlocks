package ui

import (
	"strings"
	"wlocks/internal/proc"

	tea "charm.land/bubbletea/v2"
	"github.com/sahilm/fuzzy"
)

type candidate struct {
	index int
	text  string
}

type candidateSource struct {
	items []candidate
}

func (s candidateSource) String(i int) string {
	return s.items[i].text
}

func (s candidateSource) Len() int {
	return len(s.items)
}

func (m *Model) handleSearchKey(key string) (tea.Model, tea.Cmd) {
	switch {
	case Matches(key, m.keys.Esc):
		m.mode = modeStatic
		m.searchQuery = ""
		return m, nil

	case Matches(key, m.keys.Enter):
		if m.selectedIndex < len(m.searchResults) {
			m.detailLock = m.searchResults[m.selectedIndex]
			m.mode = modeDetail
			m.detailScroll = 0
			m.killConfirm = false
		}
		return m, nil

	case Matches(key, m.keys.Down):
		if len(m.searchResults) > 0 {
			m.selectedIndex = min(m.selectedIndex+1, len(m.searchResults)-1)
		}
		return m, nil

	case Matches(key, m.keys.Up):
		m.selectedIndex = max(0, m.selectedIndex-1)
		return m, nil

	case key == "backspace" || key == "\x7f":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
			m.updateSearchResults()
		}
		return m, nil

	default:
		if len(key) == 1 && key[0] >= 32 && key[0] <= 126 {
			m.searchQuery += key
			m.updateSearchResults()
		}
		return m, nil
	}
}

func (m *Model) updateSearchResults() {
	m.selectedIndex = 0

	if m.searchQuery == "" {
		m.searchResults = m.locks
		return
	}

	candidates := make([]candidate, len(m.locks))
	for i, lock := range m.locks {
		parts := []string{
			lock.Process.Name,
			lock.Process.GetCmdlineString(),
			lock.FD.Path,
		}
		candidates[i] = candidate{
			index: i,
			text:  strings.Join(parts, " "),
		}
	}

	matches := fuzzy.FindFrom(m.searchQuery, candidateSource{candidates})

	m.searchResults = make([]*proc.LockInfo, len(matches))
	for i, match := range matches {
		m.searchResults[i] = m.locks[candidates[match.Index].index]
	}
}

func (m *Model) viewSearch() string {
	var b strings.Builder

	prompt := m.styles.Accent.Render("/") + m.styles.Primary.Render(m.searchQuery)
	b.WriteString(prompt)
	b.WriteString("\n\n")

	if len(m.searchResults) == 0 {
		empty := m.styles.Tertiary.Render("no matches.")
		b.WriteString(empty)
		b.WriteString("\n")
		return b.String()
	}

	maxVisible := m.height - 5
	if maxVisible < 1 {
		maxVisible = 1
	}

	start := 0
	end := min(start+maxVisible, len(m.searchResults))

	for i := start; i < end; i++ {
		lock := m.searchResults[i]
		row := m.renderLockRow(lock, i == m.selectedIndex)
		b.WriteString(row)
		b.WriteString("\n")
	}

	return b.String()
}
