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
	absPrefix = filepath.Clean(absPrefix)
	if realPrefix, err := filepath.EvalSymlinks(absPrefix); err == nil {
		absPrefix = filepath.Clean(realPrefix)
	}

	absPath, err := filepath.Abs(testPath)
	if err != nil {
		absPath = testPath
	}
	absPath = filepath.Clean(absPath)
	if realPath, err := filepath.EvalSymlinks(absPath); err == nil {
		absPath = filepath.Clean(realPath)
	}

	// Exact match: testPath IS the prefix directory
	if absPath == absPrefix {
		return true
	}

	// Relative path check: testPath is a descendant of prefixDir
	rel, err := filepath.Rel(absPrefix, absPath)
	if err == nil && !strings.HasPrefix(rel, "..") && rel != ".." {
		return true
	}

	// Direct prefix boundary check
	prefixWithSep := absPrefix + string(filepath.Separator)
	if strings.HasPrefix(absPath, prefixWithSep) {
		return true
	}

	// Secondary check: if testPath explicitly points into canonical pfx/drive_c of prefixDir
	canonicalPfx := filepath.Join(absPrefix, "pfx", "drive_c")
	if absPath == canonicalPfx || strings.HasPrefix(absPath, canonicalPfx+string(filepath.Separator)) {
		return true
	}

	return false
}

// SafeCleanPrefix wipes and regenerates a Wine prefix while enforcing strict safety rules:
// 1. Blocks operation if targetExe or gameDir is inside the prefix.
// 2. Automatically preserves user save games and screenshots to destDir or ~/Games/Backups/.
// 3. Gracefully flushes wineserver and socket locks before removing files.
func SafeCleanPrefix(prefixDir string, protonPath string, gameDir string, targetExe string, gameName string) (*CleanResult, error) {
	return SafeCleanPrefixCustom(prefixDir, protonPath, gameDir, targetExe, gameName, "", nil)
}

// SafeCleanPrefixCustom performs SafeCleanPrefix with customizable destination and extra paths to preserve.
func SafeCleanPrefixCustom(prefixDir string, protonPath string, gameDir string, targetExe string, gameName string, destDir string, extraPaths []string) (*CleanResult, error) {
	// Guardrail check 1: Target executable inside prefix
	if targetExe != "" && CheckPrefixContainment(prefixDir, filepath.Join(gameDir, targetExe)) {
		return nil, errors.New("SAFETY GUARDRAIL TRIGGERED: Target executable is located INSIDE the Wine prefix! Cleaning the prefix would permanently delete your installed game files. Please relocate the game to an external directory (e.g. ~/Games/Title/) and re-run")
	}

	// Guardrail check 2: Game directory itself is inside prefix
	if CheckPrefixContainment(prefixDir, gameDir) {
		return nil, errors.New("SAFETY GUARDRAIL TRIGGERED: Current working directory is inside the Wine prefix! Prefix clean aborted")
	}

	result := &CleanResult{
		Success: false,
	}

	// 1. Preserve save files and screenshots
	preserveOpts := PreserveOptions{
		PrefixDir:  prefixDir,
		GameName:   gameName,
		DestDir:    destDir,
		ExtraPaths: extraPaths,
	}
	if backupPath, err := PreserveSaves(preserveOpts); err == nil && backupPath != "" {
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

	result.Success = true
	return result, nil
}
