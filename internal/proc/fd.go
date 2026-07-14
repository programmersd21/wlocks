package proc

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FDInfo represents a single file descriptor open by a process.
type FDInfo struct {
	PID      int
	FD       int
	Path     string
	Mode     FDMode
	OpenedAt int64 // Unix timestamp, best-effort from stat
}

// FDMode represents how a file is opened.
type FDMode int

const (
	FDModeRead FDMode = iota
	FDModeWrite
	FDModeReadWrite
	FDModeUnknown
)

func (m FDMode) String() string {
	switch m {
	case FDModeRead:
		return "read"
	case FDModeWrite:
		return "write"
	case FDModeReadWrite:
		return "write" // treat read+write as write for display purposes
	default:
		return "unknown"
	}
}

// ResolveFD reads /proc/[pid]/fd/[fd] and returns the target path.
// Returns empty string if the fd doesn't exist or can't be read.
func ResolveFD(pid, fd int) string {
	link := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
	target, err := os.Readlink(link)
	if err != nil {
		return ""
	}
	return target
}

// GetFDMode reads /proc/[pid]/fdinfo/[fd] and parses the flags field
// to determine the correct O_ACCMODE (read/write mode).
// This is the only correct way per the spec - don't guess from symlink.
func GetFDMode(pid, fd int) FDMode {
	fdinfoPath := fmt.Sprintf("/proc/%d/fdinfo/%d", pid, fd)
	data, err := os.ReadFile(fdinfoPath)
	if err != nil {
		return FDModeUnknown
	}

	// Parse the flags: line
	// Example: flags:	0100002
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "flags:") {
			fields := strings.Fields(line)
			if len(fields) < 2 {
				continue
			}
			flagsStr := strings.TrimPrefix(fields[1], "0")
			if flagsStr == "" {
				flagsStr = "0"
			}
			flags, err := strconv.ParseInt(flagsStr, 8, 64)
			if err != nil {
				continue
			}

			// O_ACCMODE mask is 0x3 (lowest 2 bits)
			// O_RDONLY = 0, O_WRONLY = 1, O_RDWR = 2
			accmode := flags & 0x3
			switch accmode {
			case 0:
				return FDModeRead
			case 1:
				return FDModeWrite
			case 2:
				return FDModeReadWrite
			default:
				return FDModeUnknown
			}
		}
	}

	return FDModeUnknown
}

// GetFDInfo returns full information about a single file descriptor.
func GetFDInfo(pid, fd int) *FDInfo {
	path := ResolveFD(pid, fd)
	if path == "" {
		return nil
	}

	mode := GetFDMode(pid, fd)

	// Try to get mtime from the fd link itself as a proxy for "opened at"
	// This is best-effort; procfs timestamps aren't perfect for this
	link := fmt.Sprintf("/proc/%d/fd/%d", pid, fd)
	var openedAt int64
	if stat, err := os.Lstat(link); err == nil {
		openedAt = stat.ModTime().Unix()
	}

	return &FDInfo{
		PID:      pid,
		FD:       fd,
		Path:     path,
		Mode:     mode,
		OpenedAt: openedAt,
	}
}

// ListFDs returns all file descriptors for a given PID.
// Silently skips any that cannot be read (permission denied, etc).
func ListFDs(pid int) []int {
	fdDir := fmt.Sprintf("/proc/%d/fd", pid)
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return nil
	}

	fds := make([]int, 0, len(entries))
	for _, entry := range entries {
		if fd, err := strconv.Atoi(entry.Name()); err == nil {
			fds = append(fds, fd)
		}
	}
	return fds
}

// MatchesPath checks if the given FDInfo resolves to the target path.
// Both paths are resolved via EvalSymlinks to handle symlinks/bind mounts.
func MatchesPath(fdInfo *FDInfo, targetPath string) bool {
	if fdInfo == nil || fdInfo.Path == "" {
		return false
	}

	// Resolve both to absolute paths
	targetResolved, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		// If we can't resolve target, fall back to string comparison
		targetResolved = targetPath
	}
	targetResolved, _ = filepath.Abs(targetResolved)

	fdResolved, err := filepath.EvalSymlinks(fdInfo.Path)
	if err != nil {
		fdResolved = fdInfo.Path
	}
	fdResolved, _ = filepath.Abs(fdResolved)

	return targetResolved == fdResolved
}
