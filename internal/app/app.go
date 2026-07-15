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

type Options struct {
	TargetPath string
	Theme      string
	Debug      bool
	JSON       bool
}

func Run(opts Options) error {
	cfg, err := config.Load()
	if err != nil {
		cfg = config.Default()
	}

	themeName := cfg.Theme
	if opts.Theme != "" {
		themeName = opts.Theme
	}

	if !term.IsTerminal(int(os.Stdout.Fd())) {
		return runPlainText(opts, themeName)
	}

	model := ui.NewModel(ui.ModelConfig{
		TargetPath: opts.TargetPath,
		ThemeName:  themeName,
		Debug:      opts.Debug,
		Config:     cfg,
	})

	p := tea.NewProgram(model)
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("failed to run UI: %w", err)
	}

	return nil
}

func runPlainText(opts Options, themeName string) error {
	if opts.TargetPath == "" {
		return fmt.Errorf("path required for non-interactive mode")
	}

	result := proc.ScanForPath(opts.TargetPath, opts.Debug)

	fmt.Println(opts.TargetPath)
	fmt.Println()

	if len(result.Locks) == 0 {
		fmt.Println("nothing is holding this file.")
		return nil
	}

	fmt.Println("open by")
	fmt.Println()

	for _, lock := range result.Locks {
		name := lock.Process.Name
		if name == "" && len(lock.Process.Cmdline) > 0 {
			name = lock.Process.Cmdline[0]
		}

		mode := lock.FD.Mode.String()
		duration := formatDurationPlain(lock.Duration)

		fmt.Printf("  %-20s %-12s %s\n", truncate(name, 20), mode, duration)
	}

	if opts.Debug && result.PermissionDenied > 0 {
		fmt.Println()
		fmt.Printf("%d processes hidden - insufficient permissions\n", result.PermissionDenied)
	}

	return nil
}

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

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
