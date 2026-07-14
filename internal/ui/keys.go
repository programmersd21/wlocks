package ui

// KeyMap holds all key bindings for wlocks.
type KeyMap struct {
	Up             []string
	Down           []string
	Enter          []string
	Esc            []string
	Search         []string
	Refresh        []string
	Kill           []string
	Tree           []string
	ThemeCycle     []string
	CommandPalette []string
	Quit           []string
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:             []string{"k", "up"},   // k or up arrow
		Down:           []string{"j", "down"}, // j or down arrow
		Enter:          []string{"enter"},     // enter
		Esc:            []string{"esc"},       // escape
		Search:         []string{"/"},
		Refresh:        []string{"r"},
		Kill:           []string{"K"}, // capital K only, to avoid collision with down
		Tree:           []string{"t"},
		ThemeCycle:     []string{"T"}, // capital T
		CommandPalette: []string{"?"},
		Quit:           []string{"q", "ctrl+c"},
	}
}

// Matches checks if a key message matches any binding in the list.
func Matches(key string, bindings []string) bool {
	for _, binding := range bindings {
		if key == binding {
			return true
		}
	}
	return false
}
