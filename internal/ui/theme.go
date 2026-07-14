package ui

import "charm.land/lipgloss/v2"

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
	"aurora": {
		Name:          "aurora",
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
	"citrus": {
		Name:          "citrus",
		TextPrimary:   "#fafaf9",
		TextSecondary: "#d6d3d1",
		TextTertiary:  "#78716c",
		TextGhost:     "#44403c",
		Accent:        "#e4f222",
		AccentDim:     "#a3b10b",
		Positive:      "#84cc16",
		Warning:       "#f59e0b",
		Danger:        "#ef4444",
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
}

func ThemeNames() []string {
	return []string{"aurora", "citrus", "tokyo", "catppuccin", "nord", "gruvbox"}
}

func GetTheme(name string) *Theme {
	if t, ok := themes[name]; ok {
		return t
	}
	return themes["aurora"]
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
