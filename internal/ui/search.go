package ui

import (
	"fmt"
	"strings"
	"time"
	"wlocks/internal/proc"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
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
			m.detailTab = 0
			m.killConfirm = false
			m.pauseConfirm = false
			m.forceConfirm = false
			m.detailFiles = proc.GetProcessOpenFiles(m.detailLock.Process.PID)
			m.detailEnv = proc.GetProcessEnv(m.detailLock.Process.PID)
			m.fadeAnim.FadeIn(200 * time.Millisecond)
		}
		return m, animTickCmd()

	case Matches(key, m.keys.Down):
		if len(m.searchResults) > 0 {
			newIndex := min(m.selectedIndex+1, len(m.searchResults)-1)
			if newIndex != m.selectedIndex {
				m.selectedIndex = newIndex
				m.ensureSearchVisible()
			}
		}
		return m, animTickCmd()

	case Matches(key, m.keys.Up):
		newIndex := max(0, m.selectedIndex-1)
		if newIndex != m.selectedIndex {
			m.selectedIndex = newIndex
			m.ensureSearchVisible()
		}
		return m, animTickCmd()

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
	m.scrollOffset = 0
	m.scrollAnim = NewScrollAnimation(0)

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
	styles := m.currentStyles()

	b.WriteString("\n")

	searchPrompt := styles.Accent.Render(" / ") + styles.Primary.Render(m.searchQuery)
	searchPrompt += styles.Accent.Render("▋")

	b.WriteString("  " + searchPrompt + "\n\n")

	if len(m.searchResults) == 0 {
		b.WriteString("\n")
		empty := styles.Tertiary.Render("no matching processes found.")
		b.WriteString("  " + empty + "\n")
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
	maxVisible := m.height - overhead
	if maxVisible < 1 {
		maxVisible = 1
	}

	animatedScroll := m.scrollAnim.Update()
	if animatedScroll < 0 {
		animatedScroll = 0
	}
	if animatedScroll >= len(m.searchResults) {
		animatedScroll = max(0, len(m.searchResults)-1)
	}
	start := animatedScroll
	end := min(start+maxVisible, len(m.searchResults))

	for i := start; i < end; i++ {
		lock := m.searchResults[i]
		row := m.renderLockRow(lock, i == m.selectedIndex)
		b.WriteString("  " + row + "\n")
	}

	if len(m.searchResults) > maxVisible {
		b.WriteString("\n")
		progressText := fmt.Sprintf("  %d - %d of %d matches (j/k to scroll)", start+1, end, len(m.searchResults))
		b.WriteString(styles.Tertiary.Render(progressText))
	}

	return b.String()
}

func (m *Model) ensureSearchVisible() {
	overhead := 8
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
