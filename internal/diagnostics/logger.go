package diagnostics

import (
	"bufio"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

// LogFileInfo provides metadata for a discovered log file.
type LogFileInfo struct {
	Path         string
	RelativePath string
	Size         int64
	ModTime      time.Time
}

// EnsureLogsDir creates $GAME_DIR/.logs if it does not exist.
func EnsureLogsDir(gameDir string) (string, error) {
	logsDir := filepath.Join(gameDir, ".logs")
	err := os.MkdirAll(logsDir, 0755)
	return logsDir, err
}

// FindRecentLogs discovers all log files in $GAME_DIR/.logs, $GAME_DIR, and the Wine prefix.
func FindRecentLogs(gameDir string, prefixDir string) []LogFileInfo {
	var logs []LogFileInfo
	seen := make(map[string]bool)

	addEntry := func(path string, info os.FileInfo) {
		if seen[path] {
			return
		}
		if info == nil {
			var err error
			info, err = os.Stat(path)
			if err != nil || info.IsDir() {
				return
			}
		} else if info.IsDir() {
			return
		}
		seen[path] = true
		rel, _ := filepath.Rel(gameDir, path)
		logs = append(logs, LogFileInfo{
			Path:         path,
			RelativePath: rel,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
		})
	}

	// 1. Logs in .logs/
	logsDir := filepath.Join(gameDir, ".logs")
	_ = filepath.WalkDir(logsDir, func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() {
			if info, err := d.Info(); err == nil {
				addEntry(path, info)
			}
		}
		return nil
	})

	// 2. Legacy proton-game.log or steam-*.log in root
	rootFiles, _ := filepath.Glob(filepath.Join(gameDir, "*.log"))
	for _, f := range rootFiles {
		addEntry(f, nil)
	}

	// 3. Game crash logs in prefix AppData
	if prefixDir != "" {
		pfxUsers := filepath.Join(prefixDir, "pfx", "drive_c", "users")
		_ = filepath.WalkDir(pfxUsers, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if !d.IsDir() && (strings.HasSuffix(path, ".log") || strings.HasSuffix(path, ".dmp")) {
				if info, err := d.Info(); err == nil {
					addEntry(path, info)
				}
			}
			return nil
		})
	}

	// Sort logs: newest first
	slices.SortFunc(logs, func(a, b LogFileInfo) int {
		if a.ModTime.After(b.ModTime) {
			return -1
		}
		if a.ModTime.Before(b.ModTime) {
			return 1
		}
		return 0
	})

	return logs
}

// ReadLogTail reads the last n lines of a file safely using a fixed ring buffer.
func ReadLogTail(logPath string, n int) []string {
	if n <= 0 {
		return nil
	}
	f, err := os.Open(logPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	ring := make([]string, n)
	idx := 0
	count := 0

	scanner := bufio.NewScanner(f)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)
	for scanner.Scan() {
		ring[idx] = scanner.Text()
		idx = (idx + 1) % n
		count++
	}
	_ = scanner.Err()

	if count == 0 {
		return nil
	}
	if count < n {
		return ring[:count]
	}
	res := make([]string, n)
	for i := 0; i < n; i++ {
		res[i] = ring[(idx+i)%n]
	}
	return res
}
