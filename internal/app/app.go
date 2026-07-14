package app

import (
	"fmt"
	"os"
	"time"
	"wlocks/internal/config"
	"wlocks/internal/proc"
	"wlocks/internal/ui"

	tea "charm.land/bubbletea/v2"
	"golang.org/x/term"
)

// Options holds configuration for running the app.
type Options struct {
	TargetPath string
	Theme      string
	Debug      bool
	JSON       bool
}

// Run starts the wlocks application with the given options.
func Run(opts Options) error {
	// Load persisted config
	cfg, err := config.Load()
	if err != nil {
		// Non-fatal; continue with defaults
		cfg = config.Default()
	}

	// Command-line theme overrides config
	themeName := cfg.Theme
	if opts.Theme != "" {
		themeName = opts.Theme
	}

	// If output is not a TTY, run in plain text mode
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return runPlainText(opts, themeName)
	}

	// Create model
	model := ui.NewModel(opts.TargetPath, themeName, opts.Debug)

	// Start Bubble Tea program
	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run UI: %w", err)
	}

	return nil
}

// runPlainText runs in non-interactive mode, printing plain text output.
func runPlainText(opts Options, themeName string) error {
	if opts.TargetPath == "" {
		return fmt.Errorf("path required for non-interactive mode")
	}

	result := proc.ScanForPath(opts.TargetPath, opts.Debug)

	// Print path
	fmt.Println(opts.TargetPath)
	fmt.Println()

	if len(result.Locks) == 0 {
		fmt.Println("nothing is holding this file.")
		return nil
	}

	fmt.Println("open by")
	fmt.Println()

	// Print each lock
	for _, lock := range result.Locks {
		name := lock.Process.Name
		if name == "" && len(lock.Process.Cmdline) > 0 {
			name = lock.Process.Cmdline[0]
		}

		mode := lock.FD.Mode.String()
		duration := formatDurationPlain(lock.Duration)

		// Plain aligned output
		fmt.Printf("  %-20s %-12s %s\n", truncate(name, 20), mode, duration)
	}

	if opts.Debug && result.PermissionDenied > 0 {
		fmt.Println()
		fmt.Printf("%d processes hidden - insufficient permissions\n", result.PermissionDenied)
	}

	return nil
}

// formatDurationPlain formats duration for plain text output.
func formatDurationPlain(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh", int(d.Hours()))
	}
	return fmt.Sprintf("%dd", int(d.Hours()/24))
}

// truncate truncates a string to n runes.
func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
