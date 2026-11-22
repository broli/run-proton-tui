package prefix

import (
	"os"
	"os/exec"
	"path/filepath"
)

// FindProtonTool locates a tool binary (such as "wineserver" or "wine") in the Proton distribution
// or falls back to system PATH.
func FindProtonTool(protonPath, toolName string) string {
	if protonPath != "" {
		pDir := filepath.Dir(protonPath)
		cand1 := filepath.Join(pDir, "files/bin", toolName)
		if _, err := os.Stat(cand1); err == nil {
			return cand1
		}
		cand2 := filepath.Join(pDir, "dist/bin", toolName)
		if _, err := os.Stat(cand2); err == nil {
			return cand2
		}
	}

	if sysBin, err := exec.LookPath(toolName); err == nil {
		return sysBin
	}

	return ""
}

// GetWineserverBin resolves the path to the wineserver binary associated with the given Proton runner.
func GetWineserverBin(protonPath string) string {
	return FindProtonTool(protonPath, "wineserver")
}

// Flush performs a complete, safe flush of the Wine prefix:
// 1. Sends graceful termination signal to wineserver (-k) and waits (-w).
// 2. Kills any lingering orphan processes attached strictly to this prefix.
// 3. Deletes the orphaned IPC socket directory in /tmp/.wine-<UID>/.
func Flush(prefixDir string, protonPath string) error {
	pfx := GetCanonicalPfx(prefixDir)
	wsBin := GetWineserverBin(protonPath)

	// 1. Graceful wineserver shutdown
	if wsBin != "" {
		if _, err := os.Stat(pfx); err == nil {
			cmdK := exec.Command(wsBin, "-k")
			cmdK.Env = append(os.Environ(), "WINEPREFIX="+pfx)
			_ = cmdK.Run()

			cmdW := exec.Command(wsBin, "-w")
			cmdW.Env = append(os.Environ(), "WINEPREFIX="+pfx)
			_ = cmdW.Run()
		}
	}

	// 2. Kill prefix orphans
	_, _ = KillPrefixOrphans(pfx)

	// 3. Remove socket directory in /tmp
	if sockDir, err := GetPrefixSocketDir(pfx); err == nil {
		_ = os.RemoveAll(sockDir)
	}

	return nil
}
