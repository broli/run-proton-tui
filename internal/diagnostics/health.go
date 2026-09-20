package diagnostics

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

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
// It automatically attempts non-destructive fixes (such as adding missing +x bits).
func RunPreflightCheck(gameDir string, targetExe string, prefixDir string) *HealthReport {
	report := &HealthReport{
		TargetExeExists:    false,
		TargetExeFixed:     false,
		ShippingExesFixed:  make([]string, 0),
		UnownedFiles:       make([]string, 0),
		GameDirWritable:    false,
		PrefixDirWritable:  false,
		InsidePrefixHazard: false,
		Issues:             make([]string, 0),
		AllHealthy:         true,
	}

	uid := os.Getuid()

	// 1. Guardrail Check: Is target exe or gameDir inside prefix?
	fullExePath := filepath.Join(gameDir, targetExe)
	if prefix.CheckPrefixContainment(prefixDir, fullExePath) || prefix.CheckPrefixContainment(prefixDir, gameDir) {
		report.InsidePrefixHazard = true
		report.AllHealthy = false
		report.Issues = append(report.Issues, "HAZARD: Game executable is located inside the Wine prefix! Cleaning prefix will erase game files.")
	}

	// 2. Check Target Executable (+x permissions)
	if info, err := os.Stat(fullExePath); err == nil {
		report.TargetExeExists = true
		if info.Mode()&0111 == 0 {
			// Lacks executable bit -> attempt chmod +x
			if err := os.Chmod(fullExePath, info.Mode()|0111); err == nil {
				report.TargetExeFixed = true
			} else {
				report.AllHealthy = false
				report.Issues = append(report.Issues, fmt.Sprintf("Executable lacks +x permission: %s", targetExe))
			}
		}
	} else if targetExe != "" {
		report.AllHealthy = false
		report.Issues = append(report.Issues, fmt.Sprintf("Target executable not found: %s", targetExe))
	}

	// 3. Scan and auto-fix Unreal Engine *Shipping.exe binaries
	_ = filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(path, "proton-prefix") || strings.Contains(path, ".logs") {
			return nil
		}
		if strings.HasSuffix(strings.ToLower(info.Name()), "shipping.exe") {
			if info.Mode()&0111 == 0 {
				if err := os.Chmod(path, info.Mode()|0111); err == nil {
					rel, _ := filepath.Rel(gameDir, path)
					report.ShippingExesFixed = append(report.ShippingExesFixed, rel)
				}
			}
		}
		return nil
	})

	// 4. File Ownership Check
	_ = filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if strings.Contains(path, "proton-prefix") || strings.Contains(path, ".logs") || strings.Contains(path, ".git") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		var sysStat syscall.Stat_t
		if err := syscall.Stat(path, &sysStat); err == nil {
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

	// 5. Game Directory Writeability
	testFile := filepath.Join(gameDir, ".rpt_write_test")
	if err := os.WriteFile(testFile, []byte("ok"), 0644); err == nil {
		report.GameDirWritable = true
		_ = os.Remove(testFile)
	} else {
		report.AllHealthy = false
		report.Issues = append(report.Issues, "Game directory is not writable!")
	}

	// 6. Prefix Directory Writeability
	checkDir := prefixDir
	if _, err := os.Stat(checkDir); os.IsNotExist(err) {
		checkDir = filepath.Dir(prefixDir)
	}
	pfxTest := filepath.Join(checkDir, ".rpt_write_test")
	if err := os.WriteFile(pfxTest, []byte("ok"), 0644); err == nil {
		report.PrefixDirWritable = true
		_ = os.Remove(pfxTest)
	} else {
		report.AllHealthy = false
		report.Issues = append(report.Issues, "Wine prefix path is not writable!")
	}

	return report
}
