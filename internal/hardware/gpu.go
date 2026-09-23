package hardware

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GPUInfo encapsulates detected GPU devices, hybrid offloading availability, and display topology.
type GPUInfo struct {
	HasPrimeRun      bool     // True if /usr/bin/prime-run or prime-run is in PATH
	HasNvidia        bool     // True if NVIDIA PCI vendor 0x10de is detected
	HasIntel         bool     // True if Intel PCI vendor 0x8086 is detected
	HasAMD           bool     // True if AMD PCI vendor 0x1002 is detected
	ConnectedOutputs []string // List of connected display connectors (e.g. HDMI-A-1, eDP-1)
	PreferredOutput  string   // Selected target for Gamescope (prefers external HDMI/DP over internal eDP)
}

// DetectGPU scans DRM sysfs nodes and PATH to establish graphics capabilities.
// This is critical on hybrid laptops to avoid KWin Wayland compositor crashes
// caused by running Gamescope on the NVIDIA dGPU directly.
func DetectGPU() (*GPUInfo, error) {
	info := &GPUInfo{
		HasPrimeRun:      false,
		HasNvidia:        false,
		HasIntel:         false,
		HasAMD:           false,
		ConnectedOutputs: make([]string, 0),
		PreferredOutput:  "",
	}

	// 1. Check for prime-run in PATH
	if _, err := exec.LookPath("prime-run"); err == nil {
		info.HasPrimeRun = true
	}

	// 2. Scan /sys/class/drm/card*/device/vendor for GPU vendors
	cards, _ := filepath.Glob("/sys/class/drm/card[0-9]")
	for _, card := range cards {
		vendorPath := filepath.Join(card, "device", "vendor")
		vendorBytes, err := os.ReadFile(vendorPath)
		if err != nil {
			continue
		}
		vendor := strings.ToLower(strings.TrimSpace(string(vendorBytes)))
		switch vendor {
		case "0x10de":
			info.HasNvidia = true
		case "0x8086":
			info.HasIntel = true
		case "0x1002":
			info.HasAMD = true
		}
	}

	// 3. Scan /sys/class/drm/card*-*/status for active display connectors
	connectorStatuses, _ := filepath.Glob("/sys/class/drm/card*-*/status")
	for _, statusFile := range connectorStatuses {
		content, err := os.ReadFile(statusFile)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(content)) == "connected" {
			// Extract connector name (e.g. "card1-HDMI-A-1" -> "HDMI-A-1")
			dir := filepath.Base(filepath.Dir(statusFile))
			parts := strings.SplitN(dir, "-", 2)
			if len(parts) == 2 {
				connectorName := parts[1]
				info.ConnectedOutputs = append(info.ConnectedOutputs, connectorName)
			}
		}
	}

	// 4. Select preferred display output:
	// Prioritize external displays (HDMI-A-1, DP-*) over internal laptop displays (eDP-*)
	for _, conn := range info.ConnectedOutputs {
		if strings.HasPrefix(conn, "HDMI") {
			info.PreferredOutput = conn
			break
		}
	}
	if info.PreferredOutput == "" {
		for _, conn := range info.ConnectedOutputs {
			if strings.HasPrefix(conn, "DP") {
				info.PreferredOutput = conn
				break
			}
		}
	}
	if info.PreferredOutput == "" && len(info.ConnectedOutputs) > 0 {
		info.PreferredOutput = info.ConnectedOutputs[0]
	}

	return info, nil
}

// GetConnectedDisplayOutputs scans /sys/class/drm/card*-*/status on demand
// and returns all currently connected display connector names (e.g. ["HDMI-A-1", "eDP-1"]).
// This queries world-readable sysfs files (0444) without requiring root, sudo, polkit, or external tools.
func GetConnectedDisplayOutputs() []string {
	var outputs []string
	seen := make(map[string]bool)

	connectorStatuses, _ := filepath.Glob("/sys/class/drm/card*-*/status")
	for _, statusFile := range connectorStatuses {
		content, err := os.ReadFile(statusFile)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(content)) == "connected" {
			dir := filepath.Base(filepath.Dir(statusFile))
			parts := strings.SplitN(dir, "-", 2)
			if len(parts) == 2 {
				name := parts[1]
				if !seen[name] {
					seen[name] = true
					outputs = append(outputs, name)
				}
			}
		}
	}
	return outputs
}
