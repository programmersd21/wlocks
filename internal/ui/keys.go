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
	Help           []string
	Stats          []string
	Sort           []string
	SortReverse    []string
	Quit           []string
}

// DefaultKeyMap returns the default key bindings.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up:             []string{"k", "up"},
		Down:           []string{"j", "down"},
		Enter:          []string{"enter"},
		Esc:            []string{"esc"},
		Search:         []string{"/"},
		Refresh:        []string{"r"},
		Kill:           []string{"K"},
		Tree:           []string{"t"},
		ThemeCycle:     []string{"T"},
		CommandPalette: []string{"ctrl+p"},
		Help:           []string{"?"},
		Stats:          []string{"i"},
		Sort:           []string{"s"},
		SortReverse:    []string{"S"},
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
