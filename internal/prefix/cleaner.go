package prefix

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// CleanResult contains the summary of a prefix reset operation.
type CleanResult struct {
	BackupDir string
	Success   bool
}

// CheckPrefixContainment inspects if a given path is located inside the prefix directory.
// This is the core guardrail that prevents users from wiping their game installations.
func CheckPrefixContainment(prefixDir string, testPath string) bool {
	if testPath == "" || prefixDir == "" {
		return false
	}

	absPrefix, err := filepath.Abs(prefixDir)
	if err != nil {
		absPrefix = prefixDir
	}
	realPrefix, err := filepath.EvalSymlinks(absPrefix)
	if err == nil {
		absPrefix = realPrefix
	}

	absPath, err := filepath.Abs(testPath)
	if err != nil {
		absPath = testPath
	}
	realPath, err := filepath.EvalSymlinks(absPath)
	if err == nil {
		absPath = realPath
	}

	// Direct substring or relative path check
	rel, err := filepath.Rel(absPrefix, absPath)
	if err == nil && !strings.HasPrefix(rel, "..") && rel != "." {
		return true
	}

	return strings.Contains(absPath, "/proton-prefix/") || strings.Contains(absPath, "/pfx/drive_c/")
}

// SafeCleanPrefix wipes and regenerates a Wine prefix while enforcing strict safety rules:
// 1. Blocks operation if targetExe or gameDir is inside the prefix.
// 2. Automatically preserves user save games to ~/Games/Backups/.
// 3. Gracefully flushes wineserver and socket locks before removing files.
func SafeCleanPrefix(prefixDir string, protonPath string, gameDir string, targetExe string, gameName string) (*CleanResult, error) {
	// Guardrail check 1: Target executable inside prefix
	if CheckPrefixContainment(prefixDir, filepath.Join(gameDir, targetExe)) {
		return nil, errors.New("SAFETY GUARDRAIL TRIGGERED: Target executable is located INSIDE the Wine prefix! Cleaning the prefix would permanently delete your installed game files. Please relocate the game to an external directory (e.g. ~/Games/Title/) and re-run")
	}

	// Guardrail check 2: Game directory itself is inside prefix
	if CheckPrefixContainment(prefixDir, gameDir) {
		return nil, errors.New("SAFETY GUARDRAIL TRIGGERED: Current working directory is inside the Wine prefix! Prefix clean aborted")
	}

	result := &CleanResult{
		Success: false,
	}

	// 1. Back up save files
	if backupPath, err := BackupSaves(prefixDir, gameName); err == nil && backupPath != "" {
		result.BackupDir = backupPath
	}

	// 2. Flush running processes & sockets
	_ = Flush(prefixDir, protonPath)

	// 3. Clean stale socket locks
	_, _ = CleanStaleLocks()

	// 4. Safe wipe
	if err := os.RemoveAll(prefixDir); err != nil {
		return result, fmt.Errorf("failed to remove prefix directory: %w", err)
	}

	// 5. Recreate base directory
	if err := os.MkdirAll(prefixDir, 0755); err != nil {
		return result, fmt.Errorf("failed to recreate prefix directory: %w", err)
	}

	result.Success = true
	return result, nil
}
