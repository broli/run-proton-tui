package hardware

import (
	"os/exec"
	"strings"
)

// PowerManager tracks and manages the system power profile (via powerprofilesctl).
// When active, it temporarily switches to 'performance' mode during game execution,
// and guarantees restoration of the previous profile when the game terminates.
type PowerManager struct {
	OriginalProfile string
	Changed         bool
	Available       bool
}

// NewPowerManager initializes a PowerManager and checks if powerprofilesctl is available.
func NewPowerManager() *PowerManager {
	pm := &PowerManager{
		OriginalProfile: "",
		Changed:         false,
		Available:       false,
	}

	if _, err := exec.LookPath("powerprofilesctl"); err == nil {
		pm.Available = true
		pm.OriginalProfile = pm.getCurrentProfile()
	}

	return pm
}

// getCurrentProfile queries the current active power profile.
func (pm *PowerManager) getCurrentProfile() string {
	if !pm.Available {
		return ""
	}
	out, err := exec.Command("powerprofilesctl", "get").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

// SetPerformance boosts the system power profile to 'performance' if not already set.
func (pm *PowerManager) SetPerformance() error {
	if !pm.Available {
		return nil
	}

	current := pm.getCurrentProfile()
	if current == "" {
		return nil
	}

	pm.OriginalProfile = current
	if current != "performance" {
		err := exec.Command("powerprofilesctl", "set", "performance").Run()
		if err == nil {
			pm.Changed = true
		}
		return err
	}

	return nil
}

// Restore reverts the power profile back to the original state if modified by this manager.
func (pm *PowerManager) Restore() {
	if !pm.Available || !pm.Changed || pm.OriginalProfile == "" {
		return
	}

	if pm.OriginalProfile != "performance" {
		_ = exec.Command("powerprofilesctl", "set", pm.OriginalProfile).Run()
		pm.Changed = false
	}
}
