package proc

import (
	"path/filepath"
	"time"
)

// LockInfo represents a complete record of a process holding a file.
type LockInfo struct {
	Process  *ProcessInfo
	FD       *FDInfo
	Duration time.Duration // how long the file has been open
}

// ScanResult holds the outcome of a scan operation.
type ScanResult struct {
	Locks            []*LockInfo
	PermissionDenied int // count of processes we couldn't read
}

// ScanForPath finds all processes that have the given path open.
// This is the core operation for static mode.
func ScanForPath(targetPath string, debug bool) *ScanResult {
	result := &ScanResult{
		Locks: make([]*LockInfo, 0),
	}

	// Resolve target path once upfront
	targetResolved, err := filepath.EvalSymlinks(targetPath)
	if err != nil {
		targetResolved = targetPath
	}
	targetResolved, _ = filepath.Abs(targetResolved)

	pids := GetAllPIDs()
	now := time.Now()

	for _, pid := range pids {
		fds := ListFDs(pid)
		if fds == nil {
			// Couldn't read fd dir - permission denied or process died
			if debug {
				result.PermissionDenied++
			}
			continue
		}

		for _, fd := range fds {
			fdInfo := GetFDInfo(pid, fd)
			if fdInfo == nil {
				continue
			}

			// Check if this fd matches our target
			if !MatchesPath(fdInfo, targetResolved) {
				continue
			}

			// Match found
			procInfo := GetProcessInfo(pid)
			duration := now.Sub(time.Unix(fdInfo.OpenedAt, 0))
			if duration < 0 {
				duration = 0
			}

			result.Locks = append(result.Locks, &LockInfo{
				Process:  procInfo,
				FD:       fdInfo,
				Duration: duration,
			})

			// Only need one fd per process for display
			break
		}
	}

	return result
}

// Snapshot captures all open file descriptors across all processes.
// Used by the live watcher for diffing.
type Snapshot struct {
	Timestamp time.Time
	FDs       map[int]map[int]*FDInfo // pid -> fd -> fdinfo
}

// TakeSnapshot walks /proc and captures all fds.
// This is expensive but necessary since inotify doesn't work on procfs.
func TakeSnapshot() *Snapshot {
	snap := &Snapshot{
		Timestamp: time.Now(),
		FDs:       make(map[int]map[int]*FDInfo),
	}

	pids := GetAllPIDs()
	for _, pid := range pids {
		fds := ListFDs(pid)
		if fds == nil {
			continue
		}

		snap.FDs[pid] = make(map[int]*FDInfo)
		for _, fd := range fds {
			if fdInfo := GetFDInfo(pid, fd); fdInfo != nil {
				snap.FDs[pid][fd] = fdInfo
			}
		}
	}

	return snap
}

// DiffSnapshots compares two snapshots and returns opened/closed events.
// Used by watch.go for live mode.
func DiffSnapshots(old, new *Snapshot) (opened, closed []*FDInfo) {
	// Find newly opened fds
	for pid, newFDs := range new.FDs {
		oldFDs, hadPID := old.FDs[pid]
		for fd, newInfo := range newFDs {
			if !hadPID || oldFDs[fd] == nil {
				opened = append(opened, newInfo)
			}
		}
	}

	// Find closed fds
	for pid, oldFDs := range old.FDs {
		newFDs, hasPID := new.FDs[pid]
		for fd, oldInfo := range oldFDs {
			if !hasPID || newFDs[fd] == nil {
				closed = append(closed, oldInfo)
			}
		}
	}

	return
}
