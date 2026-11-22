package prefix

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

// ProcessConflict represents an active Wine/Proton process running on another prefix.
type ProcessConflict struct {
	PID     int
	ExeName string
	Prefix  string
}

// ListOtherWineProcesses scans /proc for processes belonging to the current user
// that are running Wine/Proton executables in a DIFFERENT prefix than currentPrefix.
// This detects game/launcher collisions before launching to prevent VRAM or /dev/ntsync contention.
func ListOtherWineProcesses(currentPrefix string) ([]ProcessConflict, error) {
	canonicalCurrent := GetCanonicalPfx(currentPrefix)
	realCurrent, err := filepath.EvalSymlinks(canonicalCurrent)
	if err == nil {
		canonicalCurrent = realCurrent
	}

	uid := os.Getuid()
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return nil, err
	}

	var conflicts []ProcessConflict

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		// Verify process owner matches current user
		procDir := filepath.Join("/proc", entry.Name())
		var sysStat syscall.Stat_t
		if err := syscall.Stat(procDir, &sysStat); err != nil || int(sysStat.Uid) != uid {
			continue
		}

		// Check executable name
		exeTarget, err := os.Readlink(filepath.Join(procDir, "exe"))
		if err != nil {
			continue
		}
		lowerExe := strings.ToLower(exeTarget)
		if !strings.HasSuffix(lowerExe, ".exe") &&
			!strings.Contains(lowerExe, "wineserver") &&
			!strings.Contains(lowerExe, "wine64-preloader") &&
			!strings.Contains(lowerExe, "wine-preloader") {
			continue
		}

		// Read environ to find WINEPREFIX
		envBytes, err := os.ReadFile(filepath.Join(procDir, "environ"))
		if err != nil {
			continue
		}

		pfxMatch := extractWinePrefix(envBytes)
		if pfxMatch != "" {
			realPfx, err := filepath.EvalSymlinks(pfxMatch)
			if err == nil {
				pfxMatch = realPfx
			}

			if pfxMatch != canonicalCurrent {
				conflicts = append(conflicts, ProcessConflict{
					PID:     pid,
					ExeName: filepath.Base(exeTarget),
					Prefix:  pfxMatch,
				})
			}
		}
	}

	return conflicts, nil
}

// KillOtherWineProcesses terminates all external Wine processes detected by ListOtherWineProcesses.
func KillOtherWineProcesses(currentPrefix string) (int, error) {
	conflicts, err := ListOtherWineProcesses(currentPrefix)
	if err != nil {
		return 0, err
	}

	killedCount := 0
	for _, c := range conflicts {
		proc, err := os.FindProcess(c.PID)
		if err == nil && proc.Signal(syscall.SIGKILL) == nil {
			killedCount++
		}
	}

	return killedCount, nil
}

// KillPrefixOrphans terminates ONLY processes that belong strictly to targetPrefix.
// This protects all other running games/launchers from being touched.
func KillPrefixOrphans(targetPrefix string) (int, error) {
	canonicalTarget := GetCanonicalPfx(targetPrefix)
	realTarget, err := filepath.EvalSymlinks(canonicalTarget)
	if err == nil {
		canonicalTarget = realTarget
	}

	uid := os.Getuid()
	entries, err := os.ReadDir("/proc")
	if err != nil {
		return 0, err
	}

	killed := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		pid, err := strconv.Atoi(entry.Name())
		if err != nil {
			continue
		}

		procDir := filepath.Join("/proc", entry.Name())
		var sysStat syscall.Stat_t
		if err := syscall.Stat(procDir, &sysStat); err != nil || int(sysStat.Uid) != uid {
			continue
		}

		envBytes, err := os.ReadFile(filepath.Join(procDir, "environ"))
		if err != nil {
			continue
		}

		pfxMatch := extractWinePrefix(envBytes)
		if pfxMatch != "" {
			if realPfx, err := filepath.EvalSymlinks(pfxMatch); err == nil {
				pfxMatch = realPfx
			}
			if pfxMatch == canonicalTarget {
				if proc, err := os.FindProcess(pid); err == nil && proc.Signal(syscall.SIGKILL) == nil {
					killed++
				}
			}
		}
	}

	return killed, nil
}

var winePrefixKey = []byte("WINEPREFIX=")

func extractWinePrefix(env []byte) string {
	remaining := env
	for len(remaining) > 0 {
		idx := bytes.Index(remaining, winePrefixKey)
		if idx == -1 {
			return ""
		}
		// Verify start boundary: must be at offset 0 or preceded by null byte
		if idx == 0 || remaining[idx-1] == 0 {
			val := remaining[idx+len(winePrefixKey):]
			if end := bytes.IndexByte(val, 0); end != -1 {
				return string(val[:end])
			}
			return string(val)
		}
		remaining = remaining[idx+len(winePrefixKey):]
	}
	return ""
}
