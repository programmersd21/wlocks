package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type Theme struct {
	Name          string
	TextPrimary   string
	TextSecondary string
	TextTertiary  string
	TextGhost     string
	Accent        string
	AccentDim     string
	Positive      string
	Warning       string
	Danger        string
}

var themes = map[string]*Theme{
	"default": {
		Name:          "default",
		TextPrimary:   "#f9fafb",
		TextSecondary: "#e5e7eb",
		TextTertiary:  "#9ca3af",
		TextGhost:     "#4b5563",
		Accent:        "#7c9eff",
		AccentDim:     "#4f64c8",
		Positive:      "#34d399",
		Warning:       "#fbbf24",
		Danger:        "#f87171",
	},
	"tokyo": {
		Name:          "tokyo",
		TextPrimary:   "#c0caf5",
		TextSecondary: "#a9b1d6",
		TextTertiary:  "#565f89",
		TextGhost:     "#3b4261",
		Accent:        "#7aa2f7",
		AccentDim:     "#3d59a1",
		Positive:      "#9ece6a",
		Warning:       "#e0af68",
		Danger:        "#f7768e",
	},
	"catppuccin": {
		Name:          "catppuccin",
		TextPrimary:   "#cdd6f4",
		TextSecondary: "#a6adc8",
		TextTertiary:  "#585b70",
		TextGhost:     "#45475a",
		Accent:        "#cba6f7",
		AccentDim:     "#89b4fa",
		Positive:      "#a6e3a1",
		Warning:       "#f9e2af",
		Danger:        "#f38ba8",
	},
	"everforest": {
		Name:          "everforest",
		TextPrimary:   "#d3c6aa",
		TextSecondary: "#a7c080",
		TextTertiary:  "#859289",
		TextGhost:     "#4a555b",
		Accent:        "#a7c080",
		AccentDim:     "#7fbbb3",
		Positive:      "#a7c080",
		Warning:       "#dbbc7f",
		Danger:        "#e67e80",
	},
	"nord": {
		Name:          "nord",
		TextPrimary:   "#eceff4",
		TextSecondary: "#d8dee9",
		TextTertiary:  "#4c566a",
		TextGhost:     "#3b4252",
		Accent:        "#88c0d0",
		AccentDim:     "#81a1c1",
		Positive:      "#a3be8c",
		Warning:       "#ebcb8b",
		Danger:        "#bf616a",
	},
	"gruvbox": {
		Name:          "gruvbox",
		TextPrimary:   "#fbf1c7",
		TextSecondary: "#d5c4a1",
		TextTertiary:  "#7c6f64",
		TextGhost:     "#504945",
		Accent:        "#fabd2f",
		AccentDim:     "#fe8019",
		Positive:      "#b8bb26",
		Warning:       "#fabd2f",
		Danger:        "#fb4934",
	},
	"apple": {
		Name:          "apple",
		TextPrimary:   "#ffffff",
		TextSecondary: "#e4e4e7",
		TextTertiary:  "#a1a1aa",
		TextGhost:     "#52525b",
		Accent:        "#ff9f0a",
		AccentDim:     "#bf7600",
		Positive:      "#30d158",
		Warning:       "#ffd60a",
		Danger:        "#ff453a",
	},
	"linear": {
		Name:          "linear",
		TextPrimary:   "#f7f8f8",
		TextSecondary: "#ced4da",
		TextTertiary:  "#868e96",
		TextGhost:     "#495057",
		Accent:        "#5e6ad2",
		AccentDim:     "#4852b1",
		Positive:      "#4cb3d4",
		Warning:       "#f59f00",
		Danger:        "#f03e3e",
	},
	"neon": {
		Name:          "neon",
		TextPrimary:   "#ffffff",
		TextSecondary: "#e2e8f0",
		TextTertiary:  "#94a3b8",
		TextGhost:     "#475569",
		Accent:        "#ff007f",
		AccentDim:     "#d9006c",
		Positive:      "#39ff14",
		Warning:       "#ffd700",
		Danger:        "#ff3333",
	},
}

func ThemeNames() []string {
	return []string{"default", "tokyo", "catppuccin", "everforest", "nord", "gruvbox", "apple", "linear", "neon"}
}

func GetTheme(name string) *Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["default"]
}

func NextTheme(current string) string {
	names := ThemeNames()
	for i, name := range names {
		if name == current {
			return names[(i+1)%len(names)]
		}
	}
	return names[0]
}

type Styles struct {
	Primary   lipgloss.Style
	Secondary lipgloss.Style
	Tertiary  lipgloss.Style
	Ghost     lipgloss.Style
	Accent    lipgloss.Style
	AccentDim lipgloss.Style
	Positive  lipgloss.Style
	Warning   lipgloss.Style
	Danger    lipgloss.Style
	Bar       lipgloss.Style
}

func NewStyles(theme *Theme) *Styles {
	return &Styles{
		Primary:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextPrimary)),
		Secondary: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextSecondary)),
		Tertiary:  lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextTertiary)),
		Ghost:     lipgloss.NewStyle().Foreground(lipgloss.Color(theme.TextGhost)),
		Accent:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent)),
		AccentDim: lipgloss.NewStyle().Foreground(lipgloss.Color(theme.AccentDim)),
		Positive:  lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Positive)),
		Warning:   lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Warning)),
		Danger:    lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Danger)),
		Bar:       lipgloss.NewStyle().Foreground(lipgloss.Color(theme.Accent)),
	}
}

func parseHex(s string) (r, g, b float64) {
	s = strings.TrimPrefix(s, "#")
	if len(s) == 3 {
		s = string([]byte{s[0], s[0], s[1], s[1], s[2], s[2]})
	}
	if len(s) != 6 {
		return 0, 0, 0
	}
	var rgb int64
	if _, err := fmt.Sscanf(s, "%x", &rgb); err != nil {
		return 0, 0, 0
	}
	r = float64((rgb >> 16) & 0xFF)
	g = float64((rgb >> 8) & 0xFF)
	b = float64(rgb & 0xFF)
	return r, g, b
}

func formatHex(r, g, b float64) string {
	// clamp to 0-255 range
	rc := int(maxFloat(0, minFloat(255, r)) + 0.5)
	gc := int(maxFloat(0, minFloat(255, g)) + 0.5)
	bc := int(maxFloat(0, minFloat(255, b)) + 0.5)
	return fmt.Sprintf("#%02x%02x%02x", rc, gc, bc)
}

func interpolateColor(c1, c2 string, t float64) string {
	r1, g1, b1 := parseHex(c1)
	r2, g2, b2 := parseHex(c2)
	r := r1 + (r2-r1)*t
	g := g1 + (g2-g1)*t
	b := b1 + (b2-b1)*t
	return formatHex(r, g, b)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
