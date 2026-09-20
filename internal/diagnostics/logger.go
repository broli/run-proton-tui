package diagnostics

import (
	"os"
	"path/filepath"
	"sort"
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

	addFile := func(path string) {
		if seen[path] {
			return
		}
		info, err := os.Stat(path)
		if err != nil || info.IsDir() {
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
	_ = filepath.Walk(logsDir, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			addFile(path)
		}
		return nil
	})

	// 2. Legacy proton-game.log or steam-*.log in root
	rootFiles, _ := filepath.Glob(filepath.Join(gameDir, "*.log"))
	for _, f := range rootFiles {
		addFile(f)
	}

	// 3. Game crash logs in prefix AppData
	if prefixDir != "" {
		pfxUsers := filepath.Join(prefixDir, "pfx", "drive_c", "users")
		_ = filepath.Walk(pfxUsers, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return nil
			}
			if !info.IsDir() && (strings.HasSuffix(path, ".log") || strings.HasSuffix(path, ".dmp")) {
				addFile(path)
			}
			return nil
		})
	}

	// Sort logs: newest first
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].ModTime.After(logs[j].ModTime)
	})

	return logs
}
