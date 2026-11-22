package diagnostics

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/broli/run-proton-tui/internal/pathutil"
	"github.com/broli/run-proton-tui/internal/prefix"
)

// HealthReport summarizes system permissions, directory health, and file ownership.
type HealthReport struct {
	TargetExeExists    bool
	TargetExeFixed     bool
	ShippingExesFixed  []string
	UnownedFiles       []string
	GameDirWritable    bool
	PrefixDirWritable  bool
	InsidePrefixHazard bool
	Issues             []string
	AllHealthy         bool
}

// RunPreflightCheck inspects file permissions, ownership, and directory safety.
func RunPreflightCheck(gameDir string, targetExe string, prefixDir string) *HealthReport {
	report := &HealthReport{
		AllHealthy: true,
	}

	uid := os.Getuid()

	// Guardrail Check: Is target exe or gameDir inside prefix?
	fullExePath := filepath.Join(gameDir, targetExe)
	if prefix.CheckPrefixContainment(prefixDir, fullExePath) || prefix.CheckPrefixContainment(prefixDir, gameDir) {
		report.InsidePrefixHazard = true
		report.AllHealthy = false
		report.Issues = append(report.Issues, "HAZARD: Game executable is located inside the Wine prefix! Cleaning prefix will erase game files.")
	}

	// Target executable existence
	if _, err := os.Stat(fullExePath); err == nil {
		report.TargetExeExists = true
	} else if targetExe != "" {
		report.AllHealthy = false
		report.Issues = append(report.Issues, fmt.Sprintf("Target executable not found: %s", targetExe))
	}

	// File ownership check across game files
	_ = filepath.WalkDir(gameDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if pathutil.IsIgnoredDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		if sysStat, ok := info.Sys().(*syscall.Stat_t); ok {
			if int(sysStat.Uid) != uid && len(report.UnownedFiles) < 5 {
				rel, _ := filepath.Rel(gameDir, path)
				report.UnownedFiles = append(report.UnownedFiles, rel)
			}
		}
		return nil
	})

	if len(report.UnownedFiles) > 0 {
		report.AllHealthy = false
		report.Issues = append(report.Issues, fmt.Sprintf("Found files not owned by current user (e.g. %s)", report.UnownedFiles[0]))
	}

	// Game directory writeability
	if isDirWritable(gameDir) {
		report.GameDirWritable = true
	} else {
		report.AllHealthy = false
		report.Issues = append(report.Issues, "Game directory is not writable!")
	}

	// Prefix directory writeability
	checkDir := prefixDir
	if _, err := os.Stat(checkDir); os.IsNotExist(err) {
		checkDir = filepath.Dir(prefixDir)
	}
	if isDirWritable(checkDir) {
		report.PrefixDirWritable = true
	} else {
		report.AllHealthy = false
		report.Issues = append(report.Issues, "Wine prefix path is not writable!")
	}

	return report
}

func isDirWritable(dir string) bool {
	return syscall.Access(dir, 2) == nil
}
