package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"wlocks/internal/app"
	"wlocks/internal/config"
)

const version = "0.1.0"

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "wlocks - show which processes have which files open\n\n")
		fmt.Fprintf(os.Stderr, "usage:\n")
		fmt.Fprintf(os.Stderr, "  wlocks [path]              show who holds this file\n")
		fmt.Fprintf(os.Stderr, "  wlocks --theme <name>      override theme for this session\n")
		fmt.Fprintf(os.Stderr, "  wlocks --debug             show permission errors and diagnostic info\n")
		fmt.Fprintf(os.Stderr, "  wlocks --version           show version\n")
		fmt.Fprintf(os.Stderr, "  wlocks --help              show this help\n\n")
		fmt.Fprintf(os.Stderr, "themes:\n")
		fmt.Fprintf(os.Stderr, "  default, tokyo, catppuccin, everforest, nord, gruvbox, apple, linear, neon\n")
		fmt.Fprintf(os.Stderr, "  cycle with T key at runtime, set permanently with --theme\n\n")
		fmt.Fprintf(os.Stderr, "keys:\n")
		fmt.Fprintf(os.Stderr, "  j/k or arrows    navigate\n")
		fmt.Fprintf(os.Stderr, "  enter            detail view\n")
		fmt.Fprintf(os.Stderr, "  esc              back / clear search\n")
		fmt.Fprintf(os.Stderr, "  /                search\n")
		fmt.Fprintf(os.Stderr, "  r                refresh\n")
		fmt.Fprintf(os.Stderr, "  s                cycle sort\n")
		fmt.Fprintf(os.Stderr, "  S                reverse sort\n")
		fmt.Fprintf(os.Stderr, "  T                cycle theme\n")
		fmt.Fprintf(os.Stderr, "  ?                help\n")
		fmt.Fprintf(os.Stderr, "  i                statistics\n")
		fmt.Fprintf(os.Stderr, "  ctrl+p           command palette\n")
		fmt.Fprintf(os.Stderr, "  q                quit\n\n")
	}

	var (
		themeFlag   = flag.String("theme", "", "override theme (default, tokyo, catppuccin, everforest, nord, gruvbox, apple, linear, neon)")
		debugFlag   = flag.Bool("debug", false, "enable debug output")
		versionFlag = flag.Bool("version", false, "show version")
	)

	flag.Parse()

	if *versionFlag {
		fmt.Printf("wlocks %s\n", version)
		os.Exit(0)
	}

	var targetPath string
	args := flag.Args()
	if len(args) > 0 {
		targetPath = args[0]
	} else {
		targetPath = "."
	}

	absPath, err := filepath.Abs(targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "invalid path: %s\n", targetPath)
		os.Exit(1)
	}
	targetPath = absPath

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "path does not exist: %s\n", targetPath)
		os.Exit(1)
	}

	if *themeFlag != "" {
		validThemes := map[string]bool{
			"default": true, "tokyo": true, "catppuccin": true,
			"everforest": true, "nord": true, "gruvbox": true,
			"apple": true, "linear": true, "neon": true,
		}
		if !validThemes[*themeFlag] {
			fmt.Fprintf(os.Stderr, "invalid theme: %s\n", *themeFlag)
			fmt.Fprintf(os.Stderr, "valid themes: default, tokyo, catppuccin, everforest, nord, gruvbox, apple, linear, neon\n")
			os.Exit(1)
		}

		cfg := &config.Config{Theme: *themeFlag}
		if err := config.Save(cfg); err != nil {
			if *debugFlag {
				fmt.Fprintf(os.Stderr, "warning: could not save theme preference: %v\n", err)
			}
		}
	}

	opts := app.Options{
		TargetPath: targetPath,
		Theme:      *themeFlag,
		Debug:      *debugFlag,
	}

	if err := app.Run(opts); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
