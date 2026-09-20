package prefix

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// BackupSaves searches the Wine prefix for standard Windows save game locations
// (AppData/Local, AppData/Roaming, Saved Games, Documents) and archives them
// to ~/Games/Backups/<gameName>/<timestamp>/ before any prefix recreation.
func BackupSaves(prefixDir string, gameName string) (string, error) {
	pfx := GetCanonicalPfx(prefixDir)
	usersDir := filepath.Join(pfx, "drive_c", "users")
	if _, err := os.Stat(usersDir); err != nil {
		return "", nil // No prefix users directory, nothing to back up
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	safeGameName := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			return r
		}
		return '_'
	}, gameName)

	destDir := filepath.Join(home, "Games", "Backups", safeGameName, timestamp)

	var backedUpFiles int
	savePaths := []string{
		"Saved Games",
		"Documents",
		"AppData/Local",
		"AppData/Roaming",
	}

	userEntries, err := os.ReadDir(usersDir)
	if err != nil {
		return "", err
	}

	for _, user := range userEntries {
		if !user.IsDir() || user.Name() == "Public" {
			continue
		}

		userRoot := filepath.Join(usersDir, user.Name())
		for _, sp := range savePaths {
			source := filepath.Join(userRoot, sp)
			if info, err := os.Stat(source); err == nil && info.IsDir() {
				target := filepath.Join(destDir, user.Name(), sp)
				count, _ := copyDir(source, target)
				backedUpFiles += count
			}
		}
	}

	if backedUpFiles > 0 {
		return destDir, nil
	}

	// If no files were copied, remove the empty directory
	_ = os.RemoveAll(destDir)
	return "", nil
}

func copyDir(src, dst string) (int, error) {
	count := 0
	err := filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		rel, err := filepath.Rel(src, path)
		if err != nil {
			return nil
		}
		targetPath := filepath.Join(dst, rel)

		if info.IsDir() {
			return os.MkdirAll(targetPath, 0755)
		}

		// Don't copy logs, temp files or large crash dumps
		lower := strings.ToLower(info.Name())
		if strings.HasSuffix(lower, ".log") || strings.HasSuffix(lower, ".tmp") || strings.HasSuffix(lower, ".dmp") {
			return nil
		}

		// Only copy files under 100MB to avoid backing up game caches
		if info.Size() > 100*1024*1024 {
			return nil
		}

		if err := copyFile(path, targetPath); err == nil {
			count++
		}
		return nil
	})

	return count, err
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
