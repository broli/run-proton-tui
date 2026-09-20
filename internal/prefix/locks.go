package prefix

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// CleanStaleLocks iterates through all Wine server directories in /tmp/.wine-<UID>/
// and tests their 'lock' files using non-blocking flock.
// If flock succeeds, the server process is dead and the lock was abandoned.
// Stale directories are safely purged to prevent lock contention on subsequent launches.
func CleanStaleLocks() ([]string, error) {
	uid := os.Getuid()
	baseDir := fmt.Sprintf("/tmp/.wine-%d", uid)

	pattern := filepath.Join(baseDir, "server-*", "lock")
	lockFiles, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	cleaned := make([]string, 0)

	for _, lockPath := range lockFiles {
		fd, err := syscall.Open(lockPath, syscall.O_RDWR, 0600)
		if err != nil {
			// If file cannot be opened (e.g. permission or already gone), continue
			continue
		}

		// Attempt non-blocking exclusive lock
		err = syscall.Flock(fd, syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			// Lock acquired! This wineserver is dead.
			sockDir := filepath.Dir(lockPath)
			_ = syscall.Close(fd)
			_ = os.RemoveAll(sockDir)
			cleaned = append(cleaned, sockDir)
		} else {
			// Lock held by an active wineserver. Do not touch.
			_ = syscall.Close(fd)
		}
	}

	return cleaned, nil
}
