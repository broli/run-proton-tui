package prefix

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// GetCanonicalPfx resolves the canonical /pfx subfolder used by Valve Proton.
// In Proton, WINEPREFIX points to $COMPAT_DATA/pfx.
func GetCanonicalPfx(prefixDir string) string {
	pfxSub := filepath.Join(prefixDir, "pfx")
	if info, err := os.Stat(pfxSub); err == nil && info.IsDir() {
		return pfxSub
	}
	return prefixDir
}

// GetPrefixSocketName calculates Wine's internal IPC socket directory name.
// Wine names server socket directories in /tmp/.wine-<UID>/ using the device and inode numbers
// of the prefix directory: 'server-<st_dev:x>-<st_ino:x>'.
func GetPrefixSocketName(pfxPath string) (string, error) {
	var stat syscall.Stat_t
	err := syscall.Stat(pfxPath, &stat)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("server-%x-%x", stat.Dev, stat.Ino), nil
}

// GetPrefixSocketDir returns the full path to the Wine socket directory for a given prefix.
func GetPrefixSocketDir(pfxPath string) (string, error) {
	sockName, err := GetPrefixSocketName(pfxPath)
	if err != nil {
		return "", err
	}
	uid := os.Getuid()
	return filepath.Join(fmt.Sprintf("/tmp/.wine-%d", uid), sockName), nil
}
