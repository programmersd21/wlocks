package ui

import (
	"fmt"
	"os"
	"syscall"
)

// killProcess sends a signal to kill a process
// force=false sends SIGTERM (graceful), force=true sends SIGKILL (immediate)
func killProcess(pid int, force bool) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	signal := syscall.SIGTERM
	if force {
		signal = syscall.SIGKILL
	}

	err = process.Signal(signal)
	if err != nil {
		return fmt.Errorf("failed to send signal: %w", err)
	}

	return nil
}

// pauseProcess sends SIGSTOP (stop = true) or SIGCONT (stop = false) to a process.
func pauseProcess(pid int, stop bool) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("process not found: %w", err)
	}

	signal := syscall.SIGCONT
	if stop {
		signal = syscall.SIGSTOP
	}

	err = process.Signal(signal)
	if err != nil {
		return fmt.Errorf("failed to send signal: %w", err)
	}

	return nil
}
