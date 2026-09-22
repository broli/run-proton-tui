package runner

import (
	"path/filepath"
	"strings"
)

// ExeType defines the classification of a Windows binary.
type ExeType string

const (
	ExeType2DUtility ExeType = "2D Utility / Launcher / Installer"
	ExeType3DGame    ExeType = "3D Game Engine"
)

// ClassificationResult provides the detected category and recommended defaults with clear rationale.
type ClassificationResult struct {
	Type               ExeType
	Rationale          string
	RecommendGamescope bool
	RecommendPrimeRun  bool
	RecommendPCores    bool
}

// ClassifyExecutable analyzes the filename and path of an executable to suggest optimal execution settings.
// Crucially, it preserves user agency: these are recommendations that the user can freely toggle.
func ClassifyExecutable(exePath string) ClassificationResult {
	base := strings.ToLower(filepath.Base(exePath))

	// 2D utilities, launchers, installers, updaters, and crash reporters
	is2D := strings.Contains(base, "launcher") ||
		strings.Contains(base, "update") ||
		strings.Contains(base, "setup") ||
		strings.Contains(base, "patch") ||
		strings.Contains(base, "installer") ||
		strings.Contains(base, "unins") ||
		strings.Contains(base, "repair") ||
		strings.Contains(base, "redist") ||
		strings.Contains(base, "dxsetup") ||
		strings.Contains(base, "vcredist") ||
		strings.Contains(base, "crashreport") ||
		strings.Contains(base, "crashhandler") ||
		strings.Contains(base, "config") ||
		strings.Contains(base, "service") ||
		strings.Contains(base, "helper") ||
		strings.Contains(base, "anticheat") ||
		strings.Contains(base, "ace-") ||
		strings.Contains(base, "cef") ||
		strings.Contains(base, "webengine") ||
		strings.Contains(base, "platformprocess") ||
		base == "games.exe"

	if is2D {
		return ClassificationResult{
			Type:               ExeType2DUtility,
			Rationale:          "Detected 2D launcher / setup utility (often Qt5/CEF WebEngine). Recommended: run on host iGPU without Gamescope to prevent NVAPI double-free crashes.",
			RecommendGamescope: false,
			RecommendPrimeRun:  false,
			RecommendPCores:    false,
		}
	}

	// 3D Game binaries
	return ClassificationResult{
		Type:               ExeType3DGame,
		Rationale:          "Detected 3D game executable. Recommended: NVIDIA dGPU offload + Gamescope sandboxing (1080p) + P-Core pinning for peak FPS and stutter prevention.",
		RecommendGamescope: true,
		RecommendPrimeRun:  true,
		RecommendPCores:    true,
	}
}
