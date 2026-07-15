package proc

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ProcessInfo holds metadata about a process.
type ProcessInfo struct {
	PID     int
	Name    string   // from /proc/[pid]/comm
	Cmdline []string // from /proc/[pid]/cmdline, split on null
	CWD     string   // from /proc/[pid]/cwd
	Exe     string   // from /proc/[pid]/exe
}

// GetProcessInfo retrieves metadata for a given PID.
// Handles EACCES gracefully - fields will be empty if unreadable.
func GetProcessInfo(pid int) *ProcessInfo {
	info := &ProcessInfo{
		PID: pid,
	}

	// Read comm (process name)
	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/comm", pid)); err == nil {
		info.Name = strings.TrimSpace(string(data))
	}

	// Read cmdline (null-separated)
	if data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid)); err == nil {
		// Split on null bytes
		parts := bytes.Split(data, []byte{0})
		for _, part := range parts {
			if len(part) > 0 {
				info.Cmdline = append(info.Cmdline, string(part))
			}
		}
	}

	// If cmdline is empty (kernel thread or zombie), fall back to comm in brackets
	if len(info.Cmdline) == 0 && info.Name != "" {
		info.Cmdline = []string{fmt.Sprintf("[%s]", info.Name)}
	}

	// Read cwd
	if cwd, err := os.Readlink(fmt.Sprintf("/proc/%d/cwd", pid)); err == nil {
		info.CWD = cwd
	}

	// Read exe
	if exe, err := os.Readlink(fmt.Sprintf("/proc/%d/exe", pid)); err == nil {
		info.Exe = exe
	}

	return info
}

// GetAllPIDs returns all numeric PIDs from /proc.
func GetAllPIDs() []int {
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil
	}

	pids := make([]int, 0, 256)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if pid, err := strconv.Atoi(entry.Name()); err == nil {
			pids = append(pids, pid)
		}
	}
	return pids
}

// GetCmdlineString returns the cmdline as a single string for display.
func (p *ProcessInfo) GetCmdlineString() string {
	if len(p.Cmdline) == 0 {
		return ""
	}
	return strings.Join(p.Cmdline, " ")
}

// GetCWDDisplay returns a display-friendly CWD (tilde-collapsed if home dir).
func (p *ProcessInfo) GetCWDDisplay() string {
	if p.CWD == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(p.CWD, home) {
		return "~" + strings.TrimPrefix(p.CWD, home)
	}
	return p.CWD
}

// CountOpenFDs returns the number of open file descriptors for this process.
func (p *ProcessInfo) CountOpenFDs() int {
	fdDir := fmt.Sprintf("/proc/%d/fd", p.PID)
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return 0
	}
	return len(entries)
}

// IsAlive checks if the process still exists.
func (p *ProcessInfo) IsAlive() bool {
	_, err := os.Stat(filepath.Join("/proc", strconv.Itoa(p.PID)))
	return err == nil
}

// GetParentPID retrieves the parent PID of a given process.
func GetParentPID(pid int) int {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if err != nil {
		return 0
	}
	idx := strings.LastIndex(string(data), ")")
	if idx == -1 || idx+2 >= len(data) {
		return 0
	}
	fields := strings.Fields(string(data[idx+2:]))
	if len(fields) < 2 {
		return 0
	}
	ppid, _ := strconv.Atoi(fields[1])
	return ppid
}

// GetProcessEnv retrieves all environment variables for a process.
func GetProcessEnv(pid int) []string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/environ", pid))
	if err != nil {
		return nil
	}
	parts := bytes.Split(data, []byte{0})
	env := make([]string, 0, len(parts))
	for _, part := range parts {
		if len(part) > 0 {
			env = append(env, string(part))
		}
	}
	return env
}

// GetProcessOpenFiles lists all non-device open files for a process.
func GetProcessOpenFiles(pid int) []string {
	fds := ListFDs(pid)
	if fds == nil {
		return nil
	}
	filesMap := make(map[string]bool)
	for _, fd := range fds {
		path := ResolveFD(pid, fd)
		if path != "" && !strings.HasPrefix(path, "/dev/") && !strings.HasPrefix(path, "pipe:") && !strings.HasPrefix(path, "socket:") {
			filesMap[path] = true
		}
	}
	files := make([]string, 0, len(filesMap))
	for f := range filesMap {
		files = append(files, f)
	}
	sort.Strings(files)
	return files
}

// GetProcessState retrieves the state description of a process.
func GetProcessState(pid int) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/status", pid))
	if err != nil {
		return "unknown"
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "State:") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				return strings.Join(fields[1:], " ")
			}
		}
	}
	return "unknown"
}
